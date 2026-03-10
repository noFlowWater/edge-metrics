import pytest
from prometheus_client import REGISTRY, CollectorRegistry


@pytest.fixture(autouse=True)
def clean_registry():
    """Prevent prometheus_client registry pollution between tests."""
    yield
