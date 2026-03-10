"""
Stub collector for testing without hardware.
Generates synthetic metrics matching real collector output.
"""
import math
import time
from typing import Dict, List
from .base import BaseCollector


class StubCollector(BaseCollector):
    """
    Generates synthetic Prometheus metrics for all supported device types.
    Useful for Helm chart testing, CI/CD, and development without hardware.
    """

    METRICS = {
        # Jetson Orin power rails (watts)
        "jetson_power_vdd_gpu_soc_watts": 3.2,
        "jetson_power_vdd_cpu_cv_watts": 4.1,
        "jetson_power_vddq_vdd2_1v8ao_watts": 1.5,
        "jetson_power_vin_sys_5v0_watts": 12.8,
        # Jetson temperatures (celsius)
        "jetson_temp_cpu_celsius": 42.0,
        "jetson_temp_gpu_celsius": 40.5,
        "jetson_temp_soc0_celsius": 39.0,
        "jetson_temp_soc1_celsius": 38.5,
        "jetson_temp_tj_celsius": 43.0,
        # Jetson utilization (percent)
        "jetson_gpu_utilization_percent": 25.0,
        "jetson_cpu_utilization_percent": 30.0,
        # Jetson memory (MB)
        "jetson_ram_used_mb": 5848.0,
        "jetson_ram_total_mb": 62801.0,
        # Shelly power (watts)
        "shelly_power_watts": 85.0,
        "shelly_voltage_volts": 220.5,
        "shelly_current_amps": 0.39,
        "shelly_energy_wh": 1250.0,
    }

    def __init__(self, config: dict):
        super().__init__(config)
        self._start = time.time()
        stub_cfg = config.get("stub_config", {})
        self._amplitude = stub_cfg.get("amplitude", 0.1)
        self._period = stub_cfg.get("period", 60)

    def get_metrics(self) -> Dict[str, float]:
        elapsed = time.time() - self._start
        wave = math.sin(2 * math.pi * elapsed / self._period) * self._amplitude
        return {k: round(v * (1.0 + wave), 4) for k, v in self.METRICS.items()}

    @classmethod
    def metric_names(cls) -> List[str]:
        return list(cls.METRICS.keys())
