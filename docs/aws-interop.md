# AWS Interoperability

**Tagline: the open-source AWS alternative — completely interoperable with AWS.**

Interoperable means existing AWS tools, SDKs, and workloads keep working, and you
can run on-prem/any-cloud/edge or AWS itself, then migrate either direction
without rewrites.

## How interoperability is achieved

| AWS surface | Open-source equivalent | Interop mechanism |
|-------------|------------------------|-------------------|
| S3 (object storage) | Rook/Ceph RGW, MinIO | **S3-compatible API**; point the AWS SDK at our endpoint (`--endpoint-url`) |
| EKS (Kubernetes) | Any conformant Kubernetes / k3s | standard Kubernetes API |
| IAM | Dex (OIDC) + OpenFGA + OPA | STS/OIDC federation; policy-compatible model |
| EC2 | Kubernetes / CloudStack / edge providers | provider seam |
| RDS | Postgres/MySQL operators (catalog) | standard wire protocols |
| SQS/SNS | NATS | adapter; CloudEvents |
| CloudWatch | Prometheus + OpenTelemetry | OTLP/Prometheus export |
| Cost Explorer | OpenCost + billing | FOCUS-based cost/usage |
| Bedrock | KServe + open-source CPU LLMs | OpenAI/Bedrock-style inference API |

## Two ways to use AWS
1. **As a provider** — register `aws` (in `DefaultRegistry`) and provision *on*
   AWS through the same control plane (lift-and-extend).
2. **As a target to replace/migrate from** — run the open-source equivalents
   anywhere; because the data-plane APIs are AWS-compatible (S3, Kubernetes,
   SQL), clients switch by changing an endpoint, not their code.

## Drop-in example (S3)
```bash
aws s3 ls --endpoint-url https://objectstore.platform.example
# same AWS CLI / boto3 / aws-sdk-go, different endpoint
```

## AWS Marketplace — direct publishing
Offerings in our marketplace can be **published directly to AWS Marketplace**, so
a listing created once is sold in both places:
- A publishing connector (an extension under `plugin.KindProtocol`) maps a
  catalog `Offering` → an AWS Marketplace product (SaaS/container listing) and
  syncs metered dimensions to AWS Marketplace Metering.
- Pay-as-you-go usage flows back through the same billing engine; AWS Marketplace
  is settled as a `marketplace` operating mode (commission via `Settle`).
- Revenue-share (`Offering.RevShare`) applies to third-party publishers exactly
  as in our own marketplace — one settlement path, two storefronts.

Status: connector is on the tracker (`docs/tracker.md`); the catalog,
metering-dimension, and settlement pieces it builds on already exist.

## Principles
- **No proprietary lock-in:** every AWS-equivalent service is an open-source
  CNCF/landscape project behind the catalog.
- **Wire-compatible first:** prefer matching the AWS API shape (S3, OIDC/STS,
  SQL) so migration is endpoint-level.
- **Bidirectional:** the `aws` provider means you can also *use* AWS where it
  makes sense — interoperable, not adversarial.
