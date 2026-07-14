package mermaid2d2

import (
	"fmt"
	"strings"

	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2target"
)

// classDiagramFromD2 renders a D2 graph made entirely of class shapes as a
// Mermaid classDiagram. Callers (D2ToMermaid, via detectMermaidKind) guarantee
// every object in graph.Root.ChildrenArray has shape: class.
func classDiagramFromD2(graph *d2graph.Graph) string {
	var b strings.Builder
	b.WriteString("classDiagram\n")
	for _, obj := range graph.Root.ChildrenArray {
		classBlock(&b, obj.IDVal, obj.Class)
	}
	return b.String()
}

func classBlock(b *strings.Builder, id string, c *d2target.Class) {
	if len(c.Fields) == 0 && len(c.Methods) == 0 {
		fmt.Fprintf(b, "    class %s\n", id)
		return
	}
	fmt.Fprintf(b, "    class %s {\n", id)
	for _, f := range c.Fields {
		fmt.Fprintf(b, "        %s\n", classFieldLine(f))
	}
	for _, m := range c.Methods {
		fmt.Fprintf(b, "        %s\n", classMethodLine(m))
	}
	b.WriteString("    }\n")
}

// classFieldLine renders "<vis><Type> <Name>" (Mermaid field syntax). D2 has
// no package-visibility (~) marker, so d2compiler leaves a leading "~" baked
// into the field name verbatim instead of a Visibility value — recover it as
// the vis token rather than double-prefixing.
func classFieldLine(f d2target.ClassField) string {
	name := f.Name
	vis := classVisibilityPrefix(f.Visibility)
	if rest, ok := strings.CutPrefix(name, "~"); ok {
		vis, name = "~", rest
	}
	if f.Type == "" {
		return vis + name
	}
	return fmt.Sprintf("%s%s %s", vis, f.Type, name)
}

// classMethodLine renders "<vis><Name>(...) <Return>". m.Name already
// contains the "(params)" suffix (d2compiler bakes it into the field id).
// Return == "void" means d2compiler saw no explicit return type, not a
// literal "void" return — omit the return type in that case.
func classMethodLine(m d2target.ClassMethod) string {
	vis := classVisibilityPrefix(m.Visibility)
	if m.Return == "" || m.Return == "void" {
		return vis + m.Name
	}
	return fmt.Sprintf("%s%s %s", vis, m.Name, m.Return)
}

func classVisibilityPrefix(v string) string {
	switch v {
	case "private":
		return "-"
	case "protected":
		return "#"
	default: // "public" and any unrecognized value
		return "+"
	}
}
