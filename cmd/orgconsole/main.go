// Command orgconsole is the organization (tenant) admin console — the
// org-scoped counterpart to the platform-wide admin panel. An org admin manages
// their members/personas, compliance posture, available catalog, and spend.
//
// FE: htmx over stdlib html/template (see docs/ui.md). Org membership and
// compliance posture are kept behind the same database-agnostic Store seams; the
// catalog/billing/compliance data comes through the SDK.
package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/unboxd-cloud/platform/internal/api"
	"github.com/unboxd-cloud/platform/internal/billing"
	"github.com/unboxd-cloud/platform/internal/catalog"
	"github.com/unboxd-cloud/platform/internal/compliance"
	"github.com/unboxd-cloud/platform/internal/metering"
	"github.com/unboxd-cloud/platform/internal/server"
	"github.com/unboxd-cloud/platform/internal/tenant"
	"github.com/unboxd-cloud/platform/pkg/sdk"
)

type console struct {
	orgID    string
	tenants  *tenant.MemStore
	profiles *compliance.MemStore
	sdk      *sdk.Client
	tmpl     *template.Template
}

func main() {
	orgID := envOr("ORG_ID", "acme")
	orgName := envOr("ORG_NAME", "Acme Inc.")

	client := sdk.New()
	overrideURL(&client.Catalog, "CATALOG_URL")
	overrideURL(&client.Billing, "BILLING_URL")
	overrideURL(&client.Compliance, "COMPLIANCE_URL")
	overrideURL(&client.Metering, "METERING_URL")
	client.Tenant = orgID

	c := &console{
		orgID:    orgID,
		tenants:  tenant.NewMemStore(),
		profiles: compliance.NewMemStore(),
		sdk:      client,
		tmpl:     template.Must(template.New("org").Parse(orgTemplate)),
	}
	// Seed the org (the console manages exactly one organization).
	if _, err := c.tenants.Create(tenant.Tenant{ID: orgID, Name: orgName,
		Members: []tenant.Member{{Subject: "owner@" + orgID, Profile: tenant.ProfileBillingAdmin}}}); err != nil {
		log.Fatalf("seed org: %v", err)
	}
	_ = c.profiles.Set(compliance.Profile{TenantID: orgID, Jurisdiction: "EU-DE",
		Frameworks: []string{"GDPR", "SOC2"}, DataResidency: []string{"EU-DE", "EU-FR"}})

	mux := http.NewServeMux()
	mux.HandleFunc("/", c.home)
	mux.HandleFunc("/members", c.addMember)
	mux.HandleFunc("/compliance", c.setCompliance)
	mux.HandleFunc("/spend", c.spend)

	addr := envOr("ORGCONSOLE_ADDR", ":8085")
	log.Printf("org admin console for %q on %s", orgID, addr)
	log.Fatal(server.New(addr, mux).ListenAndServe())
}

type view struct {
	Org        tenant.Tenant
	Profiles   []tenant.Profile
	Compliance compliance.Profile
	Offerings  []catalog.Offering
	Spend      *api.RateResponse
	Notes      []string
}

func (c *console) view(ctx context.Context) view {
	org, _ := c.tenants.Get(c.orgID)
	prof, _ := c.profiles.Get(c.orgID)
	v := view{Org: org, Compliance: prof, Profiles: tenant.ValidProfiles()}
	offers, err := c.sdk.ListOfferings(ctx, "")
	if err != nil {
		v.Notes = append(v.Notes, "catalog unavailable: "+err.Error())
	} else {
		v.Offerings = offers
	}
	return v
}

func (c *console) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	c.render(w, c.view(ctx))
}

func (c *console) addMember(w http.ResponseWriter, r *http.Request) {
	subject := r.FormValue("subject")
	profile := tenant.Profile(r.FormValue("profile"))
	org, _ := c.tenants.Get(c.orgID)
	v := view{}
	if subject == "" || !profile.Valid() {
		v = c.view(r.Context())
		v.Notes = append(v.Notes, "invalid member: subject required and profile must be valid")
		c.render(w, v)
		return
	}
	// Re-create the org with the appended member (MemStore is create/replace).
	org.Members = append(org.Members, tenant.Member{Subject: subject, Profile: profile})
	c.tenants = tenant.NewMemStore()
	_, _ = c.tenants.Create(org)
	c.render(w, c.view(r.Context()))
}

func (c *console) setCompliance(w http.ResponseWriter, r *http.Request) {
	p := compliance.Profile{
		TenantID:      c.orgID,
		Jurisdiction:  r.FormValue("jurisdiction"),
		Frameworks:    splitCSV(r.FormValue("frameworks")),
		DataResidency: splitCSV(r.FormValue("residency")),
	}
	_ = c.profiles.Set(p)
	c.render(w, c.view(r.Context()))
}

func (c *console) spend(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	prof, _ := c.profiles.Get(c.orgID)
	v := c.view(ctx)
	resp, err := c.sdk.Rate(ctx, api.RateRequest{
		TenantID:     c.orgID,
		Jurisdiction: prof.Jurisdiction,
		Events: []metering.UsageEvent{
			{TenantID: c.orgID, Meter: "compute.vcpu.hour", Quantity: 800},
			{TenantID: c.orgID, Meter: "s3.request.million", Quantity: 12},
			{TenantID: c.orgID, Meter: "ai.tokens.million", Quantity: 5},
		},
		Partner: &billing.Partner{ID: "msp", Mode: billing.ModeServiceProvider, Rate: 0.10},
	})
	if err != nil {
		v.Notes = append(v.Notes, "spend preview unavailable: "+err.Error())
	} else {
		v.Spend = &resp
	}
	c.render(w, v)
}

