package agentdb

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// surrealResult is one statement's result in a SurrealDB /sql response envelope.
type surrealResult struct {
	Status string          `json:"status"`
	Detail string          `json:"detail,omitempty"`
	Result json.RawMessage `json:"result,omitempty"`
}

// decodeEnvelope parses the JSON array SurrealDB returns from /sql.
func decodeEnvelope(body []byte) ([]surrealResult, error) {
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		return nil, nil
	}
	var res []surrealResult
	if err := json.Unmarshal([]byte(trimmed), &res); err != nil {
		return nil, fmt.Errorf("surrealdb: decode response: %w", err)
	}
	return res, nil
}

// firstErr returns the first non-OK statement status as an error.
func firstErr(res []surrealResult) error {
	for _, r := range res {
		if r.Status != "" && r.Status != "OK" {
			msg := r.Detail
			if msg == "" {
				msg = r.Status
			}
			return fmt.Errorf("surrealdb: %s", msg)
		}
	}
	return nil
}

// thing builds a SurrealDB record identifier `table:`id“ with the id safely
// backtick-quoted (backticks in the id are stripped to avoid injection).
func thing(table, id string) string {
	return table + ":`" + strings.ReplaceAll(id, "`", "") + "`"
}

// mustJSON marshals v to a compact JSON literal for embedding in SurrealQL
// (SurrealQL object/string literals are JSON-compatible).
func mustJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}

type recDoc struct {
	RecordID  string         `json:"recordId"`
	Kind      string         `json:"kind"`
	Data      map[string]any `json:"data"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
}

func recordDoc(r Record) map[string]any {
	return map[string]any{
		"recordId":  r.ID,
		"kind":      r.Kind,
		"data":      r.Data,
		"createdAt": r.CreatedAt,
		"updatedAt": r.UpdatedAt,
	}
}

func recordsFrom(res []surrealResult) []Record {
	var out []Record
	for _, r := range res {
		if len(r.Result) == 0 {
			continue
		}
		var docs []recDoc
		if err := json.Unmarshal(r.Result, &docs); err != nil {
			continue
		}
		for _, d := range docs {
			out = append(out, Record{ID: d.RecordID, Kind: d.Kind, Data: d.Data, CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt})
		}
	}
	return out
}

type edgeRow struct {
	EdgeID string         `json:"edgeId"`
	Kind   string         `json:"kind"`
	From   string         `json:"from"`
	To     string         `json:"to"`
	Data   map[string]any `json:"data"`
}

func edgeDoc(e Edge) map[string]any {
	return map[string]any{
		"edgeId": e.ID,
		"kind":   e.Kind,
		"from":   e.From,
		"to":     e.To,
		"data":   e.Data,
	}
}

func edgesFrom(res []surrealResult) []Edge {
	var out []Edge
	for _, r := range res {
		if len(r.Result) == 0 {
			continue
		}
		var rows []edgeRow
		if err := json.Unmarshal(r.Result, &rows); err != nil {
			continue
		}
		for _, d := range rows {
			out = append(out, Edge{ID: d.EdgeID, Kind: d.Kind, From: d.From, To: d.To, Data: d.Data})
		}
	}
	return out
}
