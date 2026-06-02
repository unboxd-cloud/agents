// Command admin is the platform administrator control panel.
//
// FE: htmx (lightweight, ~14kb, no build step) over stdlib html/template —
// framework-agnostic and dependency-free on the server. It provides:
//   - a chat interface that drives the control plane through the SDK,
//   - granular flow traces (APM) for every chat action, exportable as OTLP/JSON
//     to CNCF tools (Jaeger, Grafana Tempo, Prometheus via the OTel Collector),
//   - an embedded-BI section linking to Grafana/Jaeger/Superset for insights.
package main

import (
	"context"
	"fmt"
	"html"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/unboxd-cloud/platform/internal/api"
	"github.com/unboxd-cloud/platform/internal/asset"
	"github.com/unboxd-cloud/platform/internal/billing"
	"github.com/unboxd-cloud/platform/internal/metering"
	"github.com/unboxd-cloud/platform/internal/observe"
	"github.com/unboxd-cloud/platform/internal/provider"
	"github.com/unboxd-cloud/platform/internal/s3"
	"github.com/unboxd-cloud/platform/internal/server"
	"github.com/unboxd-cloud/platform/internal/ui"
	"github.com/unboxd-cloud/platform/pkg/sdk"
)

var tracer = observe.New(2000)

type biLinks struct {
	Grafana, Jaeger, Superset, Otel string
}

func main() {
	client := sdk.New()
	overrideURL(&client.Catalog, "CATALOG_URL")
	overrideURL(&client.Billing, "BILLING_URL")
	overrideURL(&client.Compliance, "COMPLIANCE_URL")
	overrideURL(&client.Metering, "METERING_URL")

	bi := biLinks{
		Grafana:  os.Getenv("GRAFANA_URL"),
		Jaeger:   os.Getenv("JAEGER_URL"),
		Superset: os.Getenv("SUPERSET_URL"),
		Otel:     os.Getenv("OTEL_COLLECTOR_URL"),
	}

	page := template.Must(template.New("page").Parse(pageTemplate))

	// Platform-wide S3 settings for k3s cluster snapshots, seeded at deploy time
	// from S3_* environment variables and editable below.
	snaps := s3.NewMemStore()
	if st, ok := s3.FromEnv("platform"); ok {
		_ = snaps.Set(st)
	}
	renderPage := func(w http.ResponseWriter) {
		cur, _ := snaps.Get("platform")
		_ = page.Execute(w, map[string]any{
			"Providers": provider.DefaultRegistry().Names(),
			"Modes":     operatingModes(),
			"BI":        bi,
			"S3":        cur.Redacted(),
		})
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		renderPage(w)
	})

	// Platform-wide S3 cluster-snapshot settings.
	mux.HandleFunc("/settings", func(w http.ResponseWriter, r *http.Request) {
		cur, _ := snaps.Get("platform")
		st := s3.Settings{
			Scope:       "platform",
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
		_ = snaps.Set(st)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		renderPage(w)
	})

	// htmx chat endpoint: returns an HTML fragment appended to the chat log.
	mux.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		msg := strings.TrimSpace(r.FormValue("msg"))
		frag := interpret(r.Context(), client, msg)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(frag))
	})

	// Flow traces (APM) as an HTML fragment.
	mux.HandleFunc("/debug/traces", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(renderTraces(tracer.Recent(50))))
	})

	// OTLP/JSON export for analysis in CNCF tools.
	mux.HandleFunc("/debug/traces.json", func(w http.ResponseWriter, _ *http.Request) {
		server.JSON(w, http.StatusOK, tracer.OTLP())
	})

	// Asset discovery & catalog for IT admins.
	assets := asset.NewMemStore()
	servicesSrc := func() []asset.Asset {
		out := []asset.Asset{}
		for name, url := range map[string]string{
			"catalog": client.Catalog, "billing": client.Billing,
			"compliance": client.Compliance, "metering": client.Metering,
		} {
			out = append(out, asset.Asset{ID: "service:" + name, Kind: "service", Name: name, Source: url})
		}
		return out
	}
	providersSrc := func() []asset.Asset {
		out := []asset.Asset{}
		for _, p := range provider.DefaultRegistry().Names() {
			out = append(out, asset.Asset{ID: "provider:" + p, Kind: "provider", Name: p, Source: "registry"})
		}
		return out
	}
	asset.Discover(assets, servicesSrc, providersSrc)

	mux.HandleFunc("/assets", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(renderAssets(assets)))
	})
	mux.HandleFunc("/assets/discover", func(w http.ResponseWriter, _ *http.Request) {
		asset.Discover(assets, servicesSrc, providersSrc)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(renderAssets(assets)))
	})
	mux.HandleFunc("/assets/tag", func(w http.ResponseWriter, r *http.Request) {
		if a, ok := assets.Get(r.FormValue("id")); ok {
			if tag := strings.TrimSpace(r.FormValue("tag")); tag != "" {
				a.Tags = append(a.Tags, tag)
				_, _ = assets.Upsert(a)
			}
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(renderAssets(assets)))
	})

	addr := envOr("ADMIN_ADDR", ":8080")
	log.Printf("admin control panel on %s", addr)
	log.Fatal(server.New(addr, mux).ListenAndServe())
}

