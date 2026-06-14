// Package controlplane is the cloud control plane: it implements the northbound
// CloudStack contract (internal/cloudstack) and reconciles the virtual machines
// clients ask for onto Kubernetes pods (internal/kube).
//
// It keeps no desired state of its own — that lives in a Store, so one core runs
// over in-memory, file, or other backends (single core, multi store). The
// control plane is an agent.Agent (the reconciler) and runs next to other
// operators such as the pod Reaper on the shared agent runtime (multi operator).
// Writes record desired state; a level-triggered reconcile loop converges actual
// pods toward it, so Kubernetes is the reconciler that realizes the CloudStack
// API (ADR-0007).
package controlplane

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/unboxd-cloud/platform/internal/cloudstack"
	"github.com/unboxd-cloud/platform/internal/kube"
)

// target is the desired lifecycle state the reconciler drives a VM toward.
type target int

const (
	targetRunning target = iota
	targetStopped
	targetDestroyed
)

func (t target) String() string {
	switch t {
	case targetRunning:
		return "running"
	case targetStopped:
		return "stopped"
	case targetDestroyed:
		return "destroyed"
	default:
		return "unknown"
	}
}

// MarshalText/UnmarshalText let stores persist the target as a stable string.
func (t target) MarshalText() ([]byte, error) { return []byte(t.String()), nil }

func (t *target) UnmarshalText(b []byte) error {
	switch string(b) {
	case "running":
		*t = targetRunning
	case "stopped":
		*t = targetStopped
	case "destroyed":
		*t = targetDestroyed
	default:
		return fmt.Errorf("controlplane: unknown target %q", b)
	}
	return nil
}

// ControlPlane implements cloudstack.Contract by reconciling VMs onto pods.
type ControlPlane struct {
	pods  kube.PodManager
	store Store

	// Deployable catalog (seeded at construction, read-only thereafter).
	zones     map[string]cloudstack.Zone
	offerings map[string]cloudstack.ServiceOffering
	templates map[string]cloudstack.Template

	now func() time.Time
}

// New returns a ControlPlane backed by an in-memory store.
func New(pods kube.PodManager) *ControlPlane { return NewWithStore(pods, NewMemStore()) }

// NewWithStore returns a ControlPlane backed by the given pod manager and store,
// seeded with a default zone, service offerings, and templates so it is usable
// out of the box.
func NewWithStore(pods kube.PodManager, store Store) *ControlPlane {
	cp := &ControlPlane{
		pods:      pods,
		store:     store,
		zones:     map[string]cloudstack.Zone{},
		offerings: map[string]cloudstack.ServiceOffering{},
		templates: map[string]cloudstack.Template{},
		now:       time.Now,
	}
	cp.zones["zone-1"] = cloudstack.Zone{ID: "zone-1", Name: "default"}
	cp.offerings["so-small"] = cloudstack.ServiceOffering{ID: "so-small", Name: "small", CPUNumber: 1, Memory: 512}
	cp.offerings["so-medium"] = cloudstack.ServiceOffering{ID: "so-medium", Name: "medium", CPUNumber: 2, Memory: 2048}
	cp.templates["tmpl-nginx"] = cloudstack.Template{ID: "tmpl-nginx", Name: "nginx", OSType: "Linux", Image: "docker.io/library/nginx:stable"}
	cp.templates["tmpl-alpine"] = cloudstack.Template{ID: "tmpl-alpine", Name: "alpine", OSType: "Linux", Image: "docker.io/library/alpine:3"}
	return cp
}

// Name implements agent.Agent.
func (c *ControlPlane) Name() string { return "cloud-control-plane" }

