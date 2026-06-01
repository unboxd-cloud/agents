// Command catalog serves the service catalog of CNCF/AI offerings.
// One service, one responsibility: read the catalog (optionally per persona).
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/unboxd-cloud/platform/internal/catalog"
	"github.com/unboxd-cloud/platform/internal/server"
)

func main() {
	store := catalog.Seeded()
	mux := http.NewServeMux()

	mux.HandleFunc("/v1/catalog", func(w http.ResponseWriter, r *http.Request) {
		if profile := r.URL.Query().Get("profile"); profile != "" {
			server.JSON(w, http.StatusOK, store.ForProfile(profile))
			return
		}
		server.JSON(w, http.StatusOK, store.List())
	})

	addr := envOr("CATALOG_ADDR", ":8083")
	log.Printf("catalog listening on %s", addr)
	log.Fatal(server.New(addr, mux).ListenAndServe())
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
