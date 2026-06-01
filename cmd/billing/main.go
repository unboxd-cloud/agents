// Command billing rates usage and serves invoices/settlements.
// One service, one responsibility: turn usage + price book into money.
package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/unboxd-cloud/platform/internal/billing"
	"github.com/unboxd-cloud/platform/internal/metering"
	"github.com/unboxd-cloud/platform/internal/server"
)

// rateRequest is the stateless rating input: usage + an optional partner
// overlay. PriceBook defaults to the sample book when omitted.
type rateRequest struct {
	TenantID  string                `json:"tenantId"`
	From      time.Time             `json:"from"`
	To        time.Time             `json:"to"`
	Events    []metering.UsageEvent `json:"events"`
	PriceBook *billing.PriceBook    `json:"priceBook,omitempty"`
	Partner   *billing.Partner      `json:"partner,omitempty"`
}

func main() {
	mux := http.NewServeMux()

	// Expose the active price book for transparency.
	mux.HandleFunc("/v1/pricebook", func(w http.ResponseWriter, _ *http.Request) {
		server.JSON(w, http.StatusOK, billing.SamplePriceBook())
	})

	// Rate usage into an invoice; if a partner is supplied, also settle it.
	mux.HandleFunc("/v1/rate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			server.Error(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		var req rateRequest
		if !server.Decode(w, r, &req) {
			return
		}
		if req.TenantID == "" {
			req.TenantID = server.TenantID(r)
		}
		if req.TenantID == "" {
			server.Error(w, http.StatusBadRequest, "tenantId required")
			return
		}
		pb := billing.SamplePriceBook()
		if req.PriceBook != nil {
			pb = *req.PriceBook
		}
		inv := billing.Rate(pb, req.TenantID, req.From, req.To, req.Events)
		if req.Partner != nil {
			server.JSON(w, http.StatusOK, billing.Settle(inv, *req.Partner))
			return
		}
		server.JSON(w, http.StatusOK, inv)
	})

	addr := envOr("BILLING_ADDR", ":8082")
	log.Printf("billing listening on %s", addr)
	log.Fatal(server.New(addr, mux).ListenAndServe())
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
