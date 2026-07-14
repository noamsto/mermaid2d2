package mermaid2d2

import "testing"

func TestD2ToMermaid(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "simple edge",
			in:   "a -> b",
			want: "flowchart TD\n" +
				"    a[\"a\"]\n" +
				"    b[\"b\"]\n" +
				"    a --> b\n",
		},
		{
			name: "labels and edge label",
			in:   "x: Hello World\ny: Bye\nx -> y: go",
			want: "flowchart TD\n" +
				"    x[\"Hello World\"]\n" +
				"    y[\"Bye\"]\n" +
				"    x -->|\"go\"| y\n",
		},
		{
			name: "container becomes subgraph",
			in:   "server: {\n  api\n  db\n}\nserver.api -> server.db",
			want: "flowchart TD\n" +
				"    subgraph server[\"server\"]\n" +
				"        server_api[\"api\"]\n" +
				"        server_db[\"db\"]\n" +
				"    end\n" +
				"    server_api --> server_db\n",
		},
		{
			name: "direction right",
			in:   "direction: right\na -> b",
			want: "flowchart LR\n" +
				"    a[\"a\"]\n" +
				"    b[\"b\"]\n" +
				"    a --> b\n",
		},
		{
			name: "undirected connection",
			in:   "a -- b",
			want: "flowchart TD\n" +
				"    a[\"a\"]\n" +
				"    b[\"b\"]\n" +
				"    a --- b\n",
		},
		{
			name: "bidirectional arrow",
			in:   "a <-> b",
			want: "flowchart TD\n" +
				"    a[\"a\"]\n" +
				"    b[\"b\"]\n" +
				"    a <--> b\n",
		},
		{
			name: "reverse arrow is flipped",
			in:   "a <- b",
			want: "flowchart TD\n" +
				"    a[\"a\"]\n" +
				"    b[\"b\"]\n" +
				"    b --> a\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := D2ToMermaid(tt.in)
			if err != nil {
				t.Fatalf("D2ToMermaid(%q) error: %v", tt.in, err)
			}
			if got != tt.want {
				t.Errorf("D2ToMermaid(%q)\n got:\n%s\nwant:\n%s", tt.in, got, tt.want)
			}
		})
	}
}

func TestDetectMermaidKindMixedShapes(t *testing.T) {
	tests := []struct {
		name string
		in   string
	}{
		{name: "table and class", in: "A: {shape: sql_table}\nB: {shape: class}\n"},
		{name: "table and plain node", in: "A: {shape: sql_table}\nB\n"},
		{name: "class and plain node", in: "A: {shape: class}\nB\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			graph := compileD2(t, tt.in)
			if _, err := detectMermaidKind(graph); err == nil {
				t.Errorf("detectMermaidKind(%q) = nil error, want an error", tt.in)
			}
		})
	}
}

func TestD2ToMermaidErAndClass(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "er diagram",
			in:   "A: {shape: sql_table}\n",
			want: "erDiagram\n    A\n",
		},
		{
			name: "class diagram",
			in:   "Dog: {shape: class}\n",
			want: "classDiagram\n    class Dog\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := D2ToMermaid(tt.in)
			if err != nil {
				t.Fatalf("D2ToMermaid(%q) error: %v", tt.in, err)
			}
			if got != tt.want {
				t.Errorf("D2ToMermaid(%q)\n got:\n%s\nwant:\n%s", tt.in, got, tt.want)
			}
		})
	}
}
