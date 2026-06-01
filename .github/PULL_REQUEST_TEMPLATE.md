<!-- Thanks for contributing! Keep PRs focused; open as draft until green. -->

## What & why
<!-- What does this change and why? Link issues. -->

## Type
- [ ] Feature  - [ ] Fix  - [ ] Docs  - [ ] Dataset (offering/pricing/policy)  - [ ] CI/infra

## Checklist
- [ ] Upholds `docs/design-principles.md` and relevant ADRs
- [ ] `make check` passes (vet + tests)
- [ ] `make e2e` passes (if behavior changed)
- [ ] `gofmt` clean
- [ ] Docs updated
- [ ] `CHANGELOG.md` (Unreleased) updated for user-facing changes
- [ ] New offering = dataset edit (not code), where applicable
- [ ] New integration = plugin (not core edit), where applicable

## Production gate (if this publishes/deploys)
- [ ] Dependencies resolve (no cycle / unmet)
- [ ] Human sign-off recorded (LLM review may assist)
