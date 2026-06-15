package agent

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

type countingAgent struct {
	n   atomic.Int64
	err error
}

func (c *countingAgent) Name() string { return "counting" }
func (c *countingAgent) Reconcile(context.Context) error {
	c.n.Add(1)
	return c.err
}

func TestRun_ReconcilesUntilCancel(t *testing.T) {
	a := &countingAgent{err: errors.New("transient")} // errors must not stop the loop
	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Millisecond)
	defer cancel()
	_ = Run(ctx, a, 10*time.Millisecond)
	if a.n.Load() < 2 {
		t.Errorf("expected multiple reconciles, got %d", a.n.Load())
	}
}

func TestOperator_SupervisesMultiple(t *testing.T) {
	a, b := &countingAgent{}, &countingAgent{}
	ctx, cancel := context.WithTimeout(context.Background(), 35*time.Millisecond)
	defer cancel()
	_ = Operator(ctx,
		Scheduled{Agent: a, Interval: 10 * time.Millisecond},
		Scheduled{Agent: b, Interval: 10 * time.Millisecond},
	)
	if a.n.Load() == 0 || b.n.Load() == 0 {
		t.Errorf("both agents should have reconciled: a=%d b=%d", a.n.Load(), b.n.Load())
	}
}
