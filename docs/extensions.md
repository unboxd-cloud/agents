# Extensions: adapters, plugins & protocols

Everything pluggable goes through a small set of **seams** (interfaces). An
extension is any Go type that satisfies a seam. The platform supports adding them
without modifying core code.

## Seams

| Seam | Interface | Example extensions |
|------|-----------|--------------------|
| Infrastructure | `provider.Provider` | kubernetes, cloudstack, edge, public clouds |
| Usage (pull) | `metering.Source` | OpenCost, Prometheus |
| Usage (push/stream) | `metering.StreamSource` | NATS/CloudEvents, OTLP, Kafka |
| Policy gate | `authz.Gate` | OPA / Gatekeeper |
| Fine-grained authz | `authz.RelationChecker` | OpenFGA |
| Tax | `billing.TaxTable` / `[]TaxRule` source | external tax engines |
| Compliance | `compliance` specs source | framework dataset providers |
| Protocol | adapter registered under `plugin.KindProtocol` | HTTP, gRPC, NATS, OTLP, Kafka |

## Native-runtime path (recommended)

Compile the plugin into the binary using the `database/sql` driver model:

1. Implement the seam interface in a package under `plugins/`.
2. Self-register in `init()` via `plugin.Register`.
3. Blank-import the package in the service you want to extend:

```go
import _ "github.com/unboxd-cloud/platform/plugins/echoprovider"
```

4. Build/instantiate by name through the registry:

```go
v, _ := plugin.New(plugin.KindProvider, "echo", map[string]string{"name": "edge-1"})
p := v.(provider.Provider) // assert to the seam interface
```

See `plugins/echoprovider/` for a complete, tested template. This path is
type-safe, cross-platform, and needs no dynamic loader — the extension is part
of the native runtime.

## Why not Go `plugin` (.so)?
Dynamic `.so` plugins exist but are platform-fragile, require identical
toolchain/deps, and don't work everywhere. Compile-time registration is the
supported default; `.so` can be layered on later for an out-of-tree story.

## Protocol adapters
Protocols are extensions too: register an adapter under `plugin.KindProtocol`
that bridges a wire protocol (e.g. OTLP, NATS, Kafka) into a `metering.Source`/
`StreamSource` or other seam. The control plane stays protocol-agnostic.