// interpret is the chat command interpreter. Each command is traced end-to-end
// so the flow is debuggable to a granular level.
func interpret(ctx context.Context, c *sdk.Client, msg string) string {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	root := tracer.Start("", "", "chat", map[string]string{"msg": msg})
	fields := strings.Fields(msg)
	cmd := ""
	if len(fields) > 0 {
		cmd = strings.ToLower(fields[0])
	}

	var body, err = func() (string, error) {
		switch cmd {
		case "", "help":
			return helpText(), nil
		case "providers":
			return list(provider.DefaultRegistry().Names()), nil
		case "modes":
			return list(operatingModes()), nil
		case "catalog":
			profile := ""
			if len(fields) > 1 {
				profile = fields[1]
			}
			s := tracer.Start(root.TraceID(), root.SpanID(), "sdk.catalog", map[string]string{"profile": profile})
			offers, e := c.ListOfferings(ctx, profile)
			s.End(e)
			if e != nil {
				return "", e
			}
			rows := make([]string, 0, len(offers))
			for _, o := range offers {
				rows = append(rows, fmt.Sprintf("%s — %s (%s) meters=%v certs=%v", o.ID, o.Project, o.Category, o.Meters, o.Certifications))
			}
			return list(rows), nil
		case "pricebook":
			s := tracer.Start(root.TraceID(), root.SpanID(), "sdk.pricebook", nil)
			pb, e := c.PriceBook(ctx)
			s.End(e)
			if e != nil {
				return "", e
			}
			meters := make([]string, 0, len(pb.Prices))
			for m := range pb.Prices {
				meters = append(meters, m)
			}
			return "price book " + pb.Version + ": " + list(meters), nil
		case "frameworks":
			s := tracer.Start(root.TraceID(), root.SpanID(), "sdk.frameworks", nil)
			fw, e := c.Frameworks(ctx)
			s.End(e)
			if e != nil {
				return "", e
			}
			rows := make([]string, 0, len(fw))
			for _, f := range fw {
				rows = append(rows, fmt.Sprintf("%s (%s) — %s", f.Framework, f.Category, f.Authority))
			}
			return list(rows), nil
		case "rate":
			s := tracer.Start(root.TraceID(), root.SpanID(), "sdk.rate", nil)
			resp, e := c.Rate(ctx, demoRate())
			s.End(e)
			if e != nil {
				return "", e
			}
			out := fmt.Sprintf("invoice total %.2f %s", resp.Invoice.Total, resp.Invoice.Currency)
			if resp.Settlement != nil {
				out += fmt.Sprintf("; gross-to-customer %.2f", resp.Settlement.GrossToCustomer)
			}
			if resp.Tax != nil {
				out += fmt.Sprintf("; tax %.2f; gross-incl-tax %.2f", resp.Tax.TaxTotal, resp.Tax.Gross)
			}
			return out, nil
		case "traces":
			return fmt.Sprintf("%d spans recorded — see the Flow Traces panel or /debug/traces.json (OTLP)", len(tracer.Recent(0))), nil
		default:
			return "", fmt.Errorf("unknown command %q (try: help)", cmd)
		}
	}()
	root.End(err)

	if err != nil {
		return bubble(msg, "error: "+err.Error(), root.TraceID(), true)
	}
	return bubble(msg, body, root.TraceID(), false)
}

