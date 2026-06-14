from __future__ import annotations

from datetime import UTC, datetime
from typing import Any

from pydantic import BaseModel, Field


class DecisionRecord(BaseModel):
    agent: str = "fabricops-deepagent"
    subagent: str
    mode: str
    action: str
    reason: str
    risk: str = "low"
    policy_allowed: bool = False
    approval_required: bool = False
    evidence: dict[str, Any] = Field(default_factory=dict)
    result: str = "pending"
    timestamp: str = Field(default_factory=lambda: datetime.now(UTC).isoformat())

    def as_fabric_event(self) -> dict[str, Any]:
        return self.model_dump()
