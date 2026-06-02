// Command graph serves the platform's model as an interactive graph — the
// homepage of the platform: every entity is a node, every relation an edge, the
// governed graph made visible. It loads the ADL model (platform.agent and the
// metamodels/blueprints) at startup, projects it to nodes and edges, and renders
// it with vis-network over the shared design system.
//
// FE: htmx-era, dependency-free server; vis-network from CDN (like htmx) for the
// force-directed view. The graph is built live from the same ADL runtime that
// validates the model — one source of truth.
package main

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"sort"
	"strings"

	platform "github.com/unboxd-cloud/platform"
	"github.com/unboxd-cloud/platform/internal/server"
	"github.com/unboxd-cloud/platform/internal/ui"
	"github.com/unboxd-cloud/platform/pkg/adl"
)

type node struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Group string `json:"group"`
	Color string `json:"color"`
	Title string `json:"title"`
}

type edge struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Label  string `json:"label"`
	Arrows string `json:"arrows"`
}

type graph struct {
	Nodes      []node `json:"nodes"`
	Edges      []edge `json:"edges"`
	Namespaces int    `json:"namespaces"`
}

// color maps a namespace to a stable HSL color.
func color(ns string) string {
	sum := md5.Sum([]byte(ns))
	h := binary.BigEndian.Uint32(sum[:4]) % 360
	return fmt.Sprintf("hsl(%d,65%%,55%%)", h)
}

// build loads every .agent file from fsys and projects entities→nodes,
// relations→edges.
func build(fsys fs.FS) graph {
	paths := []string{"platform.agent"}
	for _, pat := range []string{"metamodels/*.agent", "blueprints/*.agent"} {
		m, _ := fs.Glob(fsys, pat)
		sort.Strings(m)
		paths = append(paths, m...)
	}

	nodes := map[string]string{} // fqn -> namespace
	var edges []edge
	for _, p := range paths {
		src, err := fs.ReadFile(fsys, p)
		if err != nil {
			continue
		}
		model, _ := adl.Parse(string(src))
		if model == nil {
			continue
		}
		ns := p // overwritten by the file's Namespace declaration
		for _, d := range model.Declarations {
			switch v := d.(type) {
			case *adl.Namespace:
				ns = v.Name
			case *adl.Entity:
				nodes[ns+"."+v.Name] = ns
			case *adl.Relation:
				s := qualify(v.Source, ns)
				t := qualify(v.Target, ns)
				edges = append(edges, edge{From: s, To: t, Label: v.Name, Arrows: "to"})
				nodes[s] = nsOf(s)
				nodes[t] = nsOf(t)
			}
		}
	}

	g := graph{}
	seen := map[string]bool{}
	for fqn, ns := range nodes {
		g.Nodes = append(g.Nodes, node{ID: fqn, Label: short(fqn), Group: ns, Color: color(ns), Title: fqn})
		seen[ns] = true
	}
	sort.Slice(g.Nodes, func(i, j int) bool { return g.Nodes[i].ID < g.Nodes[j].ID })
	g.Edges = edges
	g.Namespaces = len(seen)
	return g
}

// qualify resolves a relation endpoint to a node id. Parse does not fill
// Resolved (validation does), so fall back to qualifying a bare same-namespace
// name with the file's namespace; an already-qualified name is used as-is.
func qualify(r adl.Reference, ns string) string {
	if r.Resolved != "" {
		return r.Resolved
	}
	if strings.Contains(r.Name, ".") {
		return r.Name
	}
	return ns + "." + r.Name
}

func nsOf(fqn string) string {
	if i := lastDot(fqn); i >= 0 {
		return fqn[:i]
	}
	return fqn
}

func short(fqn string) string {
	if i := lastDot(fqn); i >= 0 {
		return fqn[i+1:]
	}
	return fqn
}

func lastDot(s string) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '.' {
			return i
		}
	}
	return -1
}

func main() {
	// Default to the embedded model so the service runs anywhere (scratch image,
	// any cluster). Set MODEL_DIR to read .agent files from disk during dev.
	var fsys fs.FS = platform.Model
	if dir := os.Getenv("MODEL_DIR"); dir != "" {
		fsys = os.DirFS(dir)
	}
	g := build(fsys)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, page, len(g.Nodes), len(g.Edges), g.Namespaces)
	})
	mux.HandleFunc("/graph.json", func(w http.ResponseWriter, _ *http.Request) {
		server.JSON(w, http.StatusOK, g)
	})

	addr := envOr("GRAPH_ADDR", ":8087")
	fmt.Printf("model graph on %s — %d nodes, %d edges, %d namespaces\n", addr, len(g.Nodes), len(g.Edges), g.Namespaces)
	srv := server.New(addr, mux)
	_ = srv.ListenAndServe()
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

var page = ui.Head("Unboxd Platform — The Graph") + `
<body style="margin:0">
<div id="hdr" style="position:fixed;top:10px;left:14px;z-index:9;background:#161922cc;padding:8px 14px;border:1px solid #262b38;border-radius:8px">
  <b>Unboxd Platform — the governed graph</b> &nbsp;<span style="color:#8b93a7">%d nodes · %d edges · %d namespaces</span>
</div>
<div id="net" style="height:100vh"></div>
<script src="https://unpkg.com/vis-network/standalone/umd/vis-network.min.js"></script>
<script>
fetch('/graph.json').then(r=>r.json()).then(g=>{
  const nodes=new vis.DataSet(g.nodes), edges=new vis.DataSet(g.edges);
  new vis.Network(document.getElementById('net'),{nodes,edges},{
    nodes:{shape:'dot',size:10,font:{color:'#cdd',size:11}},
    edges:{color:{color:'#33405a',highlight:'#2f6df6'},font:{color:'#7f8aa3',size:9},smooth:{type:'continuous'},length:120},
    physics:{barnesHut:{gravitationalConstant:-4000,springLength:90},stabilization:{iterations:200}},
    interaction:{hover:true,tooltipDelay:120}});
});
</script>
</body></html>`
