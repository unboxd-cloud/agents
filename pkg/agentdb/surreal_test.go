package agentdb

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// surrealMock returns a test server that captures the last SurrealQL body and
// replies with the given JSON envelope, plus a pointer to the captured query.
func surrealMock(t *testing.T, reply string) (*SurrealStore, *string, *http.Header) {
	t.Helper()
	var gotSQL string
	var gotHeader http.Header
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotSQL = string(b)
		gotHeader = r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, reply)
	}))
	t.Cleanup(srv.Close)
	st := NewSurrealStore(SurrealConfig{Endpoint: srv.URL, Namespace: "unboxd", Database: "agentdb", User: "root", Pass: "root"})
	return st, &gotSQL, &gotHeader
}

func TestSurrealPutRecordEmitsUpsert(t *testing.T) {
	st, sql, hdr := surrealMock(t, `[{"status":"OK","result":[]}]`)
	if _, err := st.PutRecord(Record{ID: "agent:1", Kind: "agent", Data: map[string]any{"name": "orch"}}); err != nil {
		t.Fatalf("PutRecord: %v", err)
	}
	if !strings.HasPrefix(*sql, "UPSERT record:`agent:1` CONTENT ") {
		t.Errorf("unexpected SurrealQL: %s", *sql)
	}
	if !strings.Contains(*sql, `"kind":"agent"`) || !strings.Contains(*sql, `"name":"orch"`) {
		t.Errorf("content missing fields: %s", *sql)
	}
	if hdr.Get("NS") != "unboxd" || hdr.Get("DB") != "agentdb" {
		t.Errorf("NS/DB headers not set: %v", *hdr)
	}
	if hdr.Get("Authorization") == "" {
		t.Error("basic auth not set")
	}
}

func TestSurrealGetRecordParses(t *testing.T) {
	reply := `[{"status":"OK","result":[{"recordId":"agent:1","kind":"agent","data":{"name":"orch"},"createdAt":"2026-06-02T00:00:00Z","updatedAt":"2026-06-02T00:00:00Z"}]}]`
	st, sql, _ := surrealMock(t, reply)
	got, ok := st.GetRecord("agent:1")
	if !ok || got.Kind != "agent" || got.Data["name"] != "orch" {
		t.Fatalf("GetRecord = %+v ok=%v", got, ok)
	}
	if !strings.HasPrefix(*sql, "SELECT * FROM record:`agent:1`") {
		t.Errorf("unexpected SurrealQL: %s", *sql)
	}
}

func TestSurrealListRecordsFiltersByKind(t *testing.T) {
	st, sql, _ := surrealMock(t, `[{"status":"OK","result":[]}]`)
	st.ListRecords("agent")
	if !strings.Contains(*sql, `SELECT * FROM record WHERE kind = "agent"`) {
		t.Errorf("unexpected SurrealQL: %s", *sql)
	}
}

func TestSurrealEdges(t *testing.T) {
	st, sql, _ := surrealMock(t, `[{"status":"OK","result":[{"edgeId":"e1","kind":"supervises","from":"agent:1","to":"agent:2"}]}]`)
	if _, err := st.PutEdge(Edge{ID: "e1", Kind: "supervises", From: "agent:1", To: "agent:2"}); err != nil {
		t.Fatalf("PutEdge: %v", err)
	}
	got := st.Edges("agent:1")
	if len(got) != 1 || got[0].Kind != "supervises" {
		t.Fatalf("Edges = %+v", got)
	}
	if !strings.Contains(*sql, `SELECT * FROM edge WHERE "from" = "agent:1"`) {
		t.Errorf("unexpected SurrealQL: %s", *sql)
	}
}

func TestSurrealReportsStatusError(t *testing.T) {
	st, _, _ := surrealMock(t, `[{"status":"ERR","detail":"boom"}]`)
	if _, err := st.PutRecord(Record{ID: "x", Kind: "k"}); err == nil {
		t.Error("expected error from ERR status")
	}
}
