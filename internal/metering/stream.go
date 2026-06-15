package metering

import "context"

// StreamSource is a push-based usage source (e.g. CloudEvents over NATS, or an
// OpenTelemetry collector). It streams events on a channel until the context is
// cancelled — the streaming counterpart to the pull-based Source.
type StreamSource interface {
	Name() string
	Stream(ctx context.Context) (<-chan UsageEvent, error)
}

// Drain records events streamed from a StreamSource until the channel closes or
// the context is cancelled. This is the composition point for real-time meters.
func Drain(ctx context.Context, src StreamSource, store Store) (int, error) {
	ch, err := src.Stream(ctx)
	if err != nil {
		return 0, err
	}
	n := 0
	for {
		select {
		case <-ctx.Done():
			return n, ctx.Err()
		case e, ok := <-ch:
			if !ok {
				return n, nil
			}
			if err := store.Record(e); err != nil {
				return n, err
			}
			n++
		}
	}
}
