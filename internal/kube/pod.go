// Package kube is the platform's seam over Kubernetes pod lifecycle — the "pod
// manager". The cloud control plane's reconciler drives desired compute onto
// pods through PodManager, so Kubernetes is the engine that realizes the
// (CloudStack) contract. The in-memory manager here proves the reconcile path
// with no cluster dependency (Phase 0, stdlib-only, like internal/provider); a
// real client-go-backed manager drops in behind the same interface without
// touching the control plane.
package kube

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
)

// ErrNotFound is returned when a pod does not exist.
var ErrNotFound = errors.New("kube: pod not found")

// Phase mirrors the Kubernetes pod phase.
type Phase string

const (
	PodPending Phase = "Pending"
	PodRunning Phase = "Running"
	PodFailed  Phase = "Failed"
)

// PodSpec is the minimal desired spec for a pod.
type PodSpec struct {
	Namespace string            // tenant namespace
	Name      string            // unique within the namespace
	Image     string            // container image to run
	CPUNumber int               // requested vCPUs
	MemoryMB  int               // requested memory (MB)
	Ports     []int             // container ports to expose
	Labels    map[string]string // selector labels
}

// Pod is a handle to a (stub) Kubernetes pod.
type Pod struct {
	Namespace string            `json:"namespace"`
	Name      string            `json:"name"`
	UID       string            `json:"uid"`
	Image     string            `json:"image"`
	Phase     Phase             `json:"phase"`
	Node      string            `json:"node"`
	Ports     []int             `json:"ports,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
}

// PodManager is the seam over Kubernetes pod lifecycle.
type PodManager interface {
	Create(ctx context.Context, spec PodSpec) (Pod, error)
	Get(ctx context.Context, namespace, name string) (Pod, error)
	List(ctx context.Context, namespace string) ([]Pod, error)
	Delete(ctx context.Context, namespace, name string) error
}

// memManager is an in-memory PodManager. Created pods schedule immediately to a
// synthetic node and report Running, mirroring the one-shot behavior of the
// provider stubs (a real scheduler/kubelet lands behind the same interface).
type memManager struct {
	mu   sync.Mutex
	pods map[string]Pod // keyed by namespace/name
	seq  atomic.Int64
}

// NewManager returns an in-memory PodManager.
func NewManager() PodManager {
	return &memManager{pods: map[string]Pod{}}
}

func key(namespace, name string) string { return namespace + "/" + name }

func (m *memManager) Create(_ context.Context, spec PodSpec) (Pod, error) {
	if spec.Namespace == "" || spec.Name == "" {
		return Pod{}, fmt.Errorf("kube: namespace and name are required")
	}
	if spec.Image == "" {
		return Pod{}, fmt.Errorf("kube: image is required for %s/%s", spec.Namespace, spec.Name)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	k := key(spec.Namespace, spec.Name)
	if existing, ok := m.pods[k]; ok {
		return existing, nil // create is idempotent (like a server-side apply)
	}
	n := m.seq.Add(1)
	pod := Pod{
		Namespace: spec.Namespace,
		Name:      spec.Name,
		UID:       fmt.Sprintf("pod-%d", n),
		Image:     spec.Image,
		Phase:     PodRunning,
		Node:      fmt.Sprintf("node-%d", (n%3)+1),
		Ports:     spec.Ports,
		Labels:    spec.Labels,
	}
	m.pods[k] = pod
	return pod, nil
}

func (m *memManager) Get(_ context.Context, namespace, name string) (Pod, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	pod, ok := m.pods[key(namespace, name)]
	if !ok {
		return Pod{}, ErrNotFound
	}
	return pod, nil
}

func (m *memManager) List(_ context.Context, namespace string) ([]Pod, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Pod, 0, len(m.pods))
	for _, p := range m.pods {
		if namespace == "" || p.Namespace == namespace {
			out = append(out, p)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func (m *memManager) Delete(_ context.Context, namespace, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	k := key(namespace, name)
	if _, ok := m.pods[k]; !ok {
		return ErrNotFound
	}
	delete(m.pods, k)
	return nil
}
