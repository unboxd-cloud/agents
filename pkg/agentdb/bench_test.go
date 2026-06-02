package agentdb

import (
	"fmt"
	"testing"
)

func BenchmarkMemStorePutGet(b *testing.B) {
	s := NewMemStore()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		id := fmt.Sprintf("r:%d", i%1000)
		_, _ = s.PutRecord(Record{ID: id, Kind: "agent"})
		_, _ = s.GetRecord(id)
	}
}

func BenchmarkKernelProposeApply(b *testing.B) {
	k := NewKernel(NewMemStore())
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p, _ := k.Propose(Proposal{Actor: "orchestrator", Op: OpCreate,
			Record: &Record{ID: fmt.Sprintf("agent:%d", i), Kind: "agent"}})
		_, _ = k.Apply(p.ID)
	}
}

func BenchmarkKernelGoverned(b *testing.B) {
	gate := PolicyFunc("gate", func(p Proposal, _ Store) Decision {
		if p.Record != nil && p.Record.Kind == "policy" {
			return Decision{Verdict: RequireApproval, Reason: "policy change"}
		}
		return Decision{Verdict: Allow}
	})
	k := NewKernel(NewMemStore(), gate)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = k.Propose(Proposal{Actor: "orchestrator", Op: OpCreate,
			Record: &Record{ID: fmt.Sprintf("agent:%d", i), Kind: "agent"}})
	}
}
