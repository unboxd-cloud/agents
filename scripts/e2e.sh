#!/usr/bin/env bash
# End-to-end test: build the binaries, run the control plane with datasets, and
# exercise the full flow (catalog registry, pay-as-you-go rating + reseller +
# tax, compliance evaluation, metrics). Exits non-zero on any failure.
set -euo pipefail
cd "$(dirname "$0")/.."

DS=deploy/datasets
fail() { echo "E2E FAIL: $*" >&2; exit 1; }

echo "== build =="
make build >/dev/null

echo "== start services =="
CATALOG_DATASET=$DS/offerings.json ./bin/catalog >/tmp/e2e-catalog.log 2>&1 &
PRICEBOOK_DATASET=$DS/pricebook.json TAX_DATASET=$DS/tax-rules.json ./bin/billing >/tmp/e2e-billing.log 2>&1 &
COMPLIANCE_DATASET=$DS/compliance-frameworks.json ./bin/compliance >/tmp/e2e-compliance.log 2>&1 &
./bin/metering >/tmp/e2e-metering.log 2>&1 &
PIDS=$(jobs -p)
trap 'kill $PIDS 2>/dev/null || true' EXIT

# wait for readiness
for p in 8083 8082 8084 8081; do
  for i in $(seq 1 20); do
    curl -fs "localhost:$p/readyz" >/dev/null 2>&1 && break
    [ "$i" = 20 ] && fail "service on :$p not ready"
    sleep 0.25
  done
done

echo "== catalog registry =="
cats=$(curl -fs localhost:8083/v1/categories)
echo "  categories: $cats"
echo "$cats" | grep -q '"ai"' || fail "expected ai category"
curl -fs "localhost:8083/v1/catalog?category=ai" | grep -q '"bedrock"' || fail "expected bedrock offering"

echo "== rating: vcpu+gpu, reseller 15%, EU-DE VAT =="
resp=$(curl -fs -X POST localhost:8082/v1/rate -d '{
  "tenantId":"e2e","jurisdiction":"EU-DE",
  "events":[{"tenantId":"e2e","meter":"compute.vcpu.hour","quantity":1300},
            {"tenantId":"e2e","meter":"ai.cpu.hour","quantity":50}],
  "partner":{"id":"p1","mode":"reseller","rate":0.15}}')
echo "  $resp" | head -c 200; echo
echo "$resp" | grep -q '"grossToCustomer":56.58' || fail "unexpected settlement (want grossToCustomer 56.58)"
echo "$resp" | grep -q '"gross":67.33' || fail "unexpected taxed gross (want 67.33)"

echo "== compliance: GDPR residency violation must be blocked =="
code=$(curl -s -o /tmp/e2e-eval.json -w '%{http_code}' -X POST localhost:8084/v1/evaluate -d '{
  "profile":{"tenantId":"e2e","frameworks":["GDPR"],"dataResidency":["EU-DE"]},
  "placement":{"region":"US-CA","certifications":["GDPR"],"encrypted":true}}')
[ "$code" = "422" ] || fail "expected 422 for residency violation, got $code"

echo "== metrics published =="
curl -fs localhost:8083/metrics | grep -q '^platform_up 1' || fail "metrics not published"

echo "E2E PASS"
