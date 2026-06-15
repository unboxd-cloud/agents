#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

export FABRICOPS_MODE="${FABRICOPS_MODE:-observe}"
export SURREAL_URL="${SURREAL_URL:-ws://127.0.0.1:8000}"
export SURREAL_NS="${SURREAL_NS:-agennext}"
export SURREAL_DB="${SURREAL_DB:-fabric}"
export OPA_URL="${OPA_URL:-http://127.0.0.1:8181}"
export OPENFGA_API_URL="${OPENFGA_API_URL:-http://127.0.0.1:8080}"

python -m fabricops_deepagent.main