func (c *console) render(w http.ResponseWriter, v view) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := c.tmpl.Execute(w, v); err != nil {
		server.Error(w, http.StatusInternalServerError, err.Error())
	}
}

func splitCSV(s string) []string {
	var out []string
	for _, p := range splitComma(s) {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func splitComma(s string) []string {
	var out []string
	cur := ""
	for _, r := range s {
		if r == ',' || r == ' ' {
			if cur != "" {
				out = append(out, cur)
				cur = ""
			}
			continue
		}
		cur += string(r)
	}
	if cur != "" {
		out = append(out, cur)
	}
	return out
}

func overrideURL(field *string, env string) {
	if v := os.Getenv(env); v != "" {
		*field = v
	}
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

const orgTemplate = `<!doctype html>
<html lang="en"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Org Console — {{.Org.Name}}</title>
<script src="https://unpkg.com/htmx.org@1.9.12"></script>
<style>
  body{font:14px/1.5 system-ui,sans-serif;margin:0;background:#0f1117;color:#e6e6e6}
  header{padding:16px 24px;background:#161922;border-bottom:1px solid #262b38}
  h1{font-size:18px;margin:0}
  main{padding:24px;display:grid;gap:16px;grid-template-columns:repeat(auto-fit,minmax(320px,1fr))}
  .card{background:#161922;border:1px solid #262b38;border-radius:10px;padding:16px}
  .card h2{font-size:13px;text-transform:uppercase;letter-spacing:.05em;color:#8b93a7;margin:0 0 10px}
  .span2{grid-column:1/-1}
  .pill{display:inline-block;background:#222838;border:1px solid #2f3750;border-radius:999px;padding:2px 10px;margin:2px;font-size:12px}
  table{width:100%;border-collapse:collapse;font-size:13px}
  td,th{text-align:left;padding:6px 8px;border-bottom:1px solid #232838}
  input,select{background:#0f1320;border:1px solid #2a3346;color:#e6e6e6;border-radius:8px;padding:6px 10px;margin:2px}
  button{background:#2f6df6;color:#fff;border:0;border-radius:8px;padding:8px 14px;cursor:pointer}
  .note{background:#2a1d1d;border:1px solid #5a2d2d;color:#f0bcbc;padding:8px 12px;border-radius:8px}
</style></head>
<body>
<header><h1>Organization Console — {{.Org.Name}} <code>({{.Org.ID}})</code></h1></header>
<main>
  {{range .Notes}}<div class="card span2 note">{{.}}</div>{{end}}

  <div class="card"><h2>Members &amp; personas</h2>
    <table><tr><th>Subject</th><th>Profile</th></tr>
    {{range .Org.Members}}<tr><td>{{.Subject}}</td><td><span class="pill">{{.Profile}}</span></td></tr>{{end}}</table>
    <form hx-post="/members" hx-target="body" hx-swap="outerHTML" style="margin-top:10px">
      <input type="text" name="subject" placeholder="user@org" required>
      <select name="profile">
        {{range .Profiles}}<option value="{{.}}">{{.}}</option>{{end}}
      </select>
      <button type="submit">Add member</button>
    </form>
  </div>

  <div class="card"><h2>Compliance posture</h2>
    <p>Jurisdiction: <span class="pill">{{.Compliance.Jurisdiction}}</span></p>
    <p>Frameworks: {{range .Compliance.Frameworks}}<span class="pill">{{.}}</span>{{else}}<em>none</em>{{end}}</p>
    <p>Data residency: {{range .Compliance.DataResidency}}<span class="pill">{{.}}</span>{{else}}<em>any</em>{{end}}</p>
    <form hx-post="/compliance" hx-target="body" hx-swap="outerHTML">
      <input type="text" name="jurisdiction" value="{{.Compliance.Jurisdiction}}" placeholder="EU-DE">
      <input type="text" name="frameworks" placeholder="GDPR,SOC2">
      <input type="text" name="residency" placeholder="EU-DE,EU-FR">
      <button type="submit">Update</button>
    </form>
  </div>

  <div class="card span2"><h2>Available catalog</h2>
    <table><tr><th>ID</th><th>Project</th><th>Category</th><th>Meters</th><th>Certifications</th></tr>
    {{range .Offerings}}<tr>
      <td>{{.ID}}</td><td>{{.Project}}</td><td>{{.Category}}</td>
      <td>{{range .Meters}}<span class="pill">{{.}}</span>{{end}}</td>
      <td>{{range .Certifications}}<span class="pill">{{.}}</span>{{end}}</td>
    </tr>{{else}}<tr><td colspan="5"><em>unavailable</em></td></tr>{{end}}</table>
  </div>

  <div class="card span2"><h2>Spend preview (this org, MSP +10%, taxed)</h2>
    {{if .Spend}}
      <table>
        <tr><th>Usage subtotal</th><td>{{.Spend.Invoice.Total}} {{.Spend.Invoice.Currency}}</td></tr>
        {{with .Spend.Settlement}}<tr><th>Gross to customer</th><td>{{.GrossToCustomer}} {{.Currency}}</td></tr>{{end}}
        {{with .Spend.Tax}}<tr><th>Tax</th><td>{{.TaxTotal}}</td></tr><tr><th>Gross incl. tax</th><td>{{.Gross}} {{.Currency}}</td></tr>{{end}}
      </table>
    {{else}}<button hx-get="/spend" hx-target="body" hx-swap="outerHTML">Run spend preview</button>{{end}}
  </div>
</main></body></html>`
