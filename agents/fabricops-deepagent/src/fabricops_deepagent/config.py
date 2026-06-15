from __future__ import annotations

import os
from dataclasses import dataclass


@dataclass(frozen=True)
class FabricOpsConfig:
    mode: str
    surreal_url: str
    surreal_ns: str
    surreal_db: str
    opa_url: str
    openfga_api_url: str

    @classmethod
    def from_env(cls) -> "FabricOpsConfig":
        return cls(
            mode=os.getenv("FABRICOPS_MODE", "observe"),
            surreal_url=os.getenv("SURREAL_URL", "ws://surrealdb.fabric.svc.cluster.local:8000"),
            surreal_ns=os.getenv("SURREAL_NS", "agennext"),
            surreal_db=os.getenv("SURREAL_DB", "fabric"),
            opa_url=os.getenv("OPA_URL", "http://opa.open-policy-agent.svc.cluster.local:8181"),
            openfga_api_url=os.getenv("OPENFGA_API_URL", "http://openfga.openfga.svc.cluster.local:8080"),
        )
