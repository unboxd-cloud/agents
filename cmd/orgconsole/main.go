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
	"html"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/unboxd-cloud/platform/internal/api"
	"github.com/unboxd-cloud/platform/internal/billing"
	"github.com/unboxd-cloud/platform/internal/catalog"
	"github.com/unboxd-cloud/platform/internal/compliance"
	"github.com/unboxd-cloud/platform/internal/metering"
	"github.com/unboxd-cloud/platform/internal/s3"
	"github.com/unboxd-cloud/platform/internal/server"
	"github.com/unboxd-cloud/platform/internal/tenant"
	"github.com/unboxd-cloud/platform/internal/ui"
	"github.com/unboxd-cloud/platform/pkg/sdk"
)

type console struct {
	orgID     string
	tenants   *tenant.MemStore
	profiles  *compliance.MemStore
	snapshots *s3.MemStore
	sdk       *sdk.Client
	tmpl      *template.Template
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
		orgID:     orgID,
		tenants:   tenant.NewMemStore(),
		profiles:  compliance.NewMemStore(),
		snapshots: s3.NewMemStore(),
		sdk:       client,
		tmpl:      template.Must(template.New("org").Parse(orgTemplate)),
	}
	// Seed the org (the console manages exactly one organization).
	if _, err := c.tenants.Create(tenant.Tenant{ID: orgID, Name: orgName,
		Members: []tenant.Member{{Subject: "owner@" + orgID, Profile: tenant.ProfileBillingAdmin}}}); err != nil {
		log.Fatalf("seed org: %v", err)
	}
	_ = c.profiles.Set(compliance.Profile{TenantID: orgID, Jurisdiction: "EU-DE",
		Frameworks: []string{"GDPR", "SOC2"}, DataResidency: []string{"EU-DE", "EU-FR"}})
	// Seed S3 cluster-snapshot settings from the deploy-time environment.
	if st, ok := s3.FromEnv(orgID); ok {
		_ = c.snapshots.Set(st)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", c.home)
	mux.HandleFunc("/members", c.addMember)
	mux.HandleFunc("/compliance", c.setCompliance)
	mux.HandleFunc("/settings", c.setSnapshots)
	mux.HandleFunc("/spend", c.spend)
	mux.HandleFunc("/chat", c.chat)

	addr := envOr("ORGCONSOLE_ADDR", ":8085")
	log.Printf("org admin console for %q on %s", orgID, addr)
	log.Fatal(server.New(addr, mux).ListenAndServe())
}

type view struct {
	Org        tenant.Tenant
	Profiles   []tenant.Profile
	Compliance compliance.Profile
	Snapshots  s3.Settings
	Offerings  []catalog.Offering
	Spend      *api.RateResponse
	Notes      []string
}

func (c *console) view(ctx context.Context) view {
	org, _ := c.tenants.Get(c.orgID)
	prof, _ := c.profiles.Get(c.orgID)
	snap, _ := c.snapshots.Get(c.orgID)
	v := view{Org: org, Compliance: prof, Snapshots: snap.Redacted(), Profiles: tenant.ValidProfiles()}
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

func (c *console) setSnapshots(w http.ResponseWriter, r *http.Request) {
	cur, _ := c.snapshots.Get(c.orgID)
	st := s3.Settings{
		Scope:       c.orgID,
		Bucket:      strings.TrimSpace(r.FormValue("bucket")),
		Region:      strings.TrimSpace(r.FormValue("region")),
		Endpoint:    strings.TrimSpace(r.FormValue("endpoint")),
		Prefix:      strings.TrimSpace(r.FormValue("prefix")),
		AccessKeyID: strings.TrimSpace(r.FormValue("accessKeyId")),
		SecretKey:   s3.SecretOrKeep(r.FormValue("secretKey"), cur.SecretKey),
		Schedule:    strings.TrimSpace(r.FormValue("schedule")),
	}
	if v := strings.TrimSpace(r.FormValue("retentionDays")); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			st.RetentionDays = n
		}
	}
	v := c.view(r.Context())
	if err := c.snapshots.Set(st); err != nil {
		v.Notes = append(v.Notes, "invalid S3 settings: bucket and region are required")
	} else {
		v = c.view(r.Context())
	}
	c.render(w, v)
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

