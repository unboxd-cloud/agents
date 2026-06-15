// Command community is the platform's community-development surface: the tool
// that taps human intelligence to keep raising the bar. It exposes the three
// everyday gestures (input/suggestion/feedback) as a visible, social, peer-
// validated feed, and the governed innovation funnel — identify → explore →
// propose → validate → promote → mature → release — as a board.
//
// FE: htmx over stdlib html/template (see docs/ui.md) — no build step,
// dependency-free on the server. The model behind it is metamodels/community.agent.
package main

import (
	"fmt"
	"html"
	"html/template"
	"net/http"
	"os"
	"strings"

	"github.com/unboxd-cloud/platform/internal/community"
	"github.com/unboxd-cloud/platform/internal/server"
	"github.com/unboxd-cloud/platform/internal/ui"
)

func main() {
	store := community.NewStore()
	seed(store)

	page := template.Must(template.New("page").Parse(pageTemplate))
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		_ = page.Execute(w, map[string]any{"Tracks": store.Tracks()})
	})

	mux.HandleFunc("/board", func(w http.ResponseWriter, _ *http.Request) {
		writeHTML(w, renderBoard(store))
	})
	mux.HandleFunc("/feed", func(w http.ResponseWriter, _ *http.Request) {
		writeHTML(w, renderFeed(store))
	})

	mux.HandleFunc("/need", func(w http.ResponseWriter, r *http.Request) {
		_, _ = store.RaiseNeed(actor(r), r.FormValue("problem"), r.FormValue("context"),
			r.FormValue("priority"), r.FormValue("from"))
		writeHTML(w, renderBoard(store))
	})
	mux.HandleFunc("/explore", func(w http.ResponseWriter, r *http.Request) {
		sols := []community.Solution{}
		if n := strings.TrimSpace(r.FormValue("solution")); n != "" {
			sols = append(sols, community.Solution{
				Name: n, Source: r.FormValue("source"), Fit: r.FormValue("fit"),
				Reusable: r.FormValue("reusable") == "on",
			})
		}
		_, _ = store.Explore(r.FormValue("need"), r.FormValue("gap"), sols...)
		writeHTML(w, renderBoard(store))
	})
	mux.HandleFunc("/propose", func(w http.ResponseWriter, r *http.Request) {
		composed := splitCSV(r.FormValue("composedOf"))
		if _, err := store.Propose(r.FormValue("need"), actor(r), r.FormValue("summary"),
			composed, r.FormValue("composable") == "on"); err != nil {
			writeHTML(w, note(err.Error())+renderBoard(store))
			return
		}
		writeHTML(w, renderBoard(store))
	})
	mux.HandleFunc("/assess", func(w http.ResponseWriter, r *http.Request) {
		conf := 1.0
		if _, err := store.Assess(r.FormValue("idea"), actor(r),
			r.FormValue("verdict") == "agree", conf, r.FormValue("evidence")); err != nil {
			writeHTML(w, note(err.Error())+renderBoard(store))
			return
		}
		writeHTML(w, renderBoard(store))
	})
	mux.HandleFunc("/promote", func(w http.ResponseWriter, r *http.Request) {
		if _, err := store.Promote(r.FormValue("idea"), r.FormValue("track"), actor(r)); err != nil {
			writeHTML(w, note(err.Error())+renderBoard(store))
			return
		}
		writeHTML(w, renderBoard(store))
	})
	mux.HandleFunc("/advance", func(w http.ResponseWriter, r *http.Request) {
		if _, err := store.Advance(r.FormValue("idea"), actor(r)); err != nil {
			writeHTML(w, note(err.Error())+renderBoard(store))
			return
		}
		writeHTML(w, renderBoard(store))
	})
	mux.HandleFunc("/contribute", func(w http.ResponseWriter, r *http.Request) {
		_, _ = store.Contribute(community.Contribution{
			Kind: community.Kind(r.FormValue("kind")), By: actor(r),
			Target: r.FormValue("target"), Body: r.FormValue("body"), SeedOf: r.FormValue("seedOf"),
		})
		writeHTML(w, renderFeed(store))
	})

	// Export the community model as JSON (so other repos/tools can read it).
	mux.HandleFunc("/model.json", func(w http.ResponseWriter, _ *http.Request) {
		server.JSON(w, http.StatusOK, map[string]any{
			"needs": store.Needs(), "ideas": store.Ideas(""),
			"contributions": store.Contributions(), "tracks": store.Tracks(),
		})
	})

	// Assistant: answers from the community's own data.
	mux.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		msg := strings.TrimSpace(r.FormValue("msg"))
		writeHTML(w, chatBubble(msg, assist(store, msg)))
	})

	addr := envOr("COMMUNITY_ADDR", ":8086")
	srv := server.New(addr, mux)
	fmt.Printf("community development surface on %s\n", addr)
	_ = srv.ListenAndServe()
}

