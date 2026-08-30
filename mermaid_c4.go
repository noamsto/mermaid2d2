package mermaid2d2

import (
	"fmt"
	"strings"

	"github.com/sammcj/mermaid-check/ast"
)

// c4DiagramToD2 renders any Mermaid C4 diagram (Context, Container, Component,
// Dynamic, Deployment — they share one AST shape) as D2. Elements become
// nodes (Person -> shape: person, _Db variants -> cylinder, _Queue variants ->
// queue, external elements -> a dashed border), boundaries become containers,
// and relationships become labeled connections resolved to their nodes'
// qualified paths. D2 has no native C4 notation, so the mapping is lossy: the
// diagram title, sprites, tags, links, relationship descriptions, generic
// Boundary's type parameter, and UpdateElementStyle/UpdateRelStyle overrides
// are dropped.
func c4DiagramToD2(d *ast.C4Diagram) string {
	var b strings.Builder
	paths := map[string]string{}
	for i := range d.Elements {
		c4ElementLine(&b, "", &d.Elements[i])
		paths[d.Elements[i].ID] = d.Elements[i].ID
	}
	for i := range d.Boundaries {
		c4BoundaryBlock(&b, &d.Boundaries[i], "", "", paths)
	}
	for i := range d.Relationships {
		c4RelationshipLine(&b, &d.Relationships[i], paths)
	}
	return b.String()
}

// c4BoundaryBlock emits a boundary as a D2 container and records the
// qualified path of the boundary itself and of every element nested inside
// it (directly or via a nested boundary), so relationships declared at the
// diagram level can reference any of them by path regardless of nesting
// depth — a Rel may target a boundary as a whole, not just a leaf element.
// paths is shared and mutated across the whole diagram, so a duplicate ID
// reused across scopes (malformed input) silently overwrites the earlier
// entry; last write wins.
func c4BoundaryBlock(b *strings.Builder, bd *ast.C4Boundary, indent, parentPath string, paths map[string]string) {
	path := bd.ID
	if parentPath != "" {
		path = parentPath + "." + bd.ID
	}
	paths[bd.ID] = path
	fmt.Fprintf(b, "%s%s: %s {\n", indent, bd.ID, d2Label(bd.Label))
	inner := indent + "  "
	for i := range bd.Elements {
		c4ElementLine(b, inner, &bd.Elements[i])
		paths[bd.Elements[i].ID] = path + "." + bd.Elements[i].ID
	}
	for i := range bd.Boundaries {
		c4BoundaryBlock(b, &bd.Boundaries[i], inner, path, paths)
	}
	fmt.Fprintf(b, "%s}\n", indent)
}

// c4ElementLine emits a C4 element (Person, System, Container, Component, or
// a leaf Node/Deployment_Node) as a D2 node. Technology is folded into the
// label (parenthesised, since D2 requires quoting for a bare "[" and this
// codebase's d2Label doesn't quote for it); Description becomes a tooltip.
func c4ElementLine(b *strings.Builder, indent string, e *ast.C4Element) {
	label := e.Label
	if tech := strings.TrimSpace(e.Technology); tech != "" {
		label += " (" + tech + ")"
	}

	var attrs []string
	if shape := c4ElementShape(e); shape != "" {
		attrs = append(attrs, "shape: "+shape)
	}
	if e.External {
		attrs = append(attrs, "style.stroke-dash: 3")
	}
	if desc := strings.TrimSpace(e.Description); desc != "" {
		attrs = append(attrs, "tooltip: "+d2Label(desc))
	}

	if len(attrs) == 0 {
		fmt.Fprintf(b, "%s%s: %s\n", indent, e.ID, d2Label(label))
		return
	}
	fmt.Fprintf(b, "%s%s: %s {%s}\n", indent, e.ID, d2Label(label), strings.Join(attrs, "; "))
}

// c4ElementShape maps a C4 element onto a D2 shape keyword, returning "" for
// elements that render as the default rectangle (System, Container, Component).
func c4ElementShape(e *ast.C4Element) string {
	switch {
	case e.ElementType == "Person":
		return "person"
	case e.Database:
		return "cylinder"
	case e.Queue:
		return "queue"
	case e.ElementType == "Node":
		return "package"
	default:
		return ""
	}
}

// c4RelationshipLine emits a C4 relationship as a D2 connection between the
// qualified paths of its endpoints. BiRel becomes a bidirectional connection;
// Rel_Back reverses the drawn arrow (Mermaid draws it from To back to From,
// keeping the endpoint order used by Rel_Neighbor/Rel_Down/etc., which are
// pure layout hints with no bearing on arrow direction).
func c4RelationshipLine(b *strings.Builder, r *ast.C4Relationship, paths map[string]string) {
	from, to := c4Path(paths, r.From), c4Path(paths, r.To)
	arrow := "->"
	switch r.RelType {
	case "BiRel":
		arrow = "<->"
	case "Rel_Back":
		from, to = to, from
	}

	label := r.Label
	if tech := strings.TrimSpace(r.Technology); tech != "" {
		label += " (" + tech + ")"
	}

	fmt.Fprintf(b, "%s %s %s", from, arrow, to)
	if label != "" {
		fmt.Fprintf(b, ": %s", d2Label(label))
	}
	b.WriteByte('\n')
}

// c4Path resolves a C4 element or boundary ID to its qualified D2 path,
// falling back to the bare ID if it wasn't declared anywhere (a malformed
// diagram — Rel referencing an unknown ID).
func c4Path(paths map[string]string, id string) string {
	if p, ok := paths[id]; ok {
		return p
	}
	return id
}
