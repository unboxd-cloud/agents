// Package platform is the composable Go SDK for the Unboxd platform.
//
// It composes the two lower-level pieces behind one small, typed entrypoint: the
// control-plane client (pkg/sdk) for catalog/metering/billing/compliance, and
// the agent language runtime (pkg/adl) for loading and validating .agent
// definitions. Both remain usable on their own — this package just combines
// them, consistent with the platform's composability principle (small, swappable
// units that combine into larger ones).
package platform

import (
	"fmt"
	"os"

	"github.com/unboxd-cloud/platform/pkg/adl"
	"github.com/unboxd-cloud/platform/pkg/sdk"
)

// Platform is the top-level SDK handle, composed of a control-plane client and
// the ADL agent runtime.
type Platform struct {
	client *sdk.Client
}

// Option configures a Platform.
type Option func(*Platform)

// WithClient supplies a preconfigured control-plane client (apply before
// WithTenant if you use both).
func WithClient(c *sdk.Client) Option {
	return func(p *Platform) {
		if c != nil {
			p.client = c
		}
	}
}

// WithTenant sets the tenant on the control-plane client.
func WithTenant(tenant string) Option {
	return func(p *Platform) { p.client.Tenant = tenant }
}

// New builds a Platform. By default it uses sdk.New() (standard local ports);
// pass options to compose a different client or tenant.
func New(opts ...Option) *Platform {
	p := &Platform{client: sdk.New()}
	for _, o := range opts {
		if o != nil {
			o(p)
		}
	}
	return p
}

// Control returns the control-plane client for catalog/billing/metering/compliance.
func (p *Platform) Control() *sdk.Client { return p.client }

// CompileAgent parses and validates an ADL agent definition, returning the
// compiled model and diagnostics.
func (p *Platform) CompileAgent(src string) adl.Result {
	return adl.Compile(src)
}

// LoadAgent reads, parses, and validates an agent (.agent) file, returning the
// composable Agent view and its diagnostics. A non-nil error means the file
// could not be read or its definition failed validation.
func (p *Platform) LoadAgent(path string) (*adl.Agent, []adl.Diagnostic, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	agent, diags := adl.Load(string(b))
	if adl.HasErrors(diags) {
		return agent, diags, fmt.Errorf("%s: agent definition has %d error(s)", path, countErrors(diags))
	}
	return agent, diags, nil
}

func countErrors(diags []adl.Diagnostic) int {
	n := 0
	for _, d := range diags {
		if d.Severity == adl.SeverityError {
			n++
		}
	}
	return n
}
