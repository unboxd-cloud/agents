#!/usr/bin/env bash
# Post-production sanity: lightweight, read-only checks against a running
# deployment. Safe to run after every deploy. Configure hosts via env (defaults
# to local ports). Exits non-zero if any check fails.
set -uo pipefail

CATALOG=${CATALOG_URL:-http://localhost:8083}
METERING=${METERING_URL:-http://localhost:8081}
BILLING=${BILLING_URL:-http://localhost:8082}
COMPLIANCE=${COMPLIANCE_URL:-http://localhost:8084}

rc=0
check() { # name url
  if curl -fs --max-time 5 "$2" >/dev/null 2>&1; then
    echo "ok   $1 ($2)"
  else
    echo "FAIL $1 ($2)"; rc=1
  fi
}

echo "== liveness/readiness/metrics =="
for pair in "catalog $CATALOG" "metering $METERING" "billing $BILLING" "compliance $COMPLIANCE"; do
  set -- $pair
  check "$1 healthz" "$2/healthz"
  check "$1 readyz"  "$2/readyz"
  check "$1 metrics" "$2/metrics"
done

echo "== functional smoke =="
check "catalog list"   "$CATALOG/v1/catalog"
check "pricebook"      "$BILLING/v1/pricebook"
check "frameworks"     "$COMPLIANCE/v1/frameworks"

if [ "$rc" = 0 ]; then echo "POST-DEPLOY SANITY: PASS"; else echo "POST-DEPLOY SANITY: FAIL"; fi
exit $rc
