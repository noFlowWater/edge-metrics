# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 0.x.x   | :white_check_mark: |

## Reporting a Vulnerability

If you discover a security vulnerability in this project, please report it responsibly.

**Do NOT open a public GitHub issue for security vulnerabilities.**

Instead, please email: **noflowwater@gmail.com**

Include the following details:
- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

We will acknowledge receipt within 48 hours and aim to provide a fix within 7 days for critical issues.

## Security Best Practices

When deploying edge-metrics:

1. **Use non-root containers** — all containers run as non-root by default (exception: the Jetson exporter profile requires `runAsNonRoot: false` and `SYS_RAWIO` capability for tegrastats access)
2. **Enable read-only root filesystem** — configured in default security contexts
3. **Use network policies** — restrict pod-to-pod communication
4. **Keep images updated** — use Dependabot/Renovate for dependency updates
5. **Scan images** — Trivy scanning is included in CI pipeline
6. **Use Sealed Secrets** — for any sensitive configuration values
