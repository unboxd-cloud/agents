// Package cloudstack defines the platform's northbound IaaS contract, modeled on
// the Apache CloudStack API. CloudStack is used as the *contract* — the request
// and response shapes and the VM lifecycle that clients program against — so
// existing CloudStack tooling interoperates unchanged. The contract says nothing
// about how compute is realized: the platform implements it CNCF-natively by
// reconciling onto Kubernetes (see internal/controlplane), keeping CloudStack
// the API while Kubernetes does the work. This complements ADR-0004 (don't run
// the heavyweight CloudStack server) and is recorded in ADR-0007.
package cloudstack

import (
	"context"
	"errors"
	"strings"
	"time"
)

// ErrNotFound is returned when a referenced resource does not exist.
var ErrNotFound = errors.New("cloudstack: resource not found")

// VMState mirrors the lifecycle states Apache CloudStack reports for a
// VirtualMachine, so clients observe the states they already expect.
type VMState string

const (
	StateStarting  VMState = "Starting"
	StateRunning   VMState = "Running"
	StateStopping  VMState = "Stopping"
	StateStopped   VMState = "Stopped"
	StateDestroyed VMState = "Destroyed"
	StateError     VMState = "Error"
)

// Zone is an availability zone a VM can be deployed into.
type Zone struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ServiceOffering is a compute flavor (CloudStack "service offering"): the CPU
// and memory a VM is allocated.
type ServiceOffering struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CPUNumber int    `json:"cpunumber"`
	Memory    int    `json:"memory"` // MB
}

// Template is a bootable image (CloudStack "template"). Image is the OCI image
// the template maps to when realized on Kubernetes.
type Template struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	OSType string `json:"ostypename"`
	Image  string `json:"-"`
}

// VirtualMachine mirrors CloudStack's VirtualMachine response object.
type VirtualMachine struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	DisplayName       string    `json:"displayname,omitempty"`
	Account           string    `json:"account"` // tenant (CloudStack "account")
	ZoneID            string    `json:"zoneid"`
	TemplateID        string    `json:"templateid"`
	ServiceOfferingID string    `json:"serviceofferingid"`
	State             VMState   `json:"state"`
	Created           time.Time `json:"created"`
}

// DeployVMRequest mirrors the parameters of CloudStack's deployVirtualMachine
// command.
type DeployVMRequest struct {
	Account           string `json:"account"`
	Name              string `json:"name"`
	DisplayName       string `json:"displayname,omitempty"`
	ZoneID            string `json:"zoneid"`
	TemplateID        string `json:"templateid"`
	ServiceOfferingID string `json:"serviceofferingid"`
}

// Validate reports whether the request carries the fields required to deploy.
func (r DeployVMRequest) Validate() error {
	switch {
	case strings.TrimSpace(r.Account) == "":
		return errors.New("cloudstack: account is required")
	case strings.TrimSpace(r.Name) == "":
		return errors.New("cloudstack: name is required")
	case strings.TrimSpace(r.ZoneID) == "":
		return errors.New("cloudstack: zoneid is required")
	case strings.TrimSpace(r.TemplateID) == "":
		return errors.New("cloudstack: templateid is required")
	case strings.TrimSpace(r.ServiceOfferingID) == "":
		return errors.New("cloudstack: serviceofferingid is required")
	}
	return nil
}

// Contract is the northbound IaaS API the cloud control plane exposes, modeled
// on the Apache CloudStack API. Any implementation may realize it on any
// substrate; the platform's implementation reconciles it onto Kubernetes pods.
type Contract interface {
	DeployVirtualMachine(ctx context.Context, req DeployVMRequest) (VirtualMachine, error)
	StartVirtualMachine(ctx context.Context, id string) (VirtualMachine, error)
	StopVirtualMachine(ctx context.Context, id string) (VirtualMachine, error)
	DestroyVirtualMachine(ctx context.Context, id string) error
	GetVirtualMachine(ctx context.Context, id string) (VirtualMachine, error)
	ListVirtualMachines(ctx context.Context, account string) ([]VirtualMachine, error)
}
