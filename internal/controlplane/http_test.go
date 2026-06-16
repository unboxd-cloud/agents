package controlplane

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/unboxd-cloud/platform/internal/cloudstack"
	"github.com/unboxd-cloud/platform/internal/kube"
)

func TestHandler_DeployThenListViaCloudStackAPI(t *testing.T) {
	cp := New(kube.NewManager())
	h := Handler(cp)

	body, _ := json.Marshal(cloudstack.DeployVMRequest{
		Account: "t1", Name: "web-1", ZoneID: "zone-1",
		TemplateID: "tmpl-nginx", ServiceOfferingID: "so-small",
	})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/v1/vms", bytes.NewReader(body)))
	if rec.Code != http.StatusAccepted {
		t.Fatalf("deploy status = %d, body=%s", rec.Code, rec.Body)
	}
	var vm cloudstack.VirtualMachine
	if err := json.Unmarshal(rec.Body.Bytes(), &vm); err != nil {
		t.Fatal(err)
	}
	if vm.ID == "" || vm.State != cloudstack.StateStarting {
		t.Fatalf("unexpected vm: %+v", vm)
	}

	// Drive reconciliation, then read back via the CloudStack-compatible API.
	if err := cp.Reconcile(context.Background()); err != nil {
		t.Fatal(err)
	}
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/client/api?command=listVirtualMachines&account=t1", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("listVirtualMachines status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "listvirtualmachinesresponse") {
		t.Fatalf("missing CloudStack response envelope: %s", rec.Body)
	}
	if !strings.Contains(rec.Body.String(), string(cloudstack.StateRunning)) {
		t.Fatalf("vm not Running after reconcile: %s", rec.Body)
	}
}

func TestHandler_DeployViaCloudStackCommand(t *testing.T) {
	cp := New(kube.NewManager())
	h := Handler(cp)
	rec := httptest.NewRecorder()
	url := "/client/api?command=deployVirtualMachine&account=t1&name=db-1&zoneid=zone-1&templateid=tmpl-alpine&serviceofferingid=so-medium"
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, url, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("deploy via command status = %d, body=%s", rec.Code, rec.Body)
	}
	if !strings.Contains(rec.Body.String(), "deployvirtualmachineresponse") {
		t.Fatalf("missing response envelope: %s", rec.Body)
	}
}

func TestHandler_NotFoundAndBadCommand(t *testing.T) {
	cp := New(kube.NewManager())
	h := Handler(cp)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v1/vms/nope", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("want 404 for unknown vm, got %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/client/api?command=bogus", nil))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400 for unsupported command, got %d", rec.Code)
	}
}

func TestHandler_DirectDelivery(t *testing.T) {
	cp := New(kube.NewManager())
	h := Handler(cp)

	body, _ := json.Marshal(cloudstack.DeployVMRequest{
		Account: "t1", Name: "web-1", ZoneID: "zone-1",
		TemplateID: "tmpl-nginx", ServiceOfferingID: "so-small",
	})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/v1/vms?wait=true", bytes.NewReader(body)))
	if rec.Code != http.StatusOK {
		t.Fatalf("direct delivery status = %d, body=%s", rec.Code, rec.Body)
	}
	var vm cloudstack.VirtualMachine
	if err := json.Unmarshal(rec.Body.Bytes(), &vm); err != nil {
		t.Fatal(err)
	}
	if vm.State != cloudstack.StateRunning {
		t.Fatalf("want Running from direct delivery, got %s", vm.State)
	}
}