// assist answers from the community store (needs, ideas, gestures).
func assist(s *community.Store, msg string) string {
	cmd := ""
	if f := strings.Fields(strings.ToLower(msg)); len(f) > 0 {
		cmd = f[0]
	}
	switch cmd {
	case "", "help":
		return "I can help with: needs, ideas, feed, tracks. The funnel is identify → explore → propose → validate → promote → mature → release."
	case "needs":
		var b strings.Builder
		for _, n := range s.Needs() {
			b.WriteString(n.Problem + " [" + n.Stage + "]\n")
		}
		return orNone(b.String())
	case "ideas":
		var b strings.Builder
		for _, i := range s.Ideas("") {
			fmt.Fprintf(&b, "%s [%s, score %.1f]\n", i.Summary, i.State, i.Score)
		}
		return orNone(b.String())
	case "feed":
		var b strings.Builder
		for _, c := range s.Contributions() {
			b.WriteString(string(c.Kind) + ": " + c.Body + "\n")
		}
		return orNone(b.String())
	case "tracks":
		var b strings.Builder
		for _, t := range s.Tracks() {
			b.WriteString(t.Name + ": " + strings.Join(t.Stages, " → ") + "\n")
		}
		return orNone(b.String())
	default:
		return "ask me about needs, ideas, feed, or tracks."
	}
}

func orNone(s string) string {
	if strings.TrimSpace(s) == "" {
		return "(none)"
	}
	return s
}

func chatBubble(msg, resp string) string {
	return `<div class="msg user">` + html.EscapeString(msg) +
		`</div><div class="msg bot">` + html.EscapeString(resp) + `</div>`
}

// actor is the acting contributor; in a real deployment this comes from SSO.
func actor(r *http.Request) string {
	if v := strings.TrimSpace(r.FormValue("by")); v != "" {
		return v
	}
	return envOr("COMMUNITY_USER", "you")
}

