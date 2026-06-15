// Command compliance serves the compliance framework registry and evaluates
// placements against tenant compliance profiles (law-of-land, industry, and
// security frameworks). One service, one responsibility.
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/unboxd-cloud/platform/internal/api"
	"github.com/unboxd-cloud/platform/internal/compliance"
	"github.com/unboxd-cloud/platform/internal/server"
)

func main() {
	reg := compliance.NewRegistry()
	// Framework specs are loaded as a dataset at deployment time.
	if path := os.Getenv("COMPLIANCE_DATASET"); path != "" {
		f, err := os.Open(path)
		if err != nil {
			log.Fatalf("open compliance dataset: %v", err)
		}
		defer f.Close()
		n, err := reg.Load(f)
		if err != nil {
			log.Fatalf("load compliance dataset: %v", err)
		}
		log.Printf("loaded %d compliance frameworks from %s", n, path)
	} else {
		log.Printf("no COMPLIANCE_DATASET set; framework registry is empty")
	}

	mux := http.NewServeMux()

	// List frameworks (optionally by category).
	mux.HandleFunc("/v1/frameworks", func(w http.ResponseWriter, r *http.Request) {
		if c := r.URL.Query().Get("category"); c != "" {
			server.JSON(w, http.StatusOK, reg.Filter(compliance.Category(c)))
			return
		}
		server.JSON(w, http.StatusOK, reg.List())
	})

	// Evaluate a placement against a profile.
	mux.HandleFunc("/v1/evaluate", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			server.Error(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		var req api.EvalRequest
		if !server.Decode(w, r, &req) {
			return
		}
		if req.Profile.TenantID == "" {
			req.Profile.TenantID = server.TenantID(r)
		}
		rep := compliance.Evaluate(req.Profile, req.Placement, reg)
		status := http.StatusOK
		if !rep.Compliant {
			status = http.StatusUnprocessableEntity
		}
		server.JSON(w, status, rep)
	})

	addr := envOr("COMPLIANCE_ADDR", ":8084")
	log.Printf("compliance listening on %s", addr)
	log.Fatal(server.New(addr, mux).ListenAndServe())
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
