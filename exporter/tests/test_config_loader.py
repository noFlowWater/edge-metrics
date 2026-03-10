"""Tests for ConfigLoader."""
import sys
import os
import tempfile
import pytest
import yaml

sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

from config_loader import ConfigLoader


class TestConfigLoader:
    def test_load_local_yaml(self, tmp_path):
        config_file = tmp_path / "config.yaml"
        config_file.write_text(yaml.dump({
            "device_type": "stub",
            "port": 9102,
            "reload_port": 9101,
        }))

        os.environ["LOCAL_CONFIG_PATH"] = str(config_file)
        os.environ["CONFIG_SERVER_URL"] = "http://localhost:99999"  # unreachable
        loader = ConfigLoader()
        config = loader.load()

        assert config["device_type"] == "stub"
        assert config["port"] == 9102

    def test_save_to_local(self, tmp_path):
        config_file = tmp_path / "config.yaml"
        os.environ["LOCAL_CONFIG_PATH"] = str(config_file)
        loader = ConfigLoader()
        result = loader.save_to_local({"device_type": "test", "port": 1234})
        assert result is True
        saved = yaml.safe_load(config_file.read_text())
        assert saved["device_type"] == "test"

    def test_load_returns_none_without_config(self, tmp_path):
        os.environ["LOCAL_CONFIG_PATH"] = str(tmp_path / "nonexistent.yaml")
        os.environ["CONFIG_SERVER_URL"] = "http://localhost:99999"
        os.environ["CONFIG_MAX_RETRIES"] = "1"
        os.environ["CONFIG_RETRY_INTERVAL"] = "0"
        loader = ConfigLoader()
        assert loader.load() is None
