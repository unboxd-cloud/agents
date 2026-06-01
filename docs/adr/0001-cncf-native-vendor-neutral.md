# ADR-0001: CNCF-native and vendor-neutral

## Status
Accepted

## Context
The platform must deliver a full cloud experience but must not lock into a single
vendor (the original brief named Apache CloudStack). It should also avoid
rebuilding infrastructure that the CNCF ecosystem already provides well.

## Decision
1. **Compose, don't construct.** Each capability maps to an existing CNCF
   project (see `docs/cncf-stack.md`). We own only thin control-plane glue.
2. **Vendor neutrality via one seam.** All infrastructure is reached through a
   single `provider.Provider` interface. Apache CloudStack, Kubernetes, and
   public clouds are interchangeable implementations behind it — none is the
   foundation.

## Consequences
- No provider-specific logic leaks into control-plane services; swapping vendors
  is adding a `Provider`, not editing call sites (less duplication).
- We inherit upgrades/security from upstream CNCF projects.
- We must track upstream versions and run conformance against each provider.
