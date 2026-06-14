from __future__ import annotations

import asyncio
import json

from .config import FabricOpsConfig
from .decision import DecisionRecord
from .policy import check_policy


OBSERVE_CHECKS = [
    ("vps_agent", "check_vps_health", "Collect VPS host health evidence"),
    ("kubernetes_agent", "check_k3s_nodes", "Collect k3s node health evidence"),
    ("kubernetes_agent", "check_k3s_pods", "Collect k3s pod health evidence"),
    ("surrealdb_agent", "check_surrealdb_health", "Collect Fabric database health evidence"),
    ("security_agent", "check_security_posture", "Collect security posture evidence"),
    ("cost_agent", "check_cost_utilization", "Collect utilization and cost evidence"),
]


async def run_once() -> list[DecisionRecord]:
    config = FabricOpsConfig.from_env()
    records: list[DecisionRecord] = []

    for subagent, action, reason in OBSERVE_CHECKS:
        payload = {"mode": config.mode, "action": action, "execute": False}
        allowed = await check_policy(config, payload)
        record = DecisionRecord(
            subagent=subagent,
            mode=config.mode,
            action=action,
            reason=reason,
            risk="low",
            policy_allowed=allowed,
            approval_required=False,
            evidence={"status": "scaffold", "runtime": "deepagents"},
            result="observed" if allowed else "blocked",
        )
        records.append(record)

    return records


def main() -> None:
    records = asyncio.run(run_once())
    for record in records:
        print(json.dumps(record.as_fabric_event(), sort_keys=True))


if __name__ == "__main__":
    main()
