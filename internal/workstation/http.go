package workstation

import (
	"errors"
	"net/http"

	"github.com/unboxd-cloud/platform/internal/server"
)

// Handler exposes cloud workstations over HTTP under /v1/workstations.
func Handler(m *Manager) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /v1/workstations", func(w http.ResponseWriter, r *http.Request) {
		var req LaunchRequest
		if !server.Decode(w, r, &req) {
			return
		}
		if req.Account == "" {
			req.Account = server.TenantID(r)
		}
		ws, err := m.Launch(r.Context(), req)
		if err != nil {
			writeErr(w, err)
			return
		}
		server.JSON(w, http.StatusAccepted, ws)
	})

	mux.HandleFunc("GET /v1/workstations", func(w http.ResponseWriter, r *http.Request) {
		account := r.URL.Query().Get("account")
		if account == "" {
			account = server.TenantID(r)
		}
		list, err := m.List(r.Context(), account)
		if err != nil {
			writeErr(w, err)
			return
		}
		server.JSON(w, http.StatusOK, list)
	})

	mux.HandleFunc("GET /v1/workstations/{id}", func(w http.ResponseWriter, r *http.Request) {
		ws, err := m.Get(r.Context(), r.PathValue("id"))
		if err != nil {
			writeErr(w, err)
			return
		}
		server.JSON(w, http.StatusOK, ws)
	})

	mux.HandleFunc("DELETE /v1/workstations/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if err := m.Stop(r.Context(), id); err != nil {
			writeErr(w, err)
			return
		}
		server.JSON(w, http.StatusAccepted, map[string]string{"id": id, "state": string(StateStopped)})
	})

	return mux
}

func writeErr(w http.ResponseWriter, err error) {
	if errors.Is(err, ErrNotFound) {
		server.Error(w, http.StatusNotFound, err.Error())
		return
	}
	server.Error(w, http.StatusBadRequest, err.Error())
}
