from __future__ import annotations

import httpx

from .config import FabricOpsConfig


async def check_policy(config: FabricOpsConfig, payload: dict) -> bool:
    """Check OPA when available. Fail closed for non-observe modes."""
    if config.mode == "observe":
        return True

    url = f"{config.opa_url.rstrip('/')}/v1/data/fabricops/authz/allow"
    try:
        async with httpx.AsyncClient(timeout=5) as client:
            response = await client.post(url, json={"input": payload})
            response.raise_for_status()
            data = response.json()
            return bool(data.get("result", False))
    except Exception:
        return False