func bubble(msg, resp, traceID string, isErr bool) string {
	cls := "ok"
	if isErr {
		cls = "err"
	}
	return fmt.Sprintf(
		`<div class="msg user">%s</div><div class="msg bot %s">%s<div class="trace">trace: <code>%s</code></div></div>`,
		html.EscapeString(msg), cls, html.EscapeString(resp), html.EscapeString(traceID))
}

func renderTraces(spans []observe.Span) string {
	if len(spans) == 0 {
		return `<em>no spans yet — run a chat command</em>`
	}
	var b strings.Builder
	b.WriteString(`<table><tr><th>trace</th><th>span</th><th>name</th><th>ms</th><th>status</th></tr>`)
	for i := len(spans) - 1; i >= 0; i-- {
		s := spans[i]
		st := s.Status
		if s.Error != "" {
			st += ": " + s.Error
		}
		fmt.Fprintf(&b, `<tr><td><code>%s</code></td><td><code>%s</code></td><td>%s</td><td>%.2f</td><td>%s</td></tr>`,
			html.EscapeString(short(s.TraceID)), html.EscapeString(short(s.SpanID)),
			html.EscapeString(s.Name), s.DurationMs, html.EscapeString(st))
	}
	b.WriteString(`</table>`)
	return b.String()
}

func short(s string) string {
	if len(s) > 8 {
		return s[:8]
	}
	return s
}

func list(items []string) string {
	if len(items) == 0 {
		return "(none)"
	}
	return strings.Join(items, "\n")
}

func renderAssets(store *asset.MemStore) string {
	assets := store.List("")
	if len(assets) == 0 {
		return `<em>no assets cataloged — click Discover</em>`
	}
	var b strings.Builder
	b.WriteString(`<table><tr><th>kind</th><th>name</th><th>source</th><th>status</th><th>tags</th></tr>`)
	for _, a := range assets {
		fmt.Fprintf(&b, `<tr><td><span class="pill">%s</span></td><td>%s</td><td><code>%s</code></td><td>%s</td><td>`,
			html.EscapeString(a.Kind), html.EscapeString(a.Name), html.EscapeString(a.Source), html.EscapeString(a.Status))
		for _, t := range a.Tags {
			fmt.Fprintf(&b, `<span class="pill">%s</span>`, html.EscapeString(t))
		}
		b.WriteString(`</td></tr>`)
	}
	b.WriteString(`</table>`)
	return b.String()
}

func operatingModes() []string {
	return []string{
		string(billing.ModeDirect), string(billing.ModeReseller), string(billing.ModeAgency),
		string(billing.ModeMarketplace), string(billing.ModeServiceProvider),
	}
}

func demoRate() api.RateRequest {
	return api.RateRequest{
		TenantID:     "demo",
		Jurisdiction: "EU-DE",
		Events: []metering.UsageEvent{
			{TenantID: "demo", Meter: "compute.vcpu.hour", Quantity: 1300},
			{TenantID: "demo", Meter: "ai.gpu.hour", Quantity: 120},
		},
		Partner: &billing.Partner{ID: "p1", Mode: billing.ModeReseller, Rate: 0.15},
	}
}

