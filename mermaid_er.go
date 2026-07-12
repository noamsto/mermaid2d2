package mermaid2d2

import (
	"fmt"
	"strings"

	"github.com/sammcj/mermaid-check/ast"
)

// erDiagramToD2 renders a Mermaid entity-relationship diagram as D2. Entities
// become sql_table shapes (attributes are typed columns; PK/FK/UK keys become
// D2 column constraints) and relationships become connections carrying the
// relationship label and end cardinalities.
func erDiagramToD2(d *ast.ERDiagram) string {
	var b strings.Builder
	if dir := d2Direction(d.Direction); dir != "" {
		fmt.Fprintf(&b, "direction: %s\n", dir)
	}
	for i := range d.Entities {
		erEntity(&b, &d.Entities[i])
	}
	for i := range d.Relationships {
		erRelationship(&b, &d.Relationships[i])
	}
	return b.String()
}

func erEntity(b *strings.Builder, e *ast.EREntity) {
	if len(e.Attributes) == 0 {
		fmt.Fprintf(b, "%s: {shape: sql_table}\n", e.Name)
		return
	}
	fmt.Fprintf(b, "%s: {\n", e.Name)
	b.WriteString("  shape: sql_table\n")
	for _, a := range e.Attributes {
		line := fmt.Sprintf("  %s: %s", a.Name, a.Type)
		if c := erConstraint(a.Keys); c != "" {
			line += " {constraint: " + c + "}"
		}
		fmt.Fprintln(b, line)
	}
	b.WriteString("}\n")
}

// erConstraint maps Mermaid attribute keys onto D2 column constraints, using the
// D2 array form when an attribute carries more than one (e.g. a PK that is also
// an FK). Unknown keys are ignored.
func erConstraint(keys []string) string {
	var cs []string
	for _, k := range keys {
		switch k {
		case "PK":
			cs = append(cs, "primary_key")
		case "FK":
			cs = append(cs, "foreign_key")
		case "UK":
			cs = append(cs, "unique")
		}
	}
	switch len(cs) {
	case 0:
		return ""
	case 1:
		return cs[0]
	default:
		return "[" + strings.Join(cs, "; ") + "]"
	}
}

func erRelationship(b *strings.Builder, r *ast.ERRelationship) {
	fmt.Fprintf(b, "%s -> %s", r.From, r.To)
	var attrs []string
	if a := arrowheadAttr("source", "", false, erCardinality(r.FromCard)); a != "" {
		attrs = append(attrs, a)
	}
	if a := arrowheadAttr("target", "", false, erCardinality(r.ToCard)); a != "" {
		attrs = append(attrs, a)
	}
	label := strings.TrimSpace(r.Label)
	attr := strings.Join(attrs, "; ")
	switch {
	case label != "" && attr != "":
		fmt.Fprintf(b, ": %s {%s}", label, attr)
	case label != "":
		fmt.Fprintf(b, ": %s", label)
	case attr != "":
		fmt.Fprintf(b, ": {%s}", attr)
	}
	b.WriteByte('\n')
}

// erCardinality maps a Mermaid crow's-foot cardinality code onto a human label
// used as an arrowhead label. Both orientations of each code are accepted (the
// left end reads `}o`, the right end `o{`); unknown codes yield "".
func erCardinality(card string) string {
	switch card {
	case "||":
		return "1"
	case "|o", "o|":
		return "0..1"
	case "}|", "|{":
		return "1..N"
	case "}o", "o{":
		return "0..N"
	default:
		return ""
	}
}
