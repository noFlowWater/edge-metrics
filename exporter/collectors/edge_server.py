"""
Edge server collector for x86 servers with Intel RAPL.
Reads power, temperature, CPU, and memory metrics from sysfs/procfs.
"""
import os
import time
from pathlib import Path
from typing import Dict, List, Optional, Tuple

from .base import BaseCollector


class EdgeServerCollector(BaseCollector):
    """
    Collector for x86 edge servers (e.g. genesis1/2/3).
    Uses sysfs (RAPL, hwmon, cpufreq) and procfs for all metrics.
    GPU metrics are handled separately by dcgm-exporter.
    """

    def __init__(self, config: dict):
        super().__init__(config)

        # CPU count
        self._cpu_count = os.cpu_count() or 0

        # Discover RAPL domains
        self._rapl_domains = self._discover_rapl_domains()

        # Discover hwmon coretemp paths
        self._hwmon_path = self._discover_coretemp_hwmon()
        self._temp_inputs = self._discover_temp_inputs()

        # Delta state for RAPL energy counters
        self._prev_rapl: Dict[str, Tuple[int, float]] = {}  # domain -> (energy_uj, timestamp)

        # Delta state for CPU usage from /proc/stat
        self._prev_cpu_times: Dict[str, Tuple[int, int]] = {}  # cpu_id -> (idle, total)

        # Cache metric names after discovery
        self._metric_name_cache: Optional[List[str]] = None

    def _discover_rapl_domains(self) -> Dict[str, Dict]:
        """Discover available RAPL domains under intel-rapl."""
        domains = {}
        rapl_base = Path("/sys/devices/virtual/powercap/intel-rapl")
        if not rapl_base.exists():
            self.logger.warning("RAPL not available at %s", rapl_base)
            return domains

        for entry in sorted(rapl_base.iterdir()):
            if not entry.name.startswith("intel-rapl:"):
                continue
            name_file = entry / "name"
            energy_file = entry / "energy_uj"
            max_range_file = entry / "max_energy_range_uj"
            if not energy_file.exists():
                continue

            domain_name = self._read_file(name_file, "unknown").strip()
            max_range = int(self._read_file(max_range_file, "0").strip())

            domains[entry.name] = {
                "path": entry,
                "name": domain_name,
                "energy_file": energy_file,
                "max_range": max_range,
            }

            # Check sub-domains (e.g. intel-rapl:0:0 for core, dram)
            for sub in sorted(entry.iterdir()):
                if not sub.name.startswith("intel-rapl:") or not sub.is_dir():
                    continue
                sub_name_file = sub / "name"
                sub_energy_file = sub / "energy_uj"
                sub_max_range_file = sub / "max_energy_range_uj"
                if not sub_energy_file.exists():
                    continue

                sub_name = self._read_file(sub_name_file, "unknown").strip()
                sub_max_range = int(self._read_file(sub_max_range_file, "0").strip())

                domains[sub.name] = {
                    "path": sub,
                    "name": sub_name,
                    "energy_file": sub_energy_file,
                    "max_range": sub_max_range,
                }

        self.logger.info("Discovered RAPL domains: %s",
                         {k: v["name"] for k, v in domains.items()})
        return domains

    def _discover_coretemp_hwmon(self) -> Optional[Path]:
        """Find the hwmon directory for coretemp."""
        hwmon_base = Path("/sys/class/hwmon")
        if not hwmon_base.exists():
            return None

        for entry in sorted(hwmon_base.iterdir()):
            name_file = entry / "name"
            if name_file.exists():
                name = self._read_file(name_file, "").strip()
                if name == "coretemp":
                    self.logger.info("Found coretemp hwmon at %s", entry)
                    return entry

        self.logger.warning("coretemp hwmon not found")
        return None

    def _discover_temp_inputs(self) -> Dict[str, Path]:
        """Discover temperature input files from coretemp hwmon."""
        inputs = {}
        if not self._hwmon_path:
            return inputs

        for f in sorted(self._hwmon_path.iterdir()):
            if f.name.startswith("temp") and f.name.endswith("_input"):
                # e.g. temp1_input -> index 1
                idx = f.name.replace("temp", "").replace("_input", "")
                label_file = self._hwmon_path / f"temp{idx}_label"
                label = self._read_file(label_file, f"temp{idx}").strip()
                inputs[label] = f

        self.logger.info("Discovered temp inputs: %s", list(inputs.keys()))
        return inputs

    @classmethod
    def metric_names(cls) -> List[str]:
        return ["server_metric_discovered_on_first_run"]

    def get_metrics(self) -> Dict[str, float]:
        metrics = {}
        now = time.monotonic()

        self._collect_rapl(metrics, now)
        self._collect_temperatures(metrics)
        self._collect_cpu_usage(metrics)
        self._collect_cpu_freq(metrics)
        self._collect_memory(metrics)

        return metrics

    def _collect_rapl(self, metrics: Dict[str, float], now: float) -> None:
        """Collect RAPL power metrics as delta watts."""
        for domain_key, domain in self._rapl_domains.items():
            try:
                energy_uj = int(self._read_file(domain["energy_file"]).strip())
                domain_name = domain["name"]

                if domain_key in self._prev_rapl:
                    prev_energy, prev_time = self._prev_rapl[domain_key]
                    dt = now - prev_time
                    if dt > 0:
                        delta_energy = energy_uj - prev_energy
                        # Handle counter wraparound
                        if delta_energy < 0 and domain["max_range"] > 0:
                            delta_energy += domain["max_range"]
                        power_w = (delta_energy / 1_000_000.0) / dt
                        metric_name = f"server_power_{domain_name}_watts"
                        metrics[metric_name] = round(power_w, 3)

                self._prev_rapl[domain_key] = (energy_uj, now)
            except Exception as e:
                self.logger.debug("RAPL read failed for %s: %s", domain_key, e)

    def _collect_temperatures(self, metrics: Dict[str, float]) -> None:
        """Collect CPU temperatures from coretemp hwmon."""
        for label, path in self._temp_inputs.items():
            try:
                raw = int(self._read_file(path).strip())
                temp_c = raw / 1000.0
                # Normalize label: "Package id 0" -> "package", "Core 0" -> "core0"
                normalized = label.lower().replace(" ", "")
                if normalized.startswith("packageid"):
                    key = "package"
                elif normalized.startswith("core"):
                    key = normalized  # e.g. "core0"
                else:
                    key = normalized
                metrics[f"server_temp_{key}_celsius"] = round(temp_c, 2)
            except Exception as e:
                self.logger.debug("Temp read failed for %s: %s", label, e)

    def _collect_cpu_usage(self, metrics: Dict[str, float]) -> None:
        """Collect CPU usage from /proc/stat (delta-based)."""
        try:
            content = self._read_file(Path("/proc/stat"))
        except Exception as e:
            self.logger.debug("Failed to read /proc/stat: %s", e)
            return

        current_times: Dict[str, Tuple[int, int]] = {}

        for line in content.splitlines():
            if not line.startswith("cpu"):
                continue

            parts = line.split()
            cpu_id = parts[0]  # "cpu" (total) or "cpu0", "cpu1", ...
            values = list(map(int, parts[1:]))

            # user, nice, system, idle, iowait, irq, softirq, steal
            if len(values) < 4:
                continue
            idle = values[3] + (values[4] if len(values) > 4 else 0)
            total = sum(values)
            current_times[cpu_id] = (idle, total)

        for cpu_id, (idle, total) in current_times.items():
            if cpu_id in self._prev_cpu_times:
                prev_idle, prev_total = self._prev_cpu_times[cpu_id]
                d_total = total - prev_total
                d_idle = idle - prev_idle
                if d_total > 0:
                    usage = ((d_total - d_idle) / d_total) * 100.0
                    if cpu_id == "cpu":
                        metrics["server_cpu_avg_usage_percent"] = round(usage, 2)
                    else:
                        core_num = cpu_id.replace("cpu", "")
                        metrics[f"server_cpu_core{core_num}_usage_percent"] = round(usage, 2)

        self._prev_cpu_times = current_times

    def _collect_cpu_freq(self, metrics: Dict[str, float]) -> None:
        """Collect per-core CPU frequency from cpufreq."""
        for i in range(self._cpu_count):
            freq_path = Path(f"/sys/devices/system/cpu/cpu{i}/cpufreq/scaling_cur_freq")
            try:
                freq_khz = int(self._read_file(freq_path).strip())
                metrics[f"server_cpu_core{i}_freq_mhz"] = round(freq_khz / 1000.0, 1)
            except Exception:
                pass  # CPU might be offline

    def _collect_memory(self, metrics: Dict[str, float]) -> None:
        """Collect memory info from host procfs."""
        # Use /host/proc/meminfo if available (container with host mount),
        # otherwise fall back to /proc/meminfo
        meminfo_path = Path("/host/proc/meminfo")
        if not meminfo_path.exists():
            meminfo_path = Path("/proc/meminfo")

        try:
            content = self._read_file(meminfo_path)
        except Exception as e:
            self.logger.debug("Failed to read meminfo: %s", e)
            return

        mem = {}
        for line in content.splitlines():
            parts = line.split()
            if len(parts) >= 2:
                key = parts[0].rstrip(":")
                # Values are in kB
                mem[key] = int(parts[1])

        total_kb = mem.get("MemTotal", 0)
        available_kb = mem.get("MemAvailable", 0)

        if total_kb > 0:
            total_mb = total_kb / 1024.0
            available_mb = available_kb / 1024.0
            used_mb = total_mb - available_mb

            metrics["server_ram_total_mb"] = round(total_mb, 1)
            metrics["server_ram_used_mb"] = round(used_mb, 1)
            metrics["server_ram_available_mb"] = round(available_mb, 1)
            metrics["server_ram_used_percent"] = round((used_mb / total_mb) * 100.0, 2)

    @staticmethod
    def _read_file(path, default=None) -> str:
        """Read a file, returning default on failure if provided."""
        try:
            return Path(path).read_text()
        except Exception:
            if default is not None:
                return default
            raise
