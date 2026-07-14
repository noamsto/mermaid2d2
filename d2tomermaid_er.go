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
	return b.String()
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
