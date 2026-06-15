package observe

import (
	"errors"
	"testing"
)

func TestSpanLifecycleAndTrace(t *testing.T) {
	tr := New(10)
	root := tr.Start("", "", "chat", map[string]string{"msg": "catalog"})
	child := tr.Start(root.TraceID(), root.SpanID(), "sdk.catalog", nil)
	child.End(nil)
	root.End(nil)

	spans := tr.Trace(root.TraceID())
	if len(spans) != 2 {
		t.Fatalf("want 2 spans in trace, got %d", len(spans))
	}
	for _, s := range spans {
		if s.DurationMs < 0 {
			t.Errorf("negative duration: %v", s)
		}
		if s.Status != "ok" {
			t.Errorf("want ok status, got %s", s.Status)
		}
	}
}

func TestErrorStatusAndRingBuffer(t *testing.T) {
	tr := New(3)
	for i := 0; i < 5; i++ {
		s := tr.Start("", "", "op", nil)
		s.End(errors.New("boom"))
	}
	recent := tr.Recent(0)
	if len(recent) != 3 {
		t.Fatalf("ring buffer should cap at 3, got %d", len(recent))
	}
	if recent[0].Status != "error" {
		t.Errorf("want error status")
	}
}

func TestOTLPShape(t *testing.T) {
	tr := New(10)
	tr.Start("", "", "op", map[string]string{"k": "v"}).End(nil)
	o := tr.OTLP()
	if _, ok := o["resourceSpans"]; !ok {
		t.Fatal("OTLP output missing resourceSpans")
	}
}
