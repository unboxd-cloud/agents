package adl

import (
	"fmt"
	"strings"
)

// validate performs semantic checks that the grammar alone cannot express:
// it builds the entity symbol table, resolves every [Entity] cross-reference,
// and reports duplicate entity definitions. Resolved references are annotated in
// place so downstream consumers (Go services and the TS tooling) see the binding.
func validate(model *Model) []Diagnostic {
	if model == nil {
		return nil
	}
	var diags []Diagnostic

	// Build the entity index, namespace-qualifying each entity by the most
	// recent 'namespace' declaration that precedes it.
	byQualified := map[string]*Entity{}
	bySimple := map[string][]*Entity{}
	currentNS := ""
	for _, d := range model.Declarations {
		switch n := d.(type) {
		case *Namespace:
			currentNS = n.Name
		case *Entity:
			n.qualified = n.Name
			if currentNS != "" {
				n.qualified = currentNS + "." + n.Name
			}
			if prev, ok := byQualified[n.qualified]; ok {
				diags = append(diags, Diagnostic{
					Severity: SeverityError,
					Message: fmt.Sprintf("duplicate entity %q (already defined at line %d)",
						n.qualified, prev.Pos.Line),
					Pos: n.Pos,
				})
				continue
			}
			byQualified[n.qualified] = n
			bySimple[n.Name] = append(bySimple[n.Name], n)
		}
	}

	resolve := func(ref *Reference, role string) {
		if ref == nil {
			return
		}
		if e, ok := byQualified[ref.Name]; ok {
			ref.Resolved = e.qualified
			return
		}
		// Fall back to matching the final segment against simple names.
		simple := ref.Name
		if i := strings.LastIndex(simple, "."); i >= 0 {
			simple = simple[i+1:]
		}
		if cands := bySimple[simple]; len(cands) == 1 {
			ref.Resolved = cands[0].qualified
			return
		} else if len(cands) > 1 {
			diags = append(diags, Diagnostic{
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("ambiguous %s reference %q matches %d entities; qualify it", role, ref.Name, len(cands)),
				Pos:      ref.Pos,
			})
			ref.Resolved = cands[0].qualified
			return
		}
		diags = append(diags, Diagnostic{
			Severity: SeverityError,
			Message:  fmt.Sprintf("unresolved %s reference to entity %q", role, ref.Name),
			Pos:      ref.Pos,
		})
	}

	resolveType := func(t *TypeRef, role string) {
		if t != nil && t.Ref != nil {
			resolve(t.Ref, role)
		}
	}

	for _, d := range model.Declarations {
		switch n := d.(type) {
		case *Entity:
			resolve(n.Super, "supertype")
			for i := range n.Fields {
				resolveType(&n.Fields[i].Type, "field type")
			}
		case *Relation:
			resolve(&n.Source, "relation source")
			resolve(&n.Target, "relation target")
			for i := range n.Fields {
				resolveType(&n.Fields[i].Type, "field type")
			}
		case *Mind:
			resolve(&n.Subject, "mind subject")
			for i := range n.Fields {
				resolveType(&n.Fields[i].Type, "field type")
			}
		case *Belief:
			resolve(&n.Subject, "belief subject")
		case *Function:
			for i := range n.Params {
				resolveType(&n.Params[i].Type, "parameter type")
			}
			resolveType(&n.Return, "return type")
		case *SurrealMlBinding:
			resolve(&n.Input, "surrealml input")
			resolve(&n.Output, "surrealml output")
		}
	}

	return diags
}
