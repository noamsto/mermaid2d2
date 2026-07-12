package mermaid2d2

import (
	"testing"
)

func TestMermaidToD2ER(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "entities with keys and a labeled relationship",
			in: "erDiagram\n" +
				"    CUSTOMER {\n" +
				"        int id PK\n" +
				"        string name\n" +
				"    }\n" +
				"    ORDER {\n" +
				"        int id PK\n" +
				"        int customer_id FK\n" +
				"    }\n" +
				"    CUSTOMER ||--o{ ORDER : places",
			want: "CUSTOMER: {\n" +
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
		},
		{
			name: "entity without attributes",
			in:   "erDiagram\n    A\n",
			want: "A: {shape: sql_table}\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertConvertsTo(t, tt.in, tt.want)
		})
	}
}
