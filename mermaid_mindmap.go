package mermaid2d2

import (
	"fmt"
	"strings"

	"github.com/sammcj/mermaid-check/ast"
)

// mindmapToD2 renders a Mermaid mindmap as a D2 tree. Mindmap nodes have no ids
// of their own, so each is given a stable generated id (n1, n2, … in depth-first
// order) labeled with its text, and connected to its parent. Node shapes with a
// D2 equivalent are mapped; icons have no plain equivalent and are dropped.
func mindmapToD2(d *ast.MindmapDiagram) string {
	if d.Root == nil {
		return ""
	}
	var b strings.Builder
	e := &mindmapEmitter{}
	e.walk(&b, d.Root, "")
	return b.String()
}

type mindmapEmitter struct{ n int }

// walk declares a node, connects it to its parent (unless it is the root), and
// recurses into its children. Parents are declared before children, so the
// parent id in each connection already exists.
func (e *mindmapEmitter) walk(b *strings.Builder, node *ast.MindmapNode, parentID string) {
	e.n++
	id := fmt.Sprintf("n%d", e.n)

	label := d2Label(strings.TrimSpace(node.Text))
	shape := mindmapShape(node.Shape)
	switch {
	case label != "" && shape != "":
		fmt.Fprintf(b, "%s: %s {shape: %s}\n", id, label, shape)
	case shape != "":
		fmt.Fprintf(b, "%s: {shape: %s}\n", id, shape)
	case label != "":
		fmt.Fprintf(b, "%s: %s\n", id, label)
	default:
		fmt.Fprintf(b, "%s\n", id)
	}

	if parentID != "" {
		fmt.Fprintf(b, "%s -> %s\n", parentID, id)
	}
	for _, c := range node.Children {
		e.walk(b, c, id)
	}
}

// mindmapShape maps a Mermaid mindmap node's bracket pair onto a D2 shape
// keyword, returning "" for shapes that render as the default rectangle.
func mindmapShape(bracket string) string {
	switch bracket {
	case "(())":
		return "circle"
	case "{{}}":
		return "hexagon"
	case "))((":
		return "cloud"
	default:
		return ""
	}
}
