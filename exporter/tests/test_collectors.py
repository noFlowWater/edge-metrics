"""Tests for collector implementations."""
import sys
import os
import pytest

# Add exporter root to path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

from collectors import get_collector, BaseCollector
from collectors.stub import StubCollector


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

    def test_unknown_device_type(self):
        with pytest.raises(ValueError, match="Unsupported device type"):
            get_collector("unknown_xyz", {})

    def test_get_collector_returns_base(self):
        c = get_collector("stub", {"device_type": "stub"})
        assert isinstance(c, BaseCollector)
