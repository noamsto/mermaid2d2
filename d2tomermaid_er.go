package mermaid2d2

import (
	"fmt"
	"strings"

	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2target"
)

// erDiagramFromD2 renders a D2 graph made entirely of sql_table shapes as a
// Mermaid erDiagram. Callers (D2ToMermaid, via detectMermaidKind) guarantee
// every object in graph.Root.ChildrenArray has shape: sql_table.
func erDiagramFromD2(graph *d2graph.Graph) string {
	var b strings.Builder
	b.WriteString("erDiagram\n")
	for _, obj := range graph.Root.ChildrenArray {
		erEntityBlock(&b, obj.IDVal, obj.SQLTable)
	}
	for _, e := range graph.Edges {
		fmt.Fprintf(&b, "    %s\n", erRelationshipLine(e))
	}
	return b.String()
}

// erRelationshipLine renders a D2 connection between two sql_table objects as
// a Mermaid crow's-foot relationship: "From <left>--<right> To[: label]".
func erRelationshipLine(e *d2graph.Edge) string {
	left := erLeftSymbol(arrowheadLabel(e.SrcArrowhead))
	right := erRightSymbol(arrowheadLabel(e.DstArrowhead))
	line := fmt.Sprintf("%s %s--%s %s", e.Src.IDVal, left, right, e.Dst.IDVal)
	if lbl := strings.TrimSpace(e.Label.Value); lbl != "" {
		line += " : " + lbl
	}
	return line
}

// erLeftSymbol maps a source-arrowhead cardinality label onto the crow's-foot
// symbol touching the left (source) entity. Missing/unrecognized labels
// default to "one" ("||"), matching Mermaid's mandatory-cardinality grammar.
func erLeftSymbol(label string) string {
	switch label {
	case "0..1":
		return "|o"
	case "1..N":
		return "}|"
	case "0..N":
		return "}o"
	default:
		return "||"
	}
}

// erRightSymbol is erLeftSymbol's mirror for the target (right) entity.
func erRightSymbol(label string) string {
	switch label {
	case "0..1":
		return "o|"
	case "1..N":
		return "|{"
	case "0..N":
		return "o{"
	default:
		return "||"
	}
}

func erEntityBlock(b *strings.Builder, id string, t *d2target.SQLTable) {
	if len(t.Columns) == 0 {
		fmt.Fprintf(b, "    %s\n", id)
		return
	}
	fmt.Fprintf(b, "    %s {\n", id)
	for _, c := range t.Columns {
		fmt.Fprintf(b, "        %s\n", erColumnLine(c))
	}
	b.WriteString("    }\n")
}

// erColumnLine renders one column as "type name KEY,KEY" (Mermaid requires
// comma-joined keys with no space, e.g. "PK,FK").
func erColumnLine(c d2target.SQLColumn) string {
	line := fmt.Sprintf("%s %s", c.Type.Label, c.Name.Label)
	if keys := erKeys(c.Constraint); keys != "" {
		line += " " + keys
	}
	return line
}

// erKeys maps D2 column constraints onto Mermaid PK/FK/UK key tokens,
// dropping any unrecognized constraint.
func erKeys(constraints []string) string {
	abbr := map[string]string{"primary_key": "PK", "foreign_key": "FK", "unique": "UK"}
	var keys []string
	for _, c := range constraints {
		if k, ok := abbr[c]; ok {
			keys = append(keys, k)
		}
	}
	return strings.Join(keys, ",")
}