func renderBoard(s *community.Store) string {
	needs := s.Needs()
	if len(needs) == 0 {
		return `<em>no needs yet — identify one above</em>`
	}
	var b strings.Builder
	for _, n := range needs {
		fmt.Fprintf(&b, `<div class="card span2"><h2>%s &nbsp;<span class="pill">%s</span></h2>`,
			html.EscapeString(n.Problem), html.EscapeString(n.Stage))
		fmt.Fprintf(&b, `<p>raised by <code>%s</code>`, html.EscapeString(n.By))
		if n.Priority != "" {
			fmt.Fprintf(&b, ` · priority <span class="pill">%s</span>`, html.EscapeString(n.Priority))
		}
		b.WriteString(`</p>`)

		// Exploration: existing solutions (prior art).
		if len(n.Solutions) > 0 || n.Gap != "" {
			b.WriteString(`<p>existing solutions: `)
			for _, sol := range n.Solutions {
				tag := sol.Name
				if sol.Reusable {
					tag += " ↺"
				}
				fmt.Fprintf(&b, `<span class="pill">%s</span>`, html.EscapeString(tag))
			}
			if n.Gap != "" {
				fmt.Fprintf(&b, ` — gap: %s`, html.EscapeString(n.Gap))
			}
			b.WriteString(`</p>`)
		}

		// Stage-appropriate action.
		if !n.Explored {
			fmt.Fprintf(&b, `<form hx-post="/explore" hx-target="#board" hx-swap="innerHTML">
              <input type="hidden" name="need" value="%s">
              <input type="text" name="solution" placeholder="existing solution (prior art)">
              <input type="text" name="source" placeholder="source">
              <input type="text" name="gap" placeholder="gap — why none fit">
              <label><input type="checkbox" name="reusable"> reusable</label>
              <button type="submit">Explore</button></form>`, n.ID)
		} else {
			fmt.Fprintf(&b, `<form hx-post="/propose" hx-target="#board" hx-swap="innerHTML">
              <input type="hidden" name="need" value="%s">
              <input type="text" name="summary" placeholder="propose a new idea (a new composition)" required>
              <input type="text" name="composedOf" placeholder="composed of (comma-sep)">
              <label><input type="checkbox" name="composable" checked> composable</label>
              <button type="submit">Propose</button></form>`, n.ID)
		}

		// Ideas under this need.
		for _, idea := range s.Ideas(n.ID) {
			renderIdea(&b, s, idea)
		}
		b.WriteString(`</div>`)
	}
	return b.String()
}

func renderIdea(b *strings.Builder, s *community.Store, idea community.Idea) {
	fmt.Fprintf(b, `<div style="border-left:2px solid #2f3750;padding-left:12px;margin:10px 0">
      <strong>%s</strong> &nbsp;<span class="pill">%s</span> <span class="pill">score %.1f</span>`,
		html.EscapeString(idea.Summary), html.EscapeString(idea.State), idea.Score)
	if idea.Track != "" {
		fmt.Fprintf(b, ` <span class="pill">%s: %s</span>`, html.EscapeString(idea.Track), html.EscapeString(idea.Stage))
	}
	if len(idea.ComposedOf) > 0 {
		fmt.Fprintf(b, ` <em>composes %s</em>`, html.EscapeString(strings.Join(idea.ComposedOf, " + ")))
	}

	switch idea.State {
	case community.Proposed:
		fmt.Fprintf(b, `<form hx-post="/assess" hx-target="#board" hx-swap="innerHTML" style="margin-top:6px">
          <input type="hidden" name="idea" value="%s">
          <input type="text" name="evidence" placeholder="evidence">
          <button type="submit" name="verdict" value="agree">Validate</button>
          <button type="submit" name="verdict" value="refute" style="background:#5a2d2d">Refute</button></form>`, idea.ID)
	case community.Validated:
		b.WriteString(`<form hx-post="/promote" hx-target="#board" hx-swap="innerHTML" style="margin-top:6px">
          <input type="hidden" name="idea" value="` + idea.ID + `">
          <select name="track">`)
		for _, t := range s.Tracks() {
			fmt.Fprintf(b, `<option value="%s">%s</option>`, t.Name, t.Name)
		}
		b.WriteString(`</select><button type="submit">Promote</button></form>`)
	case community.Maturing:
		fmt.Fprintf(b, `<form hx-post="/advance" hx-target="#board" hx-swap="innerHTML" style="margin-top:6px">
          <input type="hidden" name="idea" value="%s">
          <button type="submit">Advance →</button></form>`, idea.ID)
	case community.Released:
		b.WriteString(`<p style="margin-top:6px">✓ released</p>`)
	}
	b.WriteString(`</div>`)
}

