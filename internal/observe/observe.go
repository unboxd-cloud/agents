// Package observe is a lightweight, dependency-free tracing layer. It records
// spans (operation, attributes, status, duration) into a ring buffer for
// in-UI flow debugging, and exports them as OTLP-style JSON so traces can be
// shipped to CNCF tools (Jaeger, Grafana Tempo, Prometheus via the
// OpenTelemetry Collector) for analysis and insights.
package observe

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

// Span is one timed operation within a trace.
type Span struct {
	TraceID    string            `json:"traceId"`
	SpanID     string            `json:"spanId"`
	Parent     string            `json:"parentSpanId,omitempty"`
	Name       string            `json:"name"`
	Start      time.Time         `json:"start"`
	End        time.Time         `json:"end"`
	DurationMs float64           `json:"durationMs"`
	Attributes map[string]string `json:"attributes,omitempty"`
	Status     string            `json:"status"` // "ok" | "error"
	Error      string            `json:"error,omitempty"`
}

// Tracer records spans in a bounded ring buffer.
type Tracer struct {
	mu    sync.Mutex
	max   int
	spans []Span
}

// New returns a Tracer that retains up to max spans.
func New(max int) *Tracer {
	if max <= 0 {
		max = 1000
	}
	return &Tracer{max: max}
}

// Active is a span being timed; call End to record it.
type Active struct {
	t    *Tracer
	span Span
}

// Start begins a span. Pass an empty traceID to start a new trace.
func (t *Tracer) Start(traceID, parent, name string, attrs map[string]string) *Active {
	if traceID == "" {
		traceID = randHex(16)
	}
	return &Active{
		t: t,
		span: Span{
			TraceID:    traceID,
			SpanID:     randHex(8),
			Parent:     parent,
			Name:       name,
			Start:      time.Now(),
			Attributes: attrs,
		},
	}
}

// TraceID exposes the active span's trace id (for child spans / linking).
func (a *Active) TraceID() string { return a.span.TraceID }

// SpanID exposes the active span's id.
func (a *Active) SpanID() string { return a.span.SpanID }

// End finalizes the span (status from err) and records it.
func (a *Active) End(err error) {
	a.span.End = time.Now()
	a.span.DurationMs = float64(a.span.End.Sub(a.span.Start).Microseconds()) / 1000.0
	a.span.Status = "ok"
	if err != nil {
		a.span.Status = "error"
		a.span.Error = err.Error()
	}
	a.t.record(a.span)
}

func (t *Tracer) record(s Span) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.spans = append(t.spans, s)
	if len(t.spans) > t.max {
		t.spans = t.spans[len(t.spans)-t.max:]
	}
}

// Recent returns up to n most-recent spans (newest last). n<=0 returns all.
func (t *Tracer) Recent(n int) []Span {
	t.mu.Lock()
	defer t.mu.Unlock()
	if n <= 0 || n > len(t.spans) {
		n = len(t.spans)
	}
	out := make([]Span, n)
	copy(out, t.spans[len(t.spans)-n:])
	return out
}

// Trace returns all spans belonging to a trace id, in start order.
func (t *Tracer) Trace(traceID string) []Span {
	t.mu.Lock()
	defer t.mu.Unlock()
	var out []Span
	for _, s := range t.spans {
		if s.TraceID == traceID {
			out = append(out, s)
		}
	}
	return out
}

// OTLP renders recent spans in an OpenTelemetry OTLP/JSON-shaped structure,
// ready to POST to an OTel Collector for export to Jaeger/Tempo/etc.
func (t *Tracer) OTLP() map[string]any {
	spans := t.Recent(0)
	otlpSpans := make([]map[string]any, 0, len(spans))
	for _, s := range spans {
		attrs := make([]map[string]any, 0, len(s.Attributes))
		for k, v := range s.Attributes {
			attrs = append(attrs, map[string]any{
				"key":   k,
				"value": map[string]any{"stringValue": v},
			})
		}
		otlpSpans = append(otlpSpans, map[string]any{
			"traceId":           s.TraceID,
			"spanId":            s.SpanID,
			"parentSpanId":      s.Parent,
			"name":              s.Name,
			"startTimeUnixNano": s.Start.UnixNano(),
			"endTimeUnixNano":   s.End.UnixNano(),
			"attributes":        attrs,
			"status":            map[string]any{"code": statusCode(s.Status)},
		})
	}
	return map[string]any{
		"resourceSpans": []map[string]any{{
			"resource": map[string]any{
				"attributes": []map[string]any{{
					"key":   "service.name",
					"value": map[string]any{"stringValue": "unboxd-platform"},
				}},
			},
			"scopeSpans": []map[string]any{{"spans": otlpSpans}},
		}},
	}
}

func statusCode(s string) int {
	if s == "error" {
		return 2 // STATUS_CODE_ERROR
	}
	return 1 // STATUS_CODE_OK
}

func randHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		// fall back to time-based; ids are for correlation, not security
		for i := range b {
			b[i] = byte(time.Now().UnixNano() >> (i % 8))
		}
	}
	return hex.EncodeToString(b)
}
