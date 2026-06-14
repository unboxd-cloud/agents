from fabricops_deepagent.decision import DecisionRecord


def test_decision_record_has_required_agent_identity():
    record = DecisionRecord(
        subagent="kubernetes_agent",
        mode="observe",
        action="check_k3s_pods",
        reason="test",
        policy_allowed=True,
    )

    event = record.as_fabric_event()

    assert event["agent"] == "fabricops-deepagent"
    assert event["subagent"] == "kubernetes_agent"
    assert event["mode"] == "observe"
    assert event["action"] == "check_k3s_pods"
    assert event["timestamp"]
