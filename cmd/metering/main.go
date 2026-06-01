// Command metering ingests and serves usage events.
// One service, one responsibility: record and query usage.
package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/unboxd-cloud/platform/internal/metering"
	"github.com/unboxd-cloud/platform/internal/server"
)

func main() {
	store := metering.NewMemStore()
	mux := http.NewServeMux()

	mux.HandleFunc("/v1/usage", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			var e metering.UsageEvent
			if !server.Decode(w, r, &e) {
				return
			}
			if e.TenantID == "" {
				e.TenantID = server.TenantID(r)
			}
			if err := store.Record(e); err != nil {
				server.Error(w, http.StatusBadRequest, err.Error())
				return
			}
			server.JSON(w, http.StatusCreated, e)
		case http.MethodGet:
			tenant := server.TenantID(r)
			if tenant == "" {
				tenant = r.URL.Query().Get("tenant")
			}
			if tenant == "" {
				server.Error(w, http.StatusBadRequest, "tenant required")
				return
			}
			server.JSON(w, http.StatusOK, store.Query(tenant, time.Time{}, time.Time{}))
		default:
			server.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	})

	addr := envOr("METERING_ADDR", ":8081")
	log.Printf("metering listening on %s", addr)
	log.Fatal(server.New(addr, mux).ListenAndServe())
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