func renderFeed(s *community.Store) string {
	cs := s.Contributions()
	if len(cs) == 0 {
		return `<em>no gestures yet — add input, a suggestion, or feedback</em>`
	}
	var b strings.Builder
	b.WriteString(`<table><tr><th>kind</th><th>by</th><th>on</th><th>says</th><th>seeds</th></tr>`)
	for _, c := range cs {
		fmt.Fprintf(&b, `<tr><td><span class="pill">%s</span></td><td><code>%s</code></td><td>%s</td><td>%s</td><td>%s</td></tr>`,
			html.EscapeString(string(c.Kind)), html.EscapeString(c.By), html.EscapeString(c.Target),
			html.EscapeString(c.Body), html.EscapeString(shortID(c.SeedOf)))
	}
	b.WriteString(`</table>`)
	return b.String()
}

func shortID(s string) string {
	if s == "" {
		return "—"
	}
	return s
}

func note(msg string) string {
	return `<div class="card span2 note">` + html.EscapeString(msg) + `</div>`
}

func writeHTML(w http.ResponseWriter, s string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(s))
}

func splitCSV(s string) []string {
	var out []string
	for _, p := range strings.FieldsFunc(s, func(r rune) bool { return r == ',' }) {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

// seed gives the board a worked example: a real need, prior art, a composable
// idea, and a couple of trusted reviewers.
func seed(s *community.Store) {
	_, _ = s.AddContributor(community.Contributor{Subject: "sme.sre", Expertise: "sre", Trust: 1.5})
	_, _ = s.AddContributor(community.Contributor{Subject: "sme.ml", Expertise: "ml", Trust: 1.5})
	first, _ := s.Contribute(community.Contribution{Kind: community.Feedback, By: "sme.ml",
		Target: "model-router", Body: "yesterday's optimal model is not today's — prices and quality drift"})
	n, _ := s.RaiseNeed("sme.ml", "Pick the cost-optimal model daily under non-stationarity",
		"prices change, models evolve, inflation", "high", first.ID)
	_, _ = s.Explore(n.ID, "good primitives, none cost-aware for our dataset out of the box",
		community.Solution{Name: "LiteLLM", Source: "oss", Fit: "routing only", Reusable: true},
		community.Solution{Name: "contextual-bandit", Source: "literature", Fit: "online selection", Reusable: true})
}

var pageTemplate = ui.Head("Unboxd Platform — Community Development") + `
<body>
<header><h1>Community Development — tap human intelligence, raise the bar</h1></header>
<main>
` + ui.ChatHome("/chat") + `
  <div class="card span2"><h2>Innovation funnel — identify → explore → propose → validate → promote → mature → release</h2>
    <p>The machine owns excellence; humans own innovation. Significant gestures escalate into this governed funnel. You may not propose before exploring prior art, and an idea counts only if it is composable.</p>
    <form hx-post="/need" hx-target="#board" hx-swap="innerHTML">
      <input type="text" name="problem" placeholder="identify a need — the problem statement" required>
      <input type="text" name="context" placeholder="context">
      <input type="text" name="priority" placeholder="priority">
      <button type="submit">Identify need</button>
    </form>
  </div>
  <div id="board" class="span2" style="display:contents" hx-get="/board" hx-trigger="load">loading…</div>

  <div class="card span2"><h2>Gestures — input · suggestion · feedback (visible &amp; social)</h2>
    <p>Inline, near-zero friction. One person's suggestion seeds another's idea.</p>
    <form hx-post="/contribute" hx-target="#feed" hx-swap="innerHTML">
      <select name="kind"><option value="input">input</option><option value="suggestion">suggestion</option><option value="feedback">feedback</option></select>
      <input type="text" name="target" placeholder="on (decision, scorecard, agent…)">
      <input type="text" name="body" placeholder="say it…" required>
      <input type="text" name="seedOf" placeholder="seeds contribution id (optional)">
      <button type="submit">Post</button>
    </form>
    <div id="feed" hx-get="/feed" hx-trigger="load" style="margin-top:8px">loading…</div>
  </div>
</main></body></html>`
