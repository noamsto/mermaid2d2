package mermaid2d2

import (
	"strings"
	"testing"

	"oss.terrastruct.com/d2/d2compiler"
	"oss.terrastruct.com/d2/d2graph"
)

func TestErDiagramFromD2Entities(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "entities with keys",
			in: "CUSTOMER: {\n" +
				"  shape: sql_table\n" +
				"  id: int {constraint: primary_key}\n" +
				"  name: string\n" +
				"}\n" +
				"ORDER: {\n" +
				"  shape: sql_table\n" +
				"  id: int {constraint: primary_key}\n" +
				"  customer_id: int {constraint: foreign_key}\n" +
				"}\n",
			want: "erDiagram\n" +
				"    CUSTOMER {\n" +
				"        int id PK\n" +
				"        string name\n" +
				"    }\n" +
				"    ORDER {\n" +
				"        int id PK\n" +
				"        int customer_id FK\n" +
				"    }\n",
		},
		{
			name: "entity without attributes",
			in:   "A: {shape: sql_table}\n",
			want: "erDiagram\n    A\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			graph := compileD2(t, tt.in)
			got := erDiagramFromD2(graph)
			if got != tt.want {
				t.Errorf("erDiagramFromD2(%q)\n got:\n%s\nwant:\n%s", tt.in, got, tt.want)
			}
		})
	}
}

// compileD2 compiles D2 source and fails the test on error. Shared by the new
// d2tomermaid_*_test.go files.
func compileD2(t *testing.T, src string) *d2graph.Graph {
	t.Helper()
	graph, _, err := d2compiler.Compile("", strings.NewReader(src), nil)
	if err != nil {
		t.Fatalf("d2compiler.Compile(%q) error: %v", src, err)
	}
	return graph
}
