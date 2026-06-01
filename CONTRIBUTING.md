# Contributing

Thanks for helping build the open-source AWS alternative! This project is
**community-driven** and welcomes issues, discussions, and pull requests.

## Principles first
Read [`docs/design-principles.md`](docs/design-principles.md) and the ADRs in
[`docs/adr/`](docs/adr/). Changes should uphold them (vendor-neutral, compose
don't rebuild, data-not-code, headless-first, design-for-failure, …).

## Dev setup
```bash
make check     # go vet + tests (the CI gate)
make build     # build all binaries
make e2e       # end-to-end run + flow checks
make sandbox-up / sandbox-down   # local podman stack
```
Requirements: Go 1.24+ (CI uses latest stable), podman for images/sandbox.

## Pull requests
1. Branch from `main`; keep PRs focused.
2. `make check` and `make e2e` must pass; run `gofmt`.
3. Add/adjust tests for behavior changes (engines are pure — test them).
4. Update docs and `CHANGELOG.md` (Unreleased) for user-facing changes.
5. Prefer **data** edits (datasets) over code for new offerings/pricing/policy.
6. New integrations are **plugins** (`docs/extensions.md`), not core edits.
7. Open as a **draft** until green; one logical change per PR.

## Publishing changes (production = enterprise)
Anything that publishes/deploys passes the **production gate**
([`docs/production-gate.md`](docs/production-gate.md)): dependency resolution +
human approval + accountability. LLM review assists; a human signs off.

## Adding an offering (most common contribution)
Edit `deploy/datasets/offerings.json` (and price it in `pricebook.json`); map
compliance via `certifications`. No code change needed — see
[`docs/registries.md`](docs/registries.md).

## Communication
- Issues: bugs/features (templates provided).
- Security: see [`SECURITY.md`](SECURITY.md) — do not open public issues for
  vulnerabilities.
- Be respectful — see [`CODE_OF_CONDUCT.md`](CODE_OF_CONDUCT.md).

## License
By contributing you agree your contributions are licensed under
[Apache-2.0](LICENSE).
