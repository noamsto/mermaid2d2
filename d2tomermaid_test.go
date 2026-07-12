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

func TestMermaidToD2NotImplemented(t *testing.T) {
	if _, err := MermaidToD2("flowchart TD\n a --> b"); err != ErrNotImplemented {
		t.Fatalf("MermaidToD2 error = %v, want ErrNotImplemented", err)
	}
}