// Zones returns the deployable availability zones, sorted by ID.
func (c *ControlPlane) Zones() []cloudstack.Zone {
	out := make([]cloudstack.Zone, 0, len(c.zones))
	for _, z := range c.zones {
		out = append(out, z)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// ServiceOfferings returns the deployable compute flavors, sorted by ID.
func (c *ControlPlane) ServiceOfferings() []cloudstack.ServiceOffering {
	out := make([]cloudstack.ServiceOffering, 0, len(c.offerings))
	for _, o := range c.offerings {
		out = append(out, o)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// Templates returns the deployable templates, sorted by ID.
func (c *ControlPlane) Templates() []cloudstack.Template {
	out := make([]cloudstack.Template, 0, len(c.templates))
	for _, t := range c.templates {
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// DeployVirtualMachine validates the request against the catalog, records the VM
// as desired-Running, and returns it in the Starting state. The reconcile loop
// creates the backing pod.
func (c *ControlPlane) DeployVirtualMachine(ctx context.Context, req cloudstack.DeployVMRequest) (cloudstack.VirtualMachine, error) {
	if err := req.Validate(); err != nil {
		return cloudstack.VirtualMachine{}, err
	}
	if _, ok := c.zones[req.ZoneID]; !ok {
		return cloudstack.VirtualMachine{}, fmt.Errorf("%w: zone %q", cloudstack.ErrNotFound, req.ZoneID)
	}
	if _, ok := c.offerings[req.ServiceOfferingID]; !ok {
		return cloudstack.VirtualMachine{}, fmt.Errorf("%w: service offering %q", cloudstack.ErrNotFound, req.ServiceOfferingID)
	}
	if _, ok := c.templates[req.TemplateID]; !ok {
		return cloudstack.VirtualMachine{}, fmt.Errorf("%w: template %q", cloudstack.ErrNotFound, req.TemplateID)
	}
	id, err := c.store.NextID(ctx)
	if err != nil {
		return cloudstack.VirtualMachine{}, fmt.Errorf("controlplane: allocate id: %w", err)
	}
	vm := cloudstack.VirtualMachine{
		ID:                id,
		Name:              req.Name,
		DisplayName:       req.DisplayName,
		Account:           req.Account,
		ZoneID:            req.ZoneID,
		TemplateID:        req.TemplateID,
		ServiceOfferingID: req.ServiceOfferingID,
		State:             cloudstack.StateStarting,
		Created:           c.now().UTC(),
	}
	if err := c.store.Put(ctx, Record{VM: vm, Target: targetRunning}); err != nil {
		return cloudstack.VirtualMachine{}, fmt.Errorf("controlplane: persist vm: %w", err)
	}
	log.Printf("control-plane: deploy vm %s (%s) for account %s", vm.ID, vm.Name, vm.Account)
	return vm, nil
}

// DeliverVirtualMachine deploys a VM and drives it to running inline — end-to-end
// "direct delivery" in a single call. It is the synchronous counterpart to
// DeployVirtualMachine plus the async reconcile loop, for clients and scripts
// that want the workload ready on return.
func (c *ControlPlane) DeliverVirtualMachine(ctx context.Context, req cloudstack.DeployVMRequest) (cloudstack.VirtualMachine, error) {
	vm, err := c.DeployVirtualMachine(ctx, req)
	if err != nil {
		return cloudstack.VirtualMachine{}, err
	}
	rec, ok, err := c.store.Get(ctx, vm.ID)
	if err != nil {
		return cloudstack.VirtualMachine{}, err
	}
	if !ok {
		return cloudstack.VirtualMachine{}, fmt.Errorf("%w: vm %q", cloudstack.ErrNotFound, vm.ID)
	}
	if err := c.reconcileVM(ctx, rec); err != nil {
		return cloudstack.VirtualMachine{}, err
	}
	return c.GetVirtualMachine(ctx, vm.ID)
}

// StartVirtualMachine marks the VM desired-Running (Starting until reconciled).
func (c *ControlPlane) StartVirtualMachine(ctx context.Context, id string) (cloudstack.VirtualMachine, error) {
	return c.transition(ctx, id, targetRunning, cloudstack.StateStarting)
}

// StopVirtualMachine marks the VM desired-Stopped (Stopping until reconciled).
func (c *ControlPlane) StopVirtualMachine(ctx context.Context, id string) (cloudstack.VirtualMachine, error) {
	return c.transition(ctx, id, targetStopped, cloudstack.StateStopping)
}

func (c *ControlPlane) transition(ctx context.Context, id string, t target, s cloudstack.VMState) (cloudstack.VirtualMachine, error) {
	rec, ok, err := c.store.Get(ctx, id)
	if err != nil {
		return cloudstack.VirtualMachine{}, err
	}
	if !ok {
		return cloudstack.VirtualMachine{}, fmt.Errorf("%w: vm %q", cloudstack.ErrNotFound, id)
	}
	rec.Target = t
	rec.VM.State = s
	if err := c.store.Put(ctx, rec); err != nil {
		return cloudstack.VirtualMachine{}, err
	}
	return rec.VM, nil
}

// DestroyVirtualMachine marks the VM for teardown; the reconcile loop deletes
// the pod and removes the record.
func (c *ControlPlane) DestroyVirtualMachine(ctx context.Context, id string) error {
	rec, ok, err := c.store.Get(ctx, id)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("%w: vm %q", cloudstack.ErrNotFound, id)
	}
	rec.Target = targetDestroyed
	rec.VM.State = cloudstack.StateStopping
	return c.store.Put(ctx, rec)
}

// GetVirtualMachine returns a VM by ID.
func (c *ControlPlane) GetVirtualMachine(ctx context.Context, id string) (cloudstack.VirtualMachine, error) {
	rec, ok, err := c.store.Get(ctx, id)
	if err != nil {
		return cloudstack.VirtualMachine{}, err
	}
	if !ok {
		return cloudstack.VirtualMachine{}, fmt.Errorf("%w: vm %q", cloudstack.ErrNotFound, id)
	}
	return rec.VM, nil
}

// ListVirtualMachines returns the VMs for an account ("" lists all), sorted by ID.
func (c *ControlPlane) ListVirtualMachines(ctx context.Context, account string) ([]cloudstack.VirtualMachine, error) {
	recs, err := c.store.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]cloudstack.VirtualMachine, 0, len(recs))
	for _, rec := range recs {
		if account == "" || rec.VM.Account == account {
			out = append(out, rec.VM)
		}
	}
	return out, nil
}

// Reconcile converges actual pods toward the desired VM set (level-triggered):
// running VMs get a pod, stopped VMs have theirs removed, destroyed VMs are torn
// down. It is idempotent. Implements agent.Agent so the operator runtime drives it.
func (c *ControlPlane) Reconcile(ctx context.Context) error {
	recs, err := c.store.List(ctx)
	if err != nil {
		return fmt.Errorf("controlplane: list desired: %w", err)
	}
	for _, rec := range recs {
		if err := c.reconcileVM(ctx, rec); err != nil {
			return err
		}
	}
	return nil
}

// reconcileVM converges a single record. Only writes back when state changes, so
// a steady state produces no store churn (important for file/remote stores).
func (c *ControlPlane) reconcileVM(ctx context.Context, rec Record) error {
	vm := rec.VM
	ns := namespace(vm.Account)
	name := podName(vm)
	switch rec.Target {
	case targetRunning:
		pod, err := c.pods.Get(ctx, ns, name)
		if errors.Is(err, kube.ErrNotFound) {
			pod, err = c.pods.Create(ctx, c.podSpec(vm))
			if err != nil {
				return fmt.Errorf("reconcile vm %s: create pod: %w", vm.ID, err)
			}
			log.Printf("control-plane: vm %s -> pod %s/%s (%s)", vm.ID, pod.Namespace, pod.Name, pod.Phase)
		} else if err != nil {
			return fmt.Errorf("reconcile vm %s: get pod: %w", vm.ID, err)
		}
		return c.setState(ctx, rec, stateForPhase(pod.Phase))
	case targetStopped:
		if err := c.pods.Delete(ctx, ns, name); err != nil && !errors.Is(err, kube.ErrNotFound) {
			return fmt.Errorf("reconcile vm %s: stop: %w", vm.ID, err)
		}
		return c.setState(ctx, rec, cloudstack.StateStopped)
	case targetDestroyed:
		if err := c.pods.Delete(ctx, ns, name); err != nil && !errors.Is(err, kube.ErrNotFound) {
			return fmt.Errorf("reconcile vm %s: destroy: %w", vm.ID, err)
		}
		log.Printf("control-plane: destroyed vm %s", vm.ID)
		if err := c.store.Delete(ctx, vm.ID); err != nil {
			return fmt.Errorf("reconcile vm %s: delete record: %w", vm.ID, err)
		}
	}
	return nil
}

// setState persists a new VM state only if it changed.
func (c *ControlPlane) setState(ctx context.Context, rec Record, s cloudstack.VMState) error {
	if rec.VM.State == s {
		return nil
	}
	rec.VM.State = s
	if err := c.store.Put(ctx, rec); err != nil {
		return fmt.Errorf("reconcile vm %s: persist state: %w", rec.VM.ID, err)
	}
	return nil
}

// podSpec maps a VM (template + service offering) to the pod that backs it.
func (c *ControlPlane) podSpec(vm cloudstack.VirtualMachine) kube.PodSpec {
	tmpl := c.templates[vm.TemplateID]
	so := c.offerings[vm.ServiceOfferingID]
	return kube.PodSpec{
		Namespace: namespace(vm.Account),
		Name:      podName(vm),
		Image:     tmpl.Image,
		CPUNumber: so.CPUNumber,
		MemoryMB:  so.Memory,
		Labels: map[string]string{
			"app.kubernetes.io/managed-by": "unboxd-control-plane",
			"unboxd.cloud/vm-id":           vm.ID,
			"unboxd.cloud/account":         vm.Account,
		},
	}
}

// namespace isolates each tenant's pods; podName backs a VM by name within it.
func namespace(account string) string             { return "tenant-" + account }
func podName(vm cloudstack.VirtualMachine) string { return "vm-" + vm.Name }

func stateForPhase(p kube.Phase) cloudstack.VMState {
	switch p {
	case kube.PodRunning:
		return cloudstack.StateRunning
	case kube.PodPending:
		return cloudstack.StateStarting
	default:
		return cloudstack.StateError
	}
}
