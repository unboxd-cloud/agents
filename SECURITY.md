# Security Policy

## Reporting a vulnerability
**Do not open a public issue for security vulnerabilities.**

Report privately via:
- GitHub **Security Advisories** (Security tab → "Report a vulnerability"), or
- email **security@unboxd.cloud**.

Please include: affected component/version, reproduction steps, impact, and any
suggested remediation. We aim to acknowledge within 3 business days and to
provide a remediation timeline after triage.

## Supported versions
Pre-1.0: the latest `main` is supported. Security fixes target `main` and the
most recent release.

## Our practices
- Static, non-root, read-only-rootfs **scratch** images on the latest patched Go.
- CI runs **govulncheck** (Go vulns) and **Trivy** (image HIGH/CRITICAL).
- Secret scanning via GitHub native push protection (enable in repo settings).
- Supply chain: SBOM + SLSA provenance + image signing are on the roadmap
  (`docs/versioning.md`).
- Least privilege: tenant isolation, OPA policy gate, OpenFGA fine-grained authz.

## Disclosure
We follow coordinated disclosure: we will credit reporters (unless anonymity is
requested) and publish an advisory once a fix is available.
