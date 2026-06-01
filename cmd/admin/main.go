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
	"strings"
	"time"

	"github.com/unboxd-cloud/platform/internal/api"
	"github.com/unboxd-cloud/platform/internal/billing"
	"github.com/unboxd-cloud/platform/internal/metering"
	"github.com/unboxd-cloud/platform/internal/observe"
	"github.com/unboxd-cloud/platform/internal/provider"
	"github.com/unboxd-cloud/platform/internal/server"
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
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		_ = page.Execute(w, map[string]any{
			"Providers": provider.DefaultRegistry().Names(),
			"Modes":     operatingModes(),
			"BI":        bi,
		})
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

const pageTemplate = `<!doctype html>
<html lang="en"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Unboxd Platform — Admin Control Panel</title>
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
  #chat-log{height:280px;overflow:auto;display:flex;flex-direction:column;gap:8px;margin-bottom:10px}
  .msg{padding:8px 12px;border-radius:10px;max-width:90%;white-space:pre-wrap}
  .msg.user{align-self:flex-end;background:#2f6df6;color:#fff}
  .msg.bot{align-self:flex-start;background:#1c2230;border:1px solid #2a3346}
  .msg.bot.err{border-color:#5a2d2d;background:#2a1d1d;color:#f0bcbc}
  .trace{margin-top:6px;font-size:11px;color:#7f8aa3}
  form.chat{display:flex;gap:8px}
  input[type=text]{flex:1;background:#0f1320;border:1px solid #2a3346;color:#e6e6e6;border-radius:8px;padding:8px 12px}
  button{background:#2f6df6;color:#fff;border:0;border-radius:8px;padding:8px 14px;cursor:pointer}
  a{color:#7fb0ff}
  table{width:100%;border-collapse:collapse;font-size:12px}
  td,th{text-align:left;padding:5px 8px;border-bottom:1px solid #232838}
  code{color:#9ad}
</style></head>
<body>
<header><h1>Unboxd Platform — Administrator Control Panel</h1></header>
<main>
  <div class="card span2"><h2>Chat</h2>
    <div id="chat-log"><div class="msg bot">Type <code>help</code> to begin.</div></div>
    <form class="chat" hx-post="/chat" hx-target="#chat-log" hx-swap="beforeend"
          hx-on::after-request="this.reset();document.getElementById('chat-log').scrollTop=1e9">
      <input type="text" name="msg" placeholder="e.g. catalog developer, rate, frameworks" autocomplete="off" autofocus>
      <button type="submit">Send</button>
    </form>
  </div>

  <div class="card"><h2>Infrastructure providers</h2>
    {{range .Providers}}<span class="pill">{{.}}</span>{{else}}<em>none</em>{{end}}</div>
  <div class="card"><h2>Operating models</h2>
    {{range .Modes}}<span class="pill">{{.}}</span>{{end}}</div>

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
