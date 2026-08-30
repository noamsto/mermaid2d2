package mermaid2d2

import (
	"strings"
	"testing"

	mermaid "github.com/sammcj/mermaid-check"
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
			name: "entities with keys and a labeled relationship",
			in: "CUSTOMER: {\n" +
				"  shape: sql_table\n" +
				"  id: int {constraint: primary_key}\n" +
				"  name: string\n" +
				"}\n" +
				"ORDER: {\n" +
				"  shape: sql_table\n" +
				"  id: int {constraint: primary_key}\n" +
				"  customer_id: int {constraint: foreign_key}\n" +
				"}\n" +
				"CUSTOMER -> ORDER: places {source-arrowhead: {label: 1}; target-arrowhead: {label: 0..N}}\n",
			want: "erDiagram\n" +
				"    CUSTOMER {\n" +
				"        int id PK\n" +
				"        string name\n" +
				"    }\n" +
				"    ORDER {\n" +
				"        int id PK\n" +
				"        int customer_id FK\n" +
				"    }\n" +
				"    CUSTOMER ||--o{ ORDER : places\n",
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

func TestErDiagramFromD2SanitizesIdentifiers(t *testing.T) {
	graph := compileD2(t, "customer: {shape: sql_table}\n\"order item\": {shape: sql_table}\ncustomer -> \"order item\"\n")
	got := erDiagramFromD2(graph)
	if _, err := mermaid.Parse(got); err != nil {
		t.Errorf("erDiagramFromD2 produced unparsable Mermaid: %v\n%s", err, got)
	}
}

func TestErLeftRightSymbolCardinalities(t *testing.T) {
	tests := []struct {
		label     string
		wantLeft  string
		wantRight string
	}{
		{"0..1", "|o", "o|"},
		{"1..N", "}|", "|{"},
	}
	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			if got := erLeftSymbol(tt.label); got != tt.wantLeft {
				t.Errorf("erLeftSymbol(%q) = %q, want %q", tt.label, got, tt.wantLeft)
			}
			if got := erRightSymbol(tt.label); got != tt.wantRight {
				t.Errorf("erRightSymbol(%q) = %q, want %q", tt.label, got, tt.wantRight)
			}
		})
	}
}

func TestErKeysMultipleConstraints(t *testing.T) {
	got := erKeys([]string{"primary_key", "foreign_key"})
	want := "PK,FK"
	if got != want {
		t.Errorf("erKeys([primary_key, foreign_key]) = %q, want %q", got, want)
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
