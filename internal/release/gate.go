// Package release enforces the production gate that every publish/deploy must
// pass. The platform treats everything as production, so the gate always
// applies — there is no ungated path. It composes existing primitives:
//   - dependency gate  : depgraph resolves (no cycle) and has no unmet deps,
//   - human approval   : the workflow is Approved and a human signed off,
//   - accountability   : every approved stage records who decided and when.
package release

import (
	"fmt"

	"github.com/unboxd-cloud/platform/internal/depgraph"
	"github.com/unboxd-cloud/platform/internal/workflow"
)

// Result is the gate outcome. OK is true only if all gates pass.
type Result struct {
	OK      bool     `json:"ok"`
	Reasons []string `json:"reasons,omitempty"` // why it failed (empty when OK)
}

// Gate checks production readiness for a target before publish/deploy.
type Gate struct {
	// Deps is the dependency graph for the deployment.
	Deps *depgraph.Graph
	// Known marks installable nodes; deps outside it are unmet.
	Known map[string]bool
}

// Check runs all three gates for target, gated by its workflow.
func (g Gate) Check(target string, wf *workflow.Workflow) Result {
	var reasons []string

	// 1) Dependency gate.
	if g.Deps == nil {
		reasons = append(reasons, "dependency gate: no dependency graph provided")
	} else {
		if _, err := g.Deps.Resolve(); err != nil {
			reasons = append(reasons, "dependency gate: "+err.Error())
		}
		if unmet := g.Deps.Missing(g.Known); len(unmet) > 0 {
			reasons = append(reasons, fmt.Sprintf("dependency gate: unmet dependencies %v", unmet))
		}
	}

	// 2) Human approval gate + 3) accountability gate (from the workflow).
	if wf == nil {
		reasons = append(reasons, "approval gate: no workflow")
	} else {
		if wf.Decision != workflow.Approved {
			reasons = append(reasons, fmt.Sprintf("approval gate: workflow not approved (%s)", wf.Decision))
		}
		humanSignoff := false
		for _, st := range wf.Stages {
			if st.Decision == workflow.Approved {
				// accountability: who + when must be recorded.
				if st.By == "" || st.At.IsZero() {
					reasons = append(reasons, fmt.Sprintf("accountability gate: stage %q missing approver/time", st.Name))
				}
				if st.Approver.Type == workflow.ApproverHuman && st.By != "" {
					humanSignoff = true
				}
			}
		}
		if !humanSignoff {
			reasons = append(reasons, "approval gate: a human sign-off is required for production")
		}
	}

	return Result{OK: len(reasons) == 0, Reasons: reasons}
}