func (c *console) chat(w http.ResponseWriter, r *http.Request) {
	msg := strings.TrimSpace(r.FormValue("msg"))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(chatBubble(msg, c.assist(r.Context(), msg))))
}

// assist is the org console's assistant: it answers from the same data the
// console manages (members, compliance, catalog, spend).
func (c *console) assist(ctx context.Context, msg string) string {
	cmd := ""
	if f := strings.Fields(strings.ToLower(msg)); len(f) > 0 {
		cmd = f[0]
	}
	switch cmd {
	case "", "help":
		return "I can help with: members, compliance, catalog, spend. Ask away."
	case "members":
		org, _ := c.tenants.Get(c.orgID)
		var b strings.Builder
		for _, m := range org.Members {
			b.WriteString(m.Subject + " (" + string(m.Profile) + ")\n")
		}
		if b.Len() == 0 {
			return "no members yet"
		}
		return b.String()
	case "compliance":
		p, _ := c.profiles.Get(c.orgID)
		return "jurisdiction " + p.Jurisdiction + "; frameworks " + strings.Join(p.Frameworks, ", ")
	case "catalog":
		offers, err := c.sdk.ListOfferings(ctx, "")
		if err != nil {
			return "catalog unavailable: " + err.Error()
		}
		var b strings.Builder
		for _, o := range offers {
			b.WriteString(o.ID + " — " + o.Project + "\n")
		}
		if b.Len() == 0 {
			return "catalog is empty"
		}
		return b.String()
	case "spend":
		return "see the Spend preview card below, or ask me to estimate usage."
	default:
		return "ask me about members, compliance, catalog, or spend."
	}
}

func chatBubble(msg, resp string) string {
	return `<div class="msg user">` + html.EscapeString(msg) +
		`</div><div class="msg bot">` + html.EscapeString(resp) + `</div>`
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

var orgTemplate = ui.Head("Org Console — {{.Org.Name}}") + `
<body>
<header><h1>Organization Console — {{.Org.Name}} <code>({{.Org.ID}})</code></h1></header>
<main>
` + ui.ChatHome("/chat") + `
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

  <div class="card"><h2>k3s cluster snapshots (S3)</h2>
    <p>Where this org's k3s cluster snapshots are stored for disaster recovery.</p>
    <p>Bucket: <span class="pill">{{if .Snapshots.Bucket}}{{.Snapshots.Bucket}}{{else}}unset{{end}}</span>
       Region: <span class="pill">{{if .Snapshots.Region}}{{.Snapshots.Region}}{{else}}unset{{end}}</span>
       {{if .Snapshots.Schedule}}Schedule: <span class="pill">{{.Snapshots.Schedule}}</span>{{end}}
       {{if .Snapshots.RetentionDays}}Retention: <span class="pill">{{.Snapshots.RetentionDays}}d</span>{{end}}</p>
    <form hx-post="/settings" hx-target="body" hx-swap="outerHTML">
      <input type="text" name="bucket" value="{{.Snapshots.Bucket}}" placeholder="bucket" required>
      <input type="text" name="region" value="{{.Snapshots.Region}}" placeholder="us-east-1" required>
      <input type="text" name="endpoint" value="{{.Snapshots.Endpoint}}" placeholder="endpoint (S3-compatible, optional)">
      <input type="text" name="prefix" value="{{.Snapshots.Prefix}}" placeholder="prefix (optional)">
      <input type="text" name="accessKeyId" value="{{.Snapshots.AccessKeyID}}" placeholder="access key id">
      <input type="password" name="secretKey" value="{{.Snapshots.SecretKey}}" placeholder="secret key">
      <input type="text" name="schedule" value="{{.Snapshots.Schedule}}" placeholder="cron e.g. 0 */6 * * *">
      <input type="text" name="retentionDays" value="{{if .Snapshots.RetentionDays}}{{.Snapshots.RetentionDays}}{{end}}" placeholder="retention days">
      <button type="submit">Save S3 settings</button>
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
