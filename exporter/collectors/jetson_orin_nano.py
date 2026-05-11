"""
Jetson Orin Nano collector for NVIDIA Jetson Orin Nano devices.
Implements Orin Nano-specific metric parsing from tegrastats.

Example tegrastats output:
RAM 3423/7620MB (lfb 78x4MB) SWAP 0/3810MB (cached 0MB)
CPU [2%@729,6%@729,2%@729,1%@729,2%@729,2%@729] GR3D_FREQ 0%
cpu@45.531C soc2@44.781C soc0@45.937C gpu@46.468C tj@46.468C soc1@46C
VDD_IN 3311mW/3311mW VDD_CPU_GPU_CV 524mW/524mW VDD_SOC 1008mW/1008mW
"""
import re
from typing import Dict
from .jetson import JetsonCollector


class JetsonOrinNanoCollector(JetsonCollector):
    """
    Collector for NVIDIA Jetson Orin Nano devices.

    Orin Nano-specific characteristics:
    - 6 CPU cores
    - Power rails: VDD_IN, VDD_CPU_GPU_CV, VDD_SOC
    - Temperature sensors commonly reported in lowercase: cpu, gpu, tj, soc0-2
    - GPU usage may be reported without an accompanying frequency: GR3D_FREQ 0%
    - SWAP includes cached info: SWAP 0/3810MB (cached 0MB)
    """

    def _parse_all_metrics(self, output: str) -> Dict[str, float]:
        """
        Parse tegrastats output for Jetson Orin Nano devices.

        Args:
            output: Raw tegrastats output line

        Returns:
            Dictionary with metric_name -> value (normalized to standard units)
        """
        metrics = {}

        # 1. Power rails: VDD_IN 3311mW/3311mW or VDD_IN 3311mW
        power_pattern = r'(\w+)\s+(\d+)mW(?:/(\d+)mW)?'
        for match in re.finditer(power_pattern, output):
            rail_name = match.group(1)
            current_mw = float(match.group(2))
            avg_mw = float(match.group(3)) if match.group(3) else current_mw

            if rail_name == "NC":
                continue

            metrics[f"jetson_power_{rail_name.lower()}_watts"] = round(current_mw / 1000.0, 3)
            if match.group(3):
                metrics[f"jetson_power_{rail_name.lower()}_avg_watts"] = round(avg_mw / 1000.0, 3)

        # 2. Temperatures: cpu@45.531C, gpu@46.468C, tj@46.468C, etc.
        temp_pattern = r'(\w+)@([-\d.]+)C'
        for match in re.finditer(temp_pattern, output):
            sensor_name = match.group(1)
            temp_c = float(match.group(2))

            if temp_c < -100:
                continue

            metrics[f"jetson_temp_{sensor_name.lower()}_celsius"] = round(temp_c, 2)

        # 3. RAM: RAM 3423/7620MB
        ram_match = re.search(r'RAM\s+(\d+)/(\d+)MB', output)
        if ram_match:
            used_mb = float(ram_match.group(1))
            total_mb = float(ram_match.group(2))
            metrics["jetson_ram_used_mb"] = used_mb
            metrics["jetson_ram_total_mb"] = total_mb
            metrics["jetson_ram_used_percent"] = round((used_mb / total_mb) * 100, 2)

        # 4. SWAP: SWAP 0/3810MB (cached 0MB)
        swap_match = re.search(r'SWAP\s+(\d+)/(\d+)MB(?:\s+\(cached\s+(\d+)MB\))?', output)
        if swap_match:
            used_mb = float(swap_match.group(1))
            total_mb = float(swap_match.group(2))
            metrics["jetson_swap_used_mb"] = used_mb
            metrics["jetson_swap_total_mb"] = total_mb

            if swap_match.group(3):
                metrics["jetson_swap_cached_mb"] = float(swap_match.group(3))

        # 5. LFB (Largest Free Block): lfb 78x4MB
        lfb_match = re.search(r'lfb\s+(\d+)x(\d+)MB', output)
        if lfb_match:
            blocks = int(lfb_match.group(1))
            block_size_mb = int(lfb_match.group(2))
            metrics["jetson_lfb_blocks"] = blocks
            metrics["jetson_lfb_total_mb"] = blocks * block_size_mb

        # 6. CPU usage: CPU [2%@729,6%@729,2%@729,1%@729,2%@729,2%@729]
        cpu_match = re.search(r'CPU\s+\[([^\]]+)\]', output)
        if cpu_match:
            cpu_data = cpu_match.group(1)
            cpu_cores = cpu_data.split(',')

            total_usage = 0
            active_cores = 0

            for i, core in enumerate(cpu_cores):
                core = core.strip()
                if core == "off":
                    metrics[f"jetson_cpu_core{i}_status"] = 0
                else:
                    core_match = re.match(r'(\d+)%@(\d+)', core)
                    if core_match:
                        usage = int(core_match.group(1))
                        freq_mhz = int(core_match.group(2))

                        metrics[f"jetson_cpu_core{i}_usage_percent"] = usage
                        metrics[f"jetson_cpu_core{i}_freq_mhz"] = freq_mhz
                        metrics[f"jetson_cpu_core{i}_status"] = 1

                        total_usage += usage
                        active_cores += 1

            if active_cores > 0:
                metrics["jetson_cpu_avg_usage_percent"] = round(total_usage / active_cores, 2)
                metrics["jetson_cpu_active_cores"] = active_cores

        # 7. EMC (memory controller) frequency, when present: EMC_FREQ 0%@2133
        emc_match = re.search(r'EMC_FREQ\s+(\d+)%(?:@(\d+))?', output)
        if emc_match:
            metrics["jetson_emc_usage_percent"] = int(emc_match.group(1))
            if emc_match.group(2):
                metrics["jetson_emc_freq_mhz"] = int(emc_match.group(2))

        # 8. GPU frequency. Orin Nano on JetPack 6 may report usage only.
        #    Supported forms: GR3D_FREQ 0%, GR3D_FREQ 0%@918, GR3D_FREQ 0%@[510]
        gpu_match = re.search(r'GR3D_FREQ\s+(\d+)%(?:@(?:\[([^\]]+)\]|(\d+)))?', output)
        if gpu_match:
            metrics["jetson_gpu_usage_percent"] = int(gpu_match.group(1))

            bracket_freqs = gpu_match.group(2)
            single_freq = gpu_match.group(3)
            if bracket_freqs:
                for i, freq in enumerate(bracket_freqs.split(',')):
                    metrics[f"jetson_gpu_freq{i}_mhz"] = int(freq.strip())
            elif single_freq:
                metrics["jetson_gpu_freq0_mhz"] = int(single_freq)

        # 9. VIC (video image compositor) frequency, when present: VIC_FREQ 729
        vic_match = re.search(r'VIC_FREQ\s+(\d+)', output)
        if vic_match:
            metrics["jetson_vic_freq_mhz"] = int(vic_match.group(1))

        # 10. APE (audio processing engine) frequency, when present: APE 174
        ape_match = re.search(r'APE\s+(\d+)', output)
        if ape_match:
            metrics["jetson_ape_freq_mhz"] = int(ape_match.group(1))

        self.logger.debug(f"Parsed {len(metrics)} Orin Nano metrics from tegrastats")
        return metrics
