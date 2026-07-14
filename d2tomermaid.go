package mermaid2d2

import (
	"fmt"
	"strings"

	"oss.terrastruct.com/d2/d2compiler"
	"oss.terrastruct.com/d2/d2graph"
)

// D2ToMermaid parses D2 source and emits an equivalent Mermaid flowchart.
//
// Containers become subgraphs, connections become edges, and the board
// direction sets the flowchart orientation. D2 features without a flowchart
// equivalent (SQL tables, class shapes, grids, styling) are dropped.
func D2ToMermaid(src string) (string, error) {
	graph, _, err := d2compiler.Compile("", strings.NewReader(src), nil)
	if err != nil {
		return "", fmt.Errorf("mermaid2d2: parse d2: %w", err)
	}

	e := &mermaidEmitter{ids: map[*d2graph.Object]string{}, used: map[string]bool{}}
	var b strings.Builder
	fmt.Fprintf(&b, "flowchart %s\n", mermaidDirection(graph.Root))
	for _, obj := range graph.Root.ChildrenArray {
		e.writeObject(&b, obj, 1)
	}
	for _, edge := range graph.Edges {
		e.writeEdge(&b, edge)
	}
	return b.String(), nil
}

type mermaidEmitter struct {
	ids  map[*d2graph.Object]string
	used map[string]bool
}

// id returns a stable, unique, Mermaid-safe identifier for obj.
func (e *mermaidEmitter) id(obj *d2graph.Object) string {
	if id, ok := e.ids[obj]; ok {
		return id
	}
	id := sanitizeID(obj.AbsID())
	// AbsID is unique, so a collision here means sanitization merged two ids.
	for e.used[id] {
		id += "_"
	}
	e.used[id] = true
	e.ids[obj] = id
	return id
}

func (e *mermaidEmitter) writeObject(b *strings.Builder, obj *d2graph.Object, depth int) {
	indent := strings.Repeat("    ", depth)
	label := mermaidLabel(obj.Label.Value, obj.ID)
	if obj.IsContainer() {
		fmt.Fprintf(b, "%ssubgraph %s[\"%s\"]\n", indent, e.id(obj), label)
		for _, child := range obj.ChildrenArray {
			e.writeObject(b, child, depth+1)
		}
		fmt.Fprintf(b, "%send\n", indent)
		return
	}
	fmt.Fprintf(b, "%s%s[\"%s\"]\n", indent, e.id(obj), label)
}

func (e *mermaidEmitter) writeEdge(b *strings.Builder, edge *d2graph.Edge) {
	src, dst := edge.Src, edge.Dst
	arrow := "-->"
	switch {
	case edge.SrcArrow && edge.DstArrow:
		arrow = "<-->"
	case edge.SrcArrow && !edge.DstArrow:
		// D2 "a <- b" points at the source; Mermaid has no left arrow, so flip.
		src, dst = dst, src
	case !edge.SrcArrow && !edge.DstArrow:
		arrow = "---"
	}
	if label := strings.TrimSpace(edge.Label.Value); label != "" {
		fmt.Fprintf(b, "    %s %s|\"%s\"| %s\n", e.id(src), arrow, mermaidText(label), e.id(dst))
		return
	}
	fmt.Fprintf(b, "    %s %s %s\n", e.id(src), arrow, e.id(dst))
}

func mermaidDirection(root *d2graph.Object) string {
	switch root.Direction.Value {
	case "up":
		return "BT"
	case "right":
		return "LR"
	case "left":
		return "RL"
	default: // "" and "down"
		return "TD"
	}
}

func mermaidLabel(label, fallback string) string {
	if strings.TrimSpace(label) == "" {
		label = fallback
	}
	return mermaidText(label)
}

// mermaidText escapes a string for use inside a Mermaid double-quoted label.
func mermaidText(s string) string {
	s = strings.ReplaceAll(s, "\"", "#quot;")
	s = strings.ReplaceAll(s, "\n", "<br/>")
	return s
}

// sanitizeID reduces a D2 AbsID to the [A-Za-z0-9_] set Mermaid accepts for
// node identifiers, ensuring the result never starts with a digit.
func sanitizeID(absID string) string {
	var b strings.Builder
	for _, r := range absID {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	id := b.String()
	if id == "" {
		return "n"
	}
	if id[0] >= '0' && id[0] <= '9' {
		return "n" + id
	}
	return id
}

// arrowheadLabel returns the label carried by a connection's source/target
// arrowhead attributes (D2's source-arrowhead/target-arrowhead), or "" if the
// arrowhead has no label or is absent entirely.
func arrowheadLabel(a *d2graph.Attributes) string {
	if a == nil {
		return ""
	}
	return strings.TrimSpace(a.Label.Value)
}

// arrowheadShape returns the D2 shape keyword set on a connection's
// source/target arrowhead (e.g. "triangle", "diamond"), or "" if unset.
func arrowheadShape(a *d2graph.Attributes) string {
	if a == nil {
		return ""
	}
	return a.Shape.Value
}

// arrowheadHollow reports whether an arrowhead is drawn unfilled
// (style.filled: false — used for inheritance/aggregation's open shapes).
func arrowheadHollow(a *d2graph.Attributes) bool {
	return a != nil && a.Style.Filled != nil && a.Style.Filled.Value == "false"
}
