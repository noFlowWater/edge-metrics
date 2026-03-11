# Contributing to Edge Metrics

## Development Setup

### Prerequisites

- Go 1.25+
- Python 3.12+
- Helm 3.x
- Docker
- k3d (for local cluster testing)

### Local Development

```bash
# Clone
git clone https://github.com/noFlowWater/edge-metrics.git
cd edge-metrics

# Server (Go)
cd server
go mod download
go run . # starts on :8081

# Exporter (Python) - Stub mode
cd exporter
pip install -r requirements.txt
STUB_MODE=true python exporter.py

# Frontend
cd front
npm install
npm run dev
```

### Running Tests

```bash
make test-all       # Run all tests
make test-server    # Go tests only
make test-exporter  # Python tests only
make test-helm      # Helm chart tests
make kubeconform    # K8s manifest validation
```

### Local Cluster (k3d)

```bash
bash e2e/setup.sh              # Create k3d cluster
bash e2e/test_install.sh       # Install via Helm
bash e2e/test_device_lifecycle.sh  # Run e2e tests
bash e2e/teardown.sh           # Cleanup
```

## Code Structure

```
edge-metrics/
├── server/          # Go backend (config management, K8s sync)
├── exporter/        # Python metrics exporter (Prometheus)
├── front/           # React frontend dashboard
├── charts/          # Helm umbrella chart
├── e2e/             # End-to-end test scripts
└── .github/         # CI/CD workflows
```

## Pull Request Process

1. Fork the repository
2. Create a feature branch from `main`
3. Make your changes
4. Ensure all tests pass: `make test-all`
5. Submit a PR with a clear description

### PR Requirements

- All CI checks must pass
- Helm lint + kubeconform must pass
- New features should include tests where applicable
