# Publishing Routes (multi-cloud marketplace distribution)

Publish an offering once in our marketplace, then **route** that listing to
external marketplaces and clouds. Each route is an extension registered under
`plugin.KindPublishRoute`, so adding a destination is a plugin, not a core change.

## How a route works
1. A catalog `Offering` (project + composition + meters + certifications) is the
   single source of truth.
2. A route connector maps the offering to the destination's listing model and
   syncs metered dimensions to that destination's metering API.
3. Pay-as-you-go usage flows back through **one** billing engine; the destination
   is settled as a `marketplace` operating mode (commission via `Settle`),
   honoring publisher `RevShare`.

One listing, one settlement path, many storefronts.

## Routes
| Route | Mechanism | Also a provider? |
|-------|-----------|------------------|
| **AWS** Marketplace | SaaS/container listing + Marketplace Metering | yes (`aws`) |
| **GCP** Marketplace | Producer Portal + Service Control metering | yes (k8s/GKE) |
| **Azure** Marketplace | Partner Center + Marketplace Metering | yes (k8s/AKS) |
| **DigitalOcean** Marketplace | 1-Click app / Marketplace listing | yes (k8s/DOKS) |
| **OpenStack** | App Catalog (Murano) / Glance images | yes (provider) |
| **OpenNebula** | Marketplace appliances | yes (provider) |
| **MicroCloud** (Canonical) | snap/LXD profiles | yes (provider) |
| **Apache CloudStack** | template/offering catalog | yes (`cloudstack`) |
| **Bare-metal / Linux** | OCI image + Helm/compose bundle | yes (k8s/k3s) |

Because most destinations are also **providers** behind the vendor-neutral seam,
the platform can both *publish to* and *run on* each — interoperable in both
directions (see `docs/aws-interop.md`).

## Status
The publishing-route connectors are tracked in `docs/tracker.md`. They build on
existing pieces — catalog offerings, metered dimensions, and marketplace
settlement — which are already implemented; each route is an incremental plugin.