func helpText() string {
	return list([]string{
		"commands:",
		"  help                 this message",
		"  providers            list infrastructure providers",
		"  modes                list operating models",
		"  catalog [profile]    list catalog offerings",
		"  pricebook            show active price book meters",
		"  frameworks           list compliance frameworks",
		"  rate                 run a demo pay-as-you-go + reseller + VAT rating",
		"  traces               trace recording status",
	})
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

var pageTemplate = ui.Head("Unboxd Platform — Admin Control Panel") + `
<body>
<header><h1>Unboxd Platform — Administrator Control Panel</h1></header>
<main>
` + ui.ChatHome("/chat") + `

  <div class="card"><h2>Infrastructure providers</h2>
    {{range .Providers}}<span class="pill">{{.}}</span>{{else}}<em>none</em>{{end}}</div>
  <div class="card"><h2>Operating models</h2>
    {{range .Modes}}<span class="pill">{{.}}</span>{{end}}</div>

  <div class="card span2"><h2>k3s cluster snapshots (S3) — platform default</h2>
    <p>Snapshots of the k3s cluster (etcd/datastore) are stored in S3 for disaster recovery. Set the platform-wide default here; orgs may override on their console.</p>
    <p>Bucket: <span class="pill">{{if .S3.Bucket}}{{.S3.Bucket}}{{else}}unset{{end}}</span>
       Region: <span class="pill">{{if .S3.Region}}{{.S3.Region}}{{else}}unset{{end}}</span>
       {{if .S3.Schedule}}Schedule: <span class="pill">{{.S3.Schedule}}</span>{{end}}
       {{if .S3.RetentionDays}}Retention: <span class="pill">{{.S3.RetentionDays}}d</span>{{end}}</p>
    <form hx-post="/settings" hx-target="body" hx-swap="outerHTML">
      <input type="text" name="bucket" value="{{.S3.Bucket}}" placeholder="bucket" required>
      <input type="text" name="region" value="{{.S3.Region}}" placeholder="us-east-1" required>
      <input type="text" name="endpoint" value="{{.S3.Endpoint}}" placeholder="endpoint (S3-compatible, optional)">
      <input type="text" name="prefix" value="{{.S3.Prefix}}" placeholder="prefix (optional)">
      <input type="text" name="accessKeyId" value="{{.S3.AccessKeyID}}" placeholder="access key id">
      <input type="password" name="secretKey" value="{{.S3.SecretKey}}" placeholder="secret key">
      <input type="text" name="schedule" value="{{.S3.Schedule}}" placeholder="cron e.g. 0 */6 * * *">
      <input type="text" name="retentionDays" value="{{if .S3.RetentionDays}}{{.S3.RetentionDays}}{{end}}" placeholder="retention days">
      <button type="submit">Save S3 settings</button>
    </form>
  </div>

  <div class="card span2"><h2>Asset catalog (IT admin)
      &nbsp;<button hx-post="/assets/discover" hx-target="#asset-table" hx-swap="innerHTML" style="font-size:11px;padding:4px 8px">Discover</button></h2>
    <div id="asset-table" hx-get="/assets" hx-trigger="load">loading…</div>
    <form hx-post="/assets/tag" hx-target="#asset-table" hx-swap="innerHTML" style="margin-top:8px">
      <input type="text" name="id" placeholder="asset id (e.g. service:catalog)">
      <input type="text" name="tag" placeholder="tag">
      <button type="submit">Tag</button>
    </form>
  </div>

  <div class="card span2"><h2>Flow traces (APM)
      &nbsp;<a href="/debug/traces.json" target="_blank">export OTLP/JSON</a></h2>
    <div hx-get="/debug/traces" hx-trigger="load, every 3s">loading…</div></div>

  <div class="card span2"><h2>Embedded BI &amp; observability</h2>
    <p>Traces export (OTLP) to the OpenTelemetry Collector and on to CNCF tools:</p>
    {{if .BI.Grafana}}<a class="pill" href="{{.BI.Grafana}}" target="_blank">Grafana</a>{{end}}
    {{if .BI.Jaeger}}<a class="pill" href="{{.BI.Jaeger}}" target="_blank">Jaeger</a>{{end}}
    {{if .BI.Superset}}<a class="pill" href="{{.BI.Superset}}" target="_blank">Apache Superset</a>{{end}}
    {{if .BI.Otel}}<a class="pill" href="{{.BI.Otel}}" target="_blank">OTel Collector</a>{{end}}
    {{if not (or .BI.Grafana .BI.Jaeger .BI.Superset .BI.Otel)}}<em>set GRAFANA_URL / JAEGER_URL / SUPERSET_URL / OTEL_COLLECTOR_URL to embed links</em>{{end}}
  </div>
</main></body></html>`
