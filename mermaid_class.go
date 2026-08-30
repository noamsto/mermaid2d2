package mermaid2d2

import (
	"fmt"
	"strings"

	"github.com/sammcj/mermaid-check/ast"
)

// classDiagramToD2 renders a Mermaid class diagram as a D2 script. Classes become
// `shape: class` nodes; relationships become connections whose arrowheads encode
// the UML relation type. Notes targeting a class render as a tooltip attribute on
// that class; standalone notes (no target class) render as their own floating
// node; comments are dropped.
func classDiagramToD2(d *ast.ClassDiagram) string {
	classNames := make(map[string]bool)
	for _, s := range d.Statements {
		if c, ok := s.(*ast.Class); ok {
			classNames[c.Name] = true
		}
	}

	var b strings.Builder
	noteCount := 0
	for _, s := range d.Statements {
		switch v := s.(type) {
		case *ast.Class:
			classNode(&b, v)
		case *ast.Relationship:
			classRelationship(&b, v)
		case *ast.ClassNote:
			classNote(&b, v, &noteCount, classNames)
		}
	}
	return b.String()
}

func classNode(b *strings.Builder, c *ast.Class) {
	if len(c.Members) == 0 {
		fmt.Fprintf(b, "%s: {shape: class}\n", c.Name)
		return
	}
	fmt.Fprintf(b, "%s: {\n", c.Name)
	b.WriteString("  shape: class\n")
	for i := range c.Members {
		fmt.Fprintf(b, "  %s\n", classMember(&c.Members[i]))
	}
	b.WriteString("}\n")
}

// classNote renders a note targeting a class as a `.tooltip` attribute on that
// class. A standalone note (ClassName == "", added in mermaid-check v0.1.4 for
// the `note "text"` form with no target) has nowhere to attach a tooltip, so it
// renders as its own floating node instead, numbered note_1, note_2, ... in
// source order (mirroring seqEmitter's note_%d numbering in mermaidtod2.go).
// classNames guards against colliding with a class that happens to be named
// note_1, note_2, etc., skipping ahead to the next free number instead.
func classNote(b *strings.Builder, n *ast.ClassNote, noteCount *int, classNames map[string]bool) {
	if n.ClassName == "" {
		var key string
		for {
			*noteCount++
			key = fmt.Sprintf("note_%d", *noteCount)
			if !classNames[key] {
				break
			}
		}
		fmt.Fprintf(b, "%s: %s\n", key, d2Label(n.Text))
		return
	}
	fmt.Fprintf(b, "%s.tooltip: %s\n", n.ClassName, d2Label(n.Text))
}

// classMember renders one field or method. The parser fills Name and Type from the
// first two tokens as written, so they are passed through positionally.
func classMember(m *ast.ClassMember) string {
	name := m.Name
	if m.IsMethod {
		name = fmt.Sprintf("%s(%s)", name, strings.Join(m.Parameters, ", "))
	}

	// d2compiler silently drops a member whose key ends in a classifier, so a
	// classified key is quoted; # needs no escaping inside the quotes.
	key := classVisibility(m.Visibility) + name
	if c := classClassifier(m); c != "" {
		key = `"` + m.Visibility + name + c + `"`
	}

	if m.Type == "" {
		return key
	}
	return fmt.Sprintf("%s: %s", key, m.Type)
}

// classClassifier renders Mermaid's trailing member classifier: $ for static,
// * for abstract. Mermaid allows at most one per member.
func classClassifier(m *ast.ClassMember) string {
	switch {
	case m.IsAbstract:
		return "*"
	case m.IsStatic:
		return "$"
	default:
		return ""
	}
}

// classVisibility escapes protected (#) so D2 does not read it as a comment. D2 has
// no package marker, so ~ is left in the name verbatim.
func classVisibility(v string) string {
	if v == "#" {
		return `\#`
	}
	return v
}

func classRelationship(b *strings.Builder, r *ast.Relationship) {
	fmt.Fprintf(b, "%s -> %s", r.From, r.To)
	attrs := classRelAttrs(r)
	switch {
	case r.Label != "" && attrs != "":
		fmt.Fprintf(b, ": %s {%s}", d2Label(r.Label), attrs)
	case r.Label != "":
		fmt.Fprintf(b, ": %s", d2Label(r.Label))
	case attrs != "":
		fmt.Fprintf(b, ": {%s}", attrs)
	}
	b.WriteByte('\n')
}

// classRelAttrs builds the connection attribute block mapping the UML relation type
// to D2 arrowhead shapes, plus any end cardinalities as arrowhead labels.
func classRelAttrs(r *ast.Relationship) string {
	var srcShape, tgtShape string
	var srcHollow, tgtHollow, dashed bool
	switch r.Type {
	case "inheritance":
		tgtShape, tgtHollow = "triangle", true
	case "realization":
		tgtShape, tgtHollow, dashed = "triangle", true, true
	case "composition":
		srcShape = "diamond"
	case "aggregation":
		srcShape, srcHollow = "diamond", true
	case "dependency":
		dashed = true
	}

	var attrs []string
	if a := arrowheadAttr("source", srcShape, srcHollow, endLabel(r.FromCardinality, r.FromMultiplicity)); a != "" {
		attrs = append(attrs, a)
	}
	if a := arrowheadAttr("target", tgtShape, tgtHollow, endLabel(r.ToCardinality, r.ToMultiplicity)); a != "" {
		attrs = append(attrs, a)
	}
	if dashed {
		attrs = append(attrs, "style.stroke-dash: 3")
	}
	return strings.Join(attrs, "; ")
}

func arrowheadAttr(end, shape string, hollow bool, label string) string {
	var parts []string
	if shape != "" {
		parts = append(parts, "shape: "+shape)
	}
	if hollow {
		parts = append(parts, "style.filled: false")
	}
	if label != "" {
		parts = append(parts, "label: "+label)
	}
	if len(parts) == 0 {
		return ""
	}
	return fmt.Sprintf("%s-arrowhead: {%s}", end, strings.Join(parts, "; "))
}

func endLabel(cardinality, multiplicity string) string {
	if cardinality != "" {
		return cardinality
	}
	return multiplicity
}
