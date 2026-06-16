// Package workstation provides cloud workstations: managed, single-pod developer
// environments ("one desk") that expose multiple ports (IDE, app previews, …)
// over the platform's Kubernetes pod manager. A workstation is the Code First
// Cloud's "where any one can build" surface — a ready-to-use dev environment
// delivered as a pod. Single desk per workstation; multi port per desk.
package workstation

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/unboxd-cloud/platform/internal/kube"
)

// ErrNotFound is returned when a workstation does not exist.
var ErrNotFound = errors.New("workstation: not found")

// State mirrors the lifecycle of a workstation.
type State string

const (
	StateStarting State = "Starting"
	StateRunning  State = "Running"
	StateStopped  State = "Stopped"
)

const (
	defaultImage = "codercom/code-server:latest"
	defaultPort  = 8080
)

// Workstation is a single developer environment (one desk) exposing one or more
// ports.
type Workstation struct {
	ID        string    `json:"id"`
	Account   string    `json:"account"`
	Name      string    `json:"name"`
	Image     string    `json:"image"`
	Ports     []int     `json:"ports"`
	State     State     `json:"state"`
	Endpoints []string  `json:"endpoints"` // host:port per exposed port
	Created   time.Time `json:"created"`
}

// LaunchRequest asks for a workstation. Image and Ports default to a code-server
// IDE on :8080 when omitted.
type LaunchRequest struct {
	Account string `json:"account"`
	Name    string `json:"name"`
	Image   string `json:"image,omitempty"`
	Ports   []int  `json:"ports,omitempty"`
}

// Manager launches and tracks workstations on a pod manager.
type Manager struct {
	pods kube.PodManager
	host string // host used to render endpoints

	mu    sync.Mutex
	desks map[string]Workstation
	seq   int64
	now   func() time.Time
}

// NewManager returns a workstation Manager. host renders endpoints (default
// "localhost").
func NewManager(pods kube.PodManager, host string) *Manager {
	if host == "" {
		host = "localhost"
	}
	return &Manager{pods: pods, host: host, desks: map[string]Workstation{}, now: time.Now}
}

// Launch creates a single workstation pod exposing the requested ports.
func (m *Manager) Launch(ctx context.Context, req LaunchRequest) (Workstation, error) {
	if req.Account == "" {
		return Workstation{}, fmt.Errorf("workstation: account is required")
	}
	if req.Name == "" {
		return Workstation{}, fmt.Errorf("workstation: name is required")
	}
	image := req.Image
	if image == "" {
		image = defaultImage
	}
	ports := req.Ports
	if len(ports) == 0 {
		ports = []int{defaultPort}
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.seq++
	id := fmt.Sprintf("ws-%d", m.seq)
	pod, err := m.pods.Create(ctx, kube.PodSpec{
		Namespace: "tenant-" + req.Account,
		Name:      "ws-" + req.Name,
		Image:     image,
		Ports:     ports,
		Labels: map[string]string{
			"app.kubernetes.io/managed-by": "unboxd-workstation",
			"unboxd.cloud/workstation-id":  id,
			"unboxd.cloud/account":         req.Account,
		},
	})
	if err != nil {
		return Workstation{}, fmt.Errorf("workstation: launch: %w", err)
	}
	ws := Workstation{
		ID:        id,
		Account:   req.Account,
		Name:      req.Name,
		Image:     image,
		Ports:     ports,
		State:     stateForPhase(pod.Phase),
		Endpoints: m.endpoints(ports),
		Created:   m.now().UTC(),
	}
	m.desks[id] = ws
	log.Printf("workstation: launched %s (%s) for %s on ports %v", id, image, req.Account, ports)
	return ws, nil
}

// Get returns a workstation by ID.
func (m *Manager) Get(_ context.Context, id string) (Workstation, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	ws, ok := m.desks[id]
	if !ok {
		return Workstation{}, ErrNotFound
	}
	return ws, nil
}

// List returns workstations for an account ("" lists all), sorted by ID.
func (m *Manager) List(_ context.Context, account string) ([]Workstation, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Workstation, 0, len(m.desks))
	for _, ws := range m.desks {
		if account == "" || ws.Account == account {
			out = append(out, ws)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

// Stop tears down a workstation and its pod.
func (m *Manager) Stop(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	ws, ok := m.desks[id]
	if !ok {
		return ErrNotFound
	}
	if err := m.pods.Delete(ctx, "tenant-"+ws.Account, "ws-"+ws.Name); err != nil && !errors.Is(err, kube.ErrNotFound) {
		return fmt.Errorf("workstation: stop: %w", err)
	}
	delete(m.desks, id)
	return nil
}

func (m *Manager) endpoints(ports []int) []string {
	out := make([]string, 0, len(ports))
	for _, p := range ports {
		out = append(out, fmt.Sprintf("%s:%d", m.host, p))
	}
	return out
}

func stateForPhase(p kube.Phase) State {
	switch p {
	case kube.PodRunning:
		return StateRunning
	case kube.PodPending:
		return StateStarting
	default:
		return StateStopped
	}
}
