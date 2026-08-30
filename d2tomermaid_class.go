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
	e := &mermaidEmitter{ids: map[*d2graph.Object]string{}, used: map[string]bool{}}
	var b strings.Builder
	b.WriteString("classDiagram\n")
	for _, obj := range graph.Root.ChildrenArray {
		classBlock(&b, e.id(obj), obj.Class)
	}
	for _, edge := range graph.Edges {
		fmt.Fprintf(&b, "    %s\n", classRelationshipLine(e, edge))
	}
	return b.String()
}

// classRelationshipLine renders a D2 connection between two class objects as
// a Mermaid UML relationship: 'From ["fromLabel"] <symbol> ["toLabel"] To[: label]'.
func classRelationshipLine(e *mermaidEmitter, edge *d2graph.Edge) string {
	var b strings.Builder
	b.WriteString(e.id(edge.Src))
	b.WriteByte(' ')
	if lbl := arrowheadLabel(edge.SrcArrowhead); lbl != "" {
		fmt.Fprintf(&b, "%q ", lbl)
	}
	b.WriteString(classSymbol(classRelType(edge)))
	if lbl := arrowheadLabel(edge.DstArrowhead); lbl != "" {
		fmt.Fprintf(&b, " %q", lbl)
	}
	b.WriteByte(' ')
	b.WriteString(e.id(edge.Dst))
	if lbl := strings.TrimSpace(edge.Label.Value); lbl != "" {
		fmt.Fprintf(&b, " : %s", lbl)
	}
	return b.String()
}

// classRelType inverts classRelAttrs: reads a connection's arrowhead
// shape/hollowness and dash style back into the UML relation type it came
// from. Anything matching none of the specific shapes is a plain association.
func classRelType(e *d2graph.Edge) string {
	dashed := e.Style.StrokeDash != nil
	switch {
	case arrowheadShape(e.DstArrowhead) == "triangle" && arrowheadHollow(e.DstArrowhead) && dashed:
		return "realization"
	case arrowheadShape(e.DstArrowhead) == "triangle" && arrowheadHollow(e.DstArrowhead):
		return "inheritance"
	case arrowheadShape(e.SrcArrowhead) == "diamond" && arrowheadHollow(e.SrcArrowhead):
		return "aggregation"
	case arrowheadShape(e.SrcArrowhead) == "diamond":
		return "composition"
	case dashed:
		return "dependency"
	default:
		return "association"
	}
}

func classSymbol(relType string) string {
	switch relType {
	case "inheritance":
		return "--|>"
	case "realization":
		return "..|>"
	case "composition":
		return "*--"
	case "aggregation":
		return "o--"
	case "dependency":
		return "..>"
	default:
		return "-->"
	}
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
	name, classifier := cutClassifier(m.Name)
	if m.Return == "" || m.Return == "void" {
		return vis + name + classifier
	}
	return fmt.Sprintf("%s%s %s%s", vis, name, m.Return, classifier)
}

// cutClassifier peels a trailing member classifier off a name. Mermaid puts it
// at the very end of the member, past the return type, so a method that has one
// has to be reassembled rather than printed straight through.
func cutClassifier(name string) (string, string) {
	for _, c := range []string{"$", "*"} {
		if rest, ok := strings.CutSuffix(name, c); ok {
			return rest, c
		}
	}
	return name, ""
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
