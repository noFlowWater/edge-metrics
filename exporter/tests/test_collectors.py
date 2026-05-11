"""Tests for collector implementations."""
import sys
import os
import pytest

# Add exporter root to path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

from collectors import get_collector, BaseCollector
from collectors.stub import StubCollector
from collectors.jetson_orin_nano import JetsonOrinNanoCollector


class TestStubCollector:
    def test_returns_metrics(self):
        cfg = {"device_type": "stub", "port": 9102, "reload_port": 9101}
        c = StubCollector(cfg)
        metrics = c.get_metrics()
        assert len(metrics) > 0

    def test_metric_names_match(self):
        cfg = {"device_type": "stub"}
        c = StubCollector(cfg)
        metrics = c.get_metrics()
        for name in StubCollector.metric_names():
            assert name in metrics, f"missing metric: {name}"

    def test_has_jetson_metrics(self):
        cfg = {"device_type": "stub"}
        c = StubCollector(cfg)
        metrics = c.get_metrics()
        assert "jetson_power_vdd_gpu_soc_watts" in metrics
        assert "jetson_temp_cpu_celsius" in metrics
        assert "jetson_gpu_utilization_percent" in metrics

    def test_has_shelly_metrics(self):
        cfg = {"device_type": "stub"}
        c = StubCollector(cfg)
        metrics = c.get_metrics()
        assert "shelly_power_watts" in metrics
        assert "shelly_voltage_volts" in metrics

    def test_values_are_positive(self):
        cfg = {"device_type": "stub"}
        c = StubCollector(cfg)
        metrics = c.get_metrics()
        for name, value in metrics.items():
            assert value > 0, f"{name} should be positive, got {value}"

    def test_safe_get_metrics(self):
        cfg = {"device_type": "stub"}
        c = StubCollector(cfg)
        metrics = c.safe_get_metrics()
        assert len(metrics) > 0

    def test_custom_amplitude(self):
        cfg = {"device_type": "stub", "stub_config": {"amplitude": 0.5, "period": 10}}
        c = StubCollector(cfg)
        metrics = c.get_metrics()
        assert len(metrics) > 0


class TestCollectorFactory:
    def test_get_stub_collector(self):
        c = get_collector("stub", {"device_type": "stub"})
        assert isinstance(c, StubCollector)

    def test_get_orin_nano_collector(self):
        c = get_collector("jetson_orin_nano", {"device_type": "jetson_orin_nano"})
        assert isinstance(c, JetsonOrinNanoCollector)

    def test_get_orin_nano_alias(self):
        c = get_collector("orin_nano", {"device_type": "orin_nano"})
        assert isinstance(c, JetsonOrinNanoCollector)

    def test_unknown_device_type(self):
        with pytest.raises(ValueError, match="Unsupported device type"):
            get_collector("unknown_xyz", {})

    def test_get_collector_returns_base(self):
        c = get_collector("stub", {"device_type": "stub"})
        assert isinstance(c, BaseCollector)


class TestJetsonOrinNanoCollector:
    def test_parses_jetpack6_tegrastats_line(self):
        sample = (
            "05-12-2026 00:30:15 RAM 3423/7620MB (lfb 78x4MB) "
            "SWAP 0/3810MB (cached 0MB) "
            "CPU [2%@729,6%@729,2%@729,1%@729,2%@729,2%@729] "
            "GR3D_FREQ 0% "
            "cpu@45.531C soc2@44.781C soc0@45.937C gpu@46.468C "
            "tj@46.468C soc1@46C "
            "VDD_IN 3311mW/3311mW VDD_CPU_GPU_CV 524mW/524mW "
            "VDD_SOC 1008mW/1008mW"
        )
        c = JetsonOrinNanoCollector({"device_type": "jetson_orin_nano"})
        metrics = c._parse_all_metrics(sample)

        assert metrics["jetson_ram_used_mb"] == 3423.0
        assert metrics["jetson_ram_total_mb"] == 7620.0
        assert metrics["jetson_ram_used_percent"] == pytest.approx(44.92)
        assert metrics["jetson_swap_used_mb"] == 0.0
        assert metrics["jetson_swap_total_mb"] == 3810.0
        assert metrics["jetson_swap_cached_mb"] == 0.0
        assert metrics["jetson_lfb_blocks"] == 78
        assert metrics["jetson_lfb_total_mb"] == 312

        assert metrics["jetson_cpu_active_cores"] == 6
        assert metrics["jetson_cpu_core0_usage_percent"] == 2
        assert metrics["jetson_cpu_core1_usage_percent"] == 6
        assert metrics["jetson_cpu_core0_freq_mhz"] == 729
        assert metrics["jetson_cpu_avg_usage_percent"] == pytest.approx(2.5)

        assert metrics["jetson_gpu_usage_percent"] == 0
        assert "jetson_gpu_freq0_mhz" not in metrics

        assert metrics["jetson_temp_cpu_celsius"] == pytest.approx(45.53)
        assert metrics["jetson_temp_gpu_celsius"] == pytest.approx(46.47)
        assert metrics["jetson_temp_tj_celsius"] == pytest.approx(46.47)
        assert metrics["jetson_temp_soc0_celsius"] == pytest.approx(45.94)
        assert metrics["jetson_temp_soc1_celsius"] == pytest.approx(46.0)
        assert metrics["jetson_temp_soc2_celsius"] == pytest.approx(44.78)

        assert metrics["jetson_power_vdd_in_watts"] == pytest.approx(3.311)
        assert metrics["jetson_power_vdd_in_avg_watts"] == pytest.approx(3.311)
        assert metrics["jetson_power_vdd_cpu_gpu_cv_watts"] == pytest.approx(0.524)
        assert metrics["jetson_power_vdd_cpu_gpu_cv_avg_watts"] == pytest.approx(0.524)
        assert metrics["jetson_power_vdd_soc_watts"] == pytest.approx(1.008)
        assert metrics["jetson_power_vdd_soc_avg_watts"] == pytest.approx(1.008)

    def test_parses_optional_gpu_frequency_forms(self):
        c = JetsonOrinNanoCollector({"device_type": "jetson_orin_nano"})

        single = c._parse_all_metrics("RAM 1/2MB CPU [0%@729] GR3D_FREQ 7%@918")
        assert single["jetson_gpu_usage_percent"] == 7
        assert single["jetson_gpu_freq0_mhz"] == 918

        bracketed = c._parse_all_metrics("RAM 1/2MB CPU [0%@729] GR3D_FREQ 8%@[510,0]")
        assert bracketed["jetson_gpu_usage_percent"] == 8
        assert bracketed["jetson_gpu_freq0_mhz"] == 510
        assert bracketed["jetson_gpu_freq1_mhz"] == 0
