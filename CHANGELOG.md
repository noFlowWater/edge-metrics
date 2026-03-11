# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed
- Translate all Korean documentation to English for public repository consistency
- Update Go version requirement from 1.23+ to 1.25+ in docs
- Update Python version requirement from 3.10+ to 3.12+ in docs
- Unify Kubernetes version requirement to 1.26+ across all docs
- Fix default exporter port in docs from 9100 to 9102
- Add API Key authentication documentation to server API.md
- Promote Helm as primary deployment method, deprecate shell scripts
- Standardize Kubernetes label selectors to `app.kubernetes.io/name`
- Fix Makefile to gracefully handle missing values-cluster.yaml

## [0.1.0] - 2026-03-01

### Added
- Initial Helm umbrella chart with server, exporter, and frontend subcharts
- OCI registry publishing via GitHub Actions
- End-to-end test suite with k3d
- Values layering (dev, prod, cluster-specific)
- API Key authentication middleware for server
- Grafana dashboard auto-loading via sidecar
- PrometheusRule alert definitions
