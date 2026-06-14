package controlplane

import (
	"errors"
	"net/http"

	"github.com/unboxd-cloud/platform/internal/cloudstack"
	"github.com/unboxd-cloud/platform/internal/server"
)

// Handler returns the control plane's HTTP API: a clean REST surface under /v1
// plus a CloudStack-compatible /client/api endpoint, so both modern clients and
// existing CloudStack tooling (which calls ?command=deployVirtualMachine&...)
// interoperate against the same control plane.
func Handler(cp *ControlPlane) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /v1/vms", func(w http.ResponseWriter, r *http.Request) {
		var req cloudstack.DeployVMRequest
		if !server.Decode(w, r, &req) {
			return
		}
		if req.Account == "" {
			req.Account = server.TenantID(r)
		}
		vm, err := cp.DeployVirtualMachine(r.Context(), req)
		if err != nil {
			writeErr(w, err)
			return
		}
		server.JSON(w, http.StatusAccepted, vm)
	})

	mux.HandleFunc("GET /v1/vms", func(w http.ResponseWriter, r *http.Request) {
		account := r.URL.Query().Get("account")
		if account == "" {
			account = server.TenantID(r)
		}
		vms, err := cp.ListVirtualMachines(r.Context(), account)
		if err != nil {
			writeErr(w, err)
			return
		}
		server.JSON(w, http.StatusOK, vms)
	})

	mux.HandleFunc("GET /v1/vms/{id}", func(w http.ResponseWriter, r *http.Request) {
		vm, err := cp.GetVirtualMachine(r.Context(), r.PathValue("id"))
		if err != nil {
			writeErr(w, err)
			return
		}
		server.JSON(w, http.StatusOK, vm)
	})

	mux.HandleFunc("POST /v1/vms/{id}/start", func(w http.ResponseWriter, r *http.Request) {
		vm, err := cp.StartVirtualMachine(r.Context(), r.PathValue("id"))
		if err != nil {
			writeErr(w, err)
			return
		}
		server.JSON(w, http.StatusAccepted, vm)
	})

	mux.HandleFunc("POST /v1/vms/{id}/stop", func(w http.ResponseWriter, r *http.Request) {
		vm, err := cp.StopVirtualMachine(r.Context(), r.PathValue("id"))
		if err != nil {
			writeErr(w, err)
			return
		}
		server.JSON(w, http.StatusAccepted, vm)
	})

	mux.HandleFunc("DELETE /v1/vms/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if err := cp.DestroyVirtualMachine(r.Context(), id); err != nil {
			writeErr(w, err)
			return
		}
		server.JSON(w, http.StatusAccepted, map[string]string{"id": id, "state": string(cloudstack.StateDestroyed)})
	})

	mux.HandleFunc("GET /v1/catalog", func(w http.ResponseWriter, _ *http.Request) {
		server.JSON(w, http.StatusOK, map[string]any{
			"zones":            cp.Zones(),
			"serviceofferings": cp.ServiceOfferings(),
			"templates":        cp.Templates(),
		})
	})

	// Apache CloudStack-compatible query API: GET/POST /client/api?command=...
	mux.HandleFunc("/client/api", func(w http.ResponseWriter, r *http.Request) {
		clientAPI(cp, w, r)
	})

	return mux
}

// clientAPI implements the subset of the Apache CloudStack query API the control
// plane supports, so existing CloudStack clients work unchanged. Responses use
// CloudStack's JSON envelope, keyed by "<command>response".
func clientAPI(cp *ControlPlane, w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	ctx := r.Context()
	switch q.Get("command") {
	case "listZones":
		respond(w, "listzones", map[string]any{"zone": cp.Zones()})
	case "listServiceOfferings":
		respond(w, "listserviceofferings", map[string]any{"serviceoffering": cp.ServiceOfferings()})
	case "listTemplates":
		respond(w, "listtemplates", map[string]any{"template": cp.Templates()})
	case "listVirtualMachines":
		vms, err := cp.ListVirtualMachines(ctx, q.Get("account"))
		if err != nil {
			writeErr(w, err)
			return
		}
		respond(w, "listvirtualmachines", map[string]any{"virtualmachine": vms})
	case "deployVirtualMachine":
		vm, err := cp.DeployVirtualMachine(ctx, cloudstack.DeployVMRequest{
			Account:           q.Get("account"),
			Name:              q.Get("name"),
			DisplayName:       q.Get("displayname"),
			ZoneID:            q.Get("zoneid"),
			TemplateID:        q.Get("templateid"),
			ServiceOfferingID: q.Get("serviceofferingid"),
		})
		if err != nil {
			writeErr(w, err)
			return
		}
		respond(w, "deployvirtualmachine", map[string]any{"virtualmachine": vm})
	case "startVirtualMachine":
		vm, err := cp.StartVirtualMachine(ctx, q.Get("id"))
		if err != nil {
			writeErr(w, err)
			return
		}
		respond(w, "startvirtualmachine", map[string]any{"virtualmachine": vm})
	case "stopVirtualMachine":
		vm, err := cp.StopVirtualMachine(ctx, q.Get("id"))
		if err != nil {
			writeErr(w, err)
			return
		}
		respond(w, "stopvirtualmachine", map[string]any{"virtualmachine": vm})
	case "destroyVirtualMachine":
		if err := cp.DestroyVirtualMachine(ctx, q.Get("id")); err != nil {
			writeErr(w, err)
			return
		}
		respond(w, "destroyvirtualmachine", map[string]any{"success": true})
	case "":
		server.Error(w, http.StatusBadRequest, "missing command parameter")
	default:
		server.Error(w, http.StatusBadRequest, "unsupported command: "+q.Get("command"))
	}
}

func respond(w http.ResponseWriter, command string, payload map[string]any) {
	server.JSON(w, http.StatusOK, map[string]any{command + "response": payload})
}

func writeErr(w http.ResponseWriter, err error) {
	if errors.Is(err, cloudstack.ErrNotFound) {
		server.Error(w, http.StatusNotFound, err.Error())
		return
	}
	server.Error(w, http.StatusBadRequest, err.Error())
}
