package sdk

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/unboxd-cloud/platform/internal/api"
	"github.com/unboxd-cloud/platform/internal/metering"
)

func TestClient_RateAndUsage(t *testing.T) {
	var gotTenant string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotTenant = r.Header.Get("X-Tenant-ID")
		switch r.URL.Path {
		case "/v1/usage":
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{}`))
		case "/v1/rate":
			_, _ = w.Write([]byte(`{"invoice":{"tenantId":"t1","total":42,"currency":"USD"}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := New()
	c.Metering, c.Billing = srv.URL, srv.URL
	c.Tenant = "t1"

	if err := c.RecordUsage(context.Background(), metering.UsageEvent{TenantID: "t1", Meter: "m", Quantity: 1}); err != nil {
		t.Fatal(err)
	}
	resp, err := c.Rate(context.Background(), api.RateRequest{TenantID: "t1"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Invoice.Total != 42 {
		t.Errorf("want total 42, got %v", resp.Invoice.Total)
	}
	if gotTenant != "t1" {
		t.Errorf("tenant header not sent: %q", gotTenant)
	}
}
