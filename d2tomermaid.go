package mermaid2d2

import (
	"fmt"
	"strings"

	"oss.terrastruct.com/d2/d2compiler"
	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2target"
)

// D2ToMermaid parses D2 source and emits the Mermaid diagram it maps to.
//
// The Mermaid diagram type is inferred from the shapes present: sql_table
// shapes emit an erDiagram, class shapes emit a classDiagram, and anything
// else emits a flowchart (containers become subgraphs, connections become
// edges, classes:/class: styling becomes classDef/class, and the board
// direction sets orientation). A graph that mixes sql_table or class shapes
// with each other or with plain nodes/containers has no single Mermaid
// diagram type and returns an error. D2 features with no Mermaid equivalent
// in the chosen diagram type (grids, and — outside flowcharts — arrows,
// styling, containers) are dropped.
func D2ToMermaid(src string) (string, error) {
	graph, _, err := d2compiler.Compile("", strings.NewReader(src), nil)
	if err != nil {
		return "", fmt.Errorf("mermaid2d2: parse d2: %w", err)
	}

	kind, err := detectMermaidKind(graph)
	if err != nil {
		return "", err
	}
	switch kind {
	case kindER:
		return erDiagramFromD2(graph), nil
	case kindClass:
		return classDiagramFromD2(graph), nil
	default:
		return flowchartFromD2(graph), nil
	}
}

type mermaidKind int

const (
	kindFlowchart mermaidKind = iota
	kindER
	kindClass
)

// detectMermaidKind scans every object in the graph and decides which single
// Mermaid diagram type it maps to. sql_table and class objects only ever
// appear at the board root with no other shape alongside them — mirroring
// exactly what erDiagramToD2/classDiagramToD2 produce on the Mermaid→D2 side —
// so any other combination is rejected rather than guessed at.
func detectMermaidKind(graph *d2graph.Graph) (mermaidKind, error) {
	var hasTable, hasClass, hasOther bool
	for _, obj := range graph.Objects {
		switch strings.ToLower(obj.Shape.Value) {
		case d2target.ShapeSQLTable:
			hasTable = true
		case d2target.ShapeClass:
			hasClass = true
		default:
			hasOther = true
		}
	}
	switch {
	case hasTable && hasClass:
		return 0, fmt.Errorf("mermaid2d2: cannot emit a single Mermaid diagram: graph mixes sql_table and class shapes")
	case hasTable && hasOther:
		return 0, fmt.Errorf("mermaid2d2: cannot emit erDiagram: graph mixes sql_table shapes with plain nodes or containers")
	case hasClass && hasOther:
		return 0, fmt.Errorf("mermaid2d2: cannot emit classDiagram: graph mixes class shapes with plain nodes or containers")
	case hasTable:
		return kindER, nil
	case hasClass:
		return kindClass, nil
	default:
		return kindFlowchart, nil
	}
}

// flowchartFromD2 renders a D2 graph with no sql_table/class shapes as a
// Mermaid flowchart.
func flowchartFromD2(graph *d2graph.Graph) string {
	e := &mermaidEmitter{ids: map[*d2graph.Object]string{}, used: map[string]bool{}}
	var b strings.Builder
	fmt.Fprintf(&b, "flowchart %s\n", mermaidDirection(graph.Root))
	for _, obj := range graph.Root.ChildrenArray {
		e.writeObject(&b, obj, 1)
	}
	for _, edge := range graph.Edges {
		e.writeEdge(&b, edge)
	}
	e.writeStyling(&b, graph)
	return b.String()
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

// erIDs assigns Mermaid-safe, all-uppercase entity/relationship identifiers
// (erDiagram's grammar accepts only [A-Z_][A-Z0-9_-]*, stricter than
// flowchart/classDiagram's case-insensitive word-character identifiers), with
// collision suffixing if uppercasing merges two distinct sanitized ids.
type erIDs struct {
	ids  map[*d2graph.Object]string
	used map[string]bool
}

func (e *erIDs) id(obj *d2graph.Object) string {
	if id, ok := e.ids[obj]; ok {
		return id
	}
	id := strings.ToUpper(sanitizeID(obj.AbsID()))
	for e.used[id] {
		id += "_"
	}
	e.used[id] = true
	e.ids[obj] = id
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
