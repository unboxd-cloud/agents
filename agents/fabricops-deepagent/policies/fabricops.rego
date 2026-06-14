package fabricops.authz

import future.keywords.if
import future.keywords.in

default allow := false

destructive_actions := {
  "delete_namespace",
  "delete_pvc",
  "delete_pv",
  "delete_database",
  "delete_node",
  "rotate_secret",
  "open_firewall_port",
  "disable_security_control",
  "force_push_branch",
  "deploy_unscanned_image",
}

safe_observe_actions := {
  "check_vps_health",
  "check_k3s_nodes",
  "check_k3s_pods",
  "check_k3s_events",
  "check_surrealdb_health",
  "check_github_actions",
  "check_security_posture",
  "check_cost_utilization",
}

allow if {
  input.mode == "observe"
  input.action in safe_observe_actions
}

allow if {
  input.mode == "recommend"
  not input.execute
}

allow if {
  input.mode == "approve"
  input.approval.id != ""
  input.approval.status == "approved"
  input.approval.actor != ""
}

allow if {
  input.mode == "auto-safe"
  input.action in safe_observe_actions
}

requires_approval if {
  input.action in destructive_actions
}

deny[msg] if {
  input.action in destructive_actions
  not input.approval.status == "approved"
  msg := sprintf("action %s requires explicit human approval", [input.action])
}

deny[msg] if {
  input.mode == "auto-safe"
  input.action in destructive_actions
  msg := sprintf("action %s is blocked in auto-safe mode", [input.action])
}

deny[msg] if {
  input.image.scanned == false
  msg := "unscanned images cannot be deployed"
}
