// Package controlplane is the cloud control plane: it implements the northbound
// CloudStack contract (internal/cloudstack) and reconciles the virtual machines
// clients ask for onto Kubernetes pods (internal/kube). Writes record desired
// state; a level-triggered reconcile loop converges actual pods toward it, so
// Kubernetes is the reconciler that realizes the CloudStack API (ADR-0007). The
// control plane is itself an agent.Agent, so it runs on the shared operator
// runtime alongside the GitOps and orchestrator loops.
package controlplane

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sort"
	"sync"
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

// ControlPlane implements cloudstack.Contract by reconciling VMs onto pods.
type ControlPlane struct {
	pods kube.PodManager

	// Deployable catalog (seeded defaults; read-only in Phase 0).
	zones     map[string]cloudstack.Zone
	offerings map[string]cloudstack.ServiceOffering
	templates map[string]cloudstack.Template

	mu      sync.Mutex
	vms     map[string]cloudstack.VirtualMachine
	targets map[string]target
	seq     int64
	now     func() time.Time
}

// New returns a ControlPlane backed by the given pod manager, seeded with a
// default zone, service offerings, and templates so it is usable out of the box.
func New(pods kube.PodManager) *ControlPlane {
	cp := &ControlPlane{
		pods:      pods,
		zones:     map[string]cloudstack.Zone{},
		offerings: map[string]cloudstack.ServiceOffering{},
		templates: map[string]cloudstack.Template{},
		vms:       map[string]cloudstack.VirtualMachine{},
		targets:   map[string]target{},
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
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]cloudstack.Zone, 0, len(c.zones))
	for _, z := range c.zones {
		out = append(out, z)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// ServiceOfferings returns the deployable compute flavors, sorted by ID.
func (c *ControlPlane) ServiceOfferings() []cloudstack.ServiceOffering {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]cloudstack.ServiceOffering, 0, len(c.offerings))
	for _, o := range c.offerings {
		out = append(out, o)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// Templates returns the deployable templates, sorted by ID.
func (c *ControlPlane) Templates() []cloudstack.Template {
	c.mu.Lock()
	defer c.mu.Unlock()
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
func (c *ControlPlane) DeployVirtualMachine(_ context.Context, req cloudstack.DeployVMRequest) (cloudstack.VirtualMachine, error) {
	if err := req.Validate(); err != nil {
		return cloudstack.VirtualMachine{}, err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.zones[req.ZoneID]; !ok {
		return cloudstack.VirtualMachine{}, fmt.Errorf("%w: zone %q", cloudstack.ErrNotFound, req.ZoneID)
	}
	if _, ok := c.offerings[req.ServiceOfferingID]; !ok {
		return cloudstack.VirtualMachine{}, fmt.Errorf("%w: service offering %q", cloudstack.ErrNotFound, req.ServiceOfferingID)
	}
	if _, ok := c.templates[req.TemplateID]; !ok {
		return cloudstack.VirtualMachine{}, fmt.Errorf("%w: template %q", cloudstack.ErrNotFound, req.TemplateID)
	}
	c.seq++
	vm := cloudstack.VirtualMachine{
		ID:                fmt.Sprintf("vm-%d", c.seq),
		Name:              req.Name,
		DisplayName:       req.DisplayName,
		Account:           req.Account,
		ZoneID:            req.ZoneID,
		TemplateID:        req.TemplateID,
		ServiceOfferingID: req.ServiceOfferingID,
		State:             cloudstack.StateStarting,
		Created:           c.now().UTC(),
	}
	c.vms[vm.ID] = vm
	c.targets[vm.ID] = targetRunning
	log.Printf("control-plane: deploy vm %s (%s) for account %s", vm.ID, vm.Name, vm.Account)
	return vm, nil
}

// StartVirtualMachine marks the VM desired-Running (Starting until reconciled).
func (c *ControlPlane) StartVirtualMachine(_ context.Context, id string) (cloudstack.VirtualMachine, error) {
	return c.transition(id, targetRunning, cloudstack.StateStarting)
}

// StopVirtualMachine marks the VM desired-Stopped (Stopping until reconciled).
func (c *ControlPlane) StopVirtualMachine(_ context.Context, id string) (cloudstack.VirtualMachine, error) {
	return c.transition(id, targetStopped, cloudstack.StateStopping)
}

func (c *ControlPlane) transition(id string, t target, s cloudstack.VMState) (cloudstack.VirtualMachine, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	vm, ok := c.vms[id]
	if !ok {
		return cloudstack.VirtualMachine{}, fmt.Errorf("%w: vm %q", cloudstack.ErrNotFound, id)
	}
	c.targets[id] = t
	vm.State = s
	c.vms[id] = vm
	return vm, nil
}

// DestroyVirtualMachine marks the VM for teardown; the reconcile loop deletes
// the pod and removes the record.
func (c *ControlPlane) DestroyVirtualMachine(_ context.Context, id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	vm, ok := c.vms[id]
	if !ok {
		return fmt.Errorf("%w: vm %q", cloudstack.ErrNotFound, id)
	}
	c.targets[id] = targetDestroyed
	vm.State = cloudstack.StateStopping
	c.vms[id] = vm
	return nil
}

// GetVirtualMachine returns a VM by ID.
func (c *ControlPlane) GetVirtualMachine(_ context.Context, id string) (cloudstack.VirtualMachine, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	vm, ok := c.vms[id]
	if !ok {
		return cloudstack.VirtualMachine{}, fmt.Errorf("%w: vm %q", cloudstack.ErrNotFound, id)
	}
	return vm, nil
}

// ListVirtualMachines returns the VMs for an account ("" lists all), sorted by ID.
func (c *ControlPlane) ListVirtualMachines(_ context.Context, account string) ([]cloudstack.VirtualMachine, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]cloudstack.VirtualMachine, 0, len(c.vms))
	for _, vm := range c.vms {
		if account == "" || vm.Account == account {
			out = append(out, vm)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

// Reconcile converges actual pods toward the desired VM set (level-triggered):
// running VMs get a pod, stopped VMs have theirs removed, destroyed VMs are torn
// down. It is idempotent. Implements agent.Agent so the operator runtime drives it.
func (c *ControlPlane) Reconcile(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for id := range c.targets {
		if err := c.reconcileVM(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

// reconcileVM converges a single VM. Caller holds c.mu.
func (c *ControlPlane) reconcileVM(ctx context.Context, id string) error {
	vm := c.vms[id]
	ns := namespace(vm.Account)
	name := podName(vm)
	switch c.targets[id] {
	case targetRunning:
		pod, err := c.pods.Get(ctx, ns, name)
		if errors.Is(err, kube.ErrNotFound) {
			pod, err = c.pods.Create(ctx, c.podSpec(vm))
			if err != nil {
				return fmt.Errorf("reconcile vm %s: create pod: %w", id, err)
			}
			log.Printf("control-plane: vm %s -> pod %s/%s (%s)", id, pod.Namespace, pod.Name, pod.Phase)
		} else if err != nil {
			return fmt.Errorf("reconcile vm %s: get pod: %w", id, err)
		}
		vm.State = stateForPhase(pod.Phase)
		c.vms[id] = vm
	case targetStopped:
		if err := c.pods.Delete(ctx, ns, name); err != nil && !errors.Is(err, kube.ErrNotFound) {
			return fmt.Errorf("reconcile vm %s: stop: %w", id, err)
		}
		vm.State = cloudstack.StateStopped
		c.vms[id] = vm
	case targetDestroyed:
		if err := c.pods.Delete(ctx, ns, name); err != nil && !errors.Is(err, kube.ErrNotFound) {
			return fmt.Errorf("reconcile vm %s: destroy: %w", id, err)
		}
		log.Printf("control-plane: destroyed vm %s", id)
		delete(c.vms, id)
		delete(c.targets, id)
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
