// Command billing rates usage and serves invoices/settlements.
// One service, one responsibility: turn usage + price book into money.
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/unboxd-cloud/platform/internal/api"
	"github.com/unboxd-cloud/platform/internal/billing"
	"github.com/unboxd-cloud/platform/internal/server"
)

func main() {
	// Datasets (price book, tax table) are loaded at deployment time, falling
	// back to built-in dev defaults when no dataset path is configured.
	priceBook := loadPriceBook(os.Getenv("PRICEBOOK_DATASET"))
	taxTable := loadTaxTable(os.Getenv("TAX_DATASET"))

	mux := http.NewServeMux()

	// Expose the active price book for transparency.
	mux.HandleFunc("/v1/pricebook", func(w http.ResponseWriter, _ *http.Request) {
		server.JSON(w, http.StatusOK, priceBook)
	})

	// Rate usage -> invoice; optionally settle a partner overlay; optionally tax
	// the customer-facing amount by jurisdiction.
	mux.HandleFunc("/v1/rate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			server.Error(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		var req api.RateRequest
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
		pb := priceBook
		if req.PriceBook != nil {
			pb = *req.PriceBook
		}

		inv := billing.Rate(pb, req.TenantID, req.From, req.To, req.Events)
		resp := api.RateResponse{Invoice: inv}

		// Customer-facing amount: gross after partner settlement, else invoice total.
		customerAmount := inv.Total
		if req.Partner != nil {
			s := billing.Settle(inv, *req.Partner)
			resp.Settlement = &s
			customerAmount = s.GrossToCustomer
		}

		if req.Jurisdiction != "" {
			tax := billing.ApplyTax(customerAmount, inv.Currency, taxTable.For(req.Jurisdiction))
			resp.Tax = &tax
		}

		server.JSON(w, http.StatusOK, resp)
	})

	addr := envOr("BILLING_ADDR", ":8082")
	log.Printf("billing listening on %s", addr)
	log.Fatal(server.New(addr, mux).ListenAndServe())
}

func loadPriceBook(path string) billing.PriceBook {
	if path == "" {
		return billing.SamplePriceBook()
	}
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("open price book dataset: %v", err)
	}
	defer f.Close()
	pb, err := billing.LoadPriceBook(f)
	if err != nil {
		log.Fatalf("load price book dataset: %v", err)
	}
	log.Printf("loaded price book %s from %s", pb.Version, path)
	return pb
}

func loadTaxTable(path string) billing.TaxTable {
	if path == "" {
		return nil // nil table falls back to built-in TaxRulesFor
	}
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("open tax dataset: %v", err)
	}
	defer f.Close()
	t, err := billing.LoadTaxTable(f)
	if err != nil {
		log.Fatalf("load tax dataset: %v", err)
	}
	log.Printf("loaded tax table (%d jurisdictions) from %s", len(t), path)
	return t
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
