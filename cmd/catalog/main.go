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
	store := loadCatalog(os.Getenv("CATALOG_DATASET"))
	mux := http.NewServeMux()

	mux.HandleFunc("/v1/catalog", func(w http.ResponseWriter, r *http.Request) {
		if profile := r.URL.Query().Get("profile"); profile != "" {
			server.JSON(w, http.StatusOK, store.ForProfile(profile))
			return
		}
		if category := r.URL.Query().Get("category"); category != "" {
			server.JSON(w, http.StatusOK, store.ForCategory(category))
			return
		}
		server.JSON(w, http.StatusOK, store.List())
	})

	// Category-wise registry index (the composable full-stack catalog).
	mux.HandleFunc("/v1/categories", func(w http.ResponseWriter, _ *http.Request) {
		server.JSON(w, http.StatusOK, store.Categories())
	})

	addr := envOr("CATALOG_ADDR", ":8083")
	log.Printf("catalog listening on %s", addr)
	log.Fatal(server.New(addr, mux).ListenAndServe())
}

func loadCatalog(path string) *catalog.MemStore {
	if path == "" {
		return catalog.Seeded() // built-in dev default
	}
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("open catalog dataset: %v", err)
	}
	defer f.Close()
	store, err := catalog.Load(f)
	if err != nil {
		log.Fatalf("load catalog dataset: %v", err)
	}
	log.Printf("loaded %d catalog offerings from %s", len(store.List()), path)
	return store
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
