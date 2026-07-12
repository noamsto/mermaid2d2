package mermaid2d2

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"oss.terrastruct.com/d2/d2compiler"
)

func TestMermaidToD2Flowchart(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "simple edge",
			in:   "flowchart TD\n    a --> b",
			want: "a -> b\n",
		},
		{
			name: "labels and edge label",
			in:   "flowchart LR\n    A[Start] --> B[End]\n    B -->|loop| A",
			want: "direction: right\n" +
				"A: Start\n" +
				"B: End\n" +
				"A -> B\n" +
				"B -> A: loop\n",
		},
		{
			name: "undirected and bidirectional",
			in:   "flowchart TD\n    a --- b\n    a <--> c",
			want: "a -- b\n" +
				"a <-> c\n",
		},
		{
			name: "dotted and thick arrows carry style",
			in:   "graph LR\n    a -.-> b\n    b ==> c",
			want: "direction: right\n" +
				"a -> b {style.stroke-dash: 3}\n" +
				"b -> c {style.stroke-width: 3}\n",
		},
		{
			name: "subgraph becomes container",
			in:   "flowchart TB\n    subgraph one[Group One]\n        a1 --> a2\n    end\n    a2 --> b",
			want: "one: Group One {\n" +
				"  a1 -> a2\n" +
				"}\n" +
				"one.a2 -> b\n",
		},
		{
			name: "labeled node inside subgraph",
			in:   "flowchart TD\n    subgraph grp[My Group]\n        x[Node X] --> y\n    end",
			want: "grp: My Group {\n" +
				"  x: Node X\n" +
				"  x -> y\n" +
				"}\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertConvertsTo(t, tt.in, tt.want)
		})
	}
}

func TestMermaidToD2NodeShapes(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "diamond",
			in:   "flowchart TD\n    a{Is it ok?}\n    a --> b",
			want: "a: Is it ok? {shape: diamond}\n" +
				"a -> b\n",
		},
		{
			name: "hexagon",
			in:   "flowchart TD\n    a{{Prepare}}\n    a --> b",
			want: "a: Prepare {shape: hexagon}\n" +
				"a -> b\n",
		},
		{
			name: "cylinder",
			in:   "flowchart TD\n    db[(Store)]\n    db --> b",
			want: "db: Store {shape: cylinder}\n" +
				"db -> b\n",
		},
		{
			name: "stadium becomes oval",
			in:   "flowchart TD\n    s([Begin])\n    s --> b",
			want: "s: Begin {shape: oval}\n" +
				"s -> b\n",
		},
		{
			name: "circle with label equal to id",
			in:   "flowchart TD\n    x((x))\n    x --> b",
			want: "x: {shape: circle}\n" +
				"x -> b\n",
		},
		{
			name: "rectangle and rounded fall back to default",
			in:   "flowchart TD\n    a[Box] --> b(Round)",
			want: "a: Box\n" +
				"b: Round\n" +
				"a -> b\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertConvertsTo(t, tt.in, tt.want)
		})
	}
}

func TestMermaidToD2EdgeStyling(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "dotted edge",
			in:   "flowchart TD\n    a -.-> b",
			want: "a -> b {style.stroke-dash: 3}\n",
		},
		{
			name: "thick edge",
			in:   "flowchart TD\n    a ==> b",
			want: "a -> b {style.stroke-width: 3}\n",
		},
		{
			name: "dotted edge with label",
			in:   "flowchart TD\n    a -.->|retry| b",
			want: "a -> b: retry {style.stroke-dash: 3}\n",
		},
		{
			name: "solid edge unchanged",
			in:   "flowchart TD\n    a --> b",
			want: "a -> b\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertConvertsTo(t, tt.in, tt.want)
		})
	}
}

func TestMermaidToD2Sequence(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "aliased participants and messages",
			in:   "sequenceDiagram\n    participant A as Alice\n    participant B as Bob\n    A->>B: Hello Bob\n    B-->>A: Hi Alice",
			want: "shape: sequence_diagram\n" +
				"A: Alice\n" +
				"B: Bob\n" +
				"A -> B: Hello Bob\n" +
				"B -> A: Hi Alice\n",
		},
		{
			name: "actor declaration and implicit self message",
			in:   "sequenceDiagram\n    actor U\n    U->>S: Request\n    S->>S: process",
			want: "shape: sequence_diagram\n" +
				"U\n" +
				"U -> S: Request\n" +
				"S -> S: process\n",
		},
		{
			name: "loop and alt become labeled groups",
			in:   "sequenceDiagram\n    participant A as Alice\n    participant B as Bob\n    loop every minute\n        A->>B: poll\n    end\n    alt success\n        B-->>A: 200\n    else failure\n        B-->>A: 500\n    end",
			want: "shape: sequence_diagram\n" +
				"A: Alice\n" +
				"B: Bob\n" +
				"loop_1: {\n" +
				"  label: every minute\n" +
				"  A -> B: poll\n" +
				"}\n" +
				"alt_1: {\n" +
				"  label: success\n" +
				"  B -> A: 200\n" +
				"}\n" +
				"alt_2: {\n" +
				"  label: failure\n" +
				"  B -> A: 500\n" +
				"}\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertConvertsTo(t, tt.in, tt.want)
		})
	}
}

// assertConvertsTo checks that MermaidToD2(in) equals want and that the output
// is valid D2 (it compiles).
func assertConvertsTo(t *testing.T, in, want string) {
	t.Helper()
	got, err := MermaidToD2(in)
	if err != nil {
		t.Fatalf("MermaidToD2(%q) error: %v", in, err)
	}
	if got != want {
		t.Errorf("MermaidToD2(%q)\n got:\n%s\nwant:\n%s", in, got, want)
	}
	if _, _, err := d2compiler.Compile("", strings.NewReader(got), nil); err != nil {
		t.Errorf("MermaidToD2(%q) produced invalid D2: %v\n%s", in, err, got)
	}
}

// TestMermaidToD2Testdata converts every Mermaid fixture and checks the output
// compiles as D2, exercising richer real-world inputs than the unit tables.
func TestMermaidToD2Testdata(t *testing.T) {
	files, err := filepath.Glob("testdata/*.mmd")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("no testdata/*.mmd fixtures found")
	}
	for _, f := range files {
		t.Run(filepath.Base(f), func(t *testing.T) {
			src, err := os.ReadFile(f)
			if err != nil {
				t.Fatal(err)
			}
			got, err := MermaidToD2(string(src))
			if err != nil {
				t.Fatalf("MermaidToD2(%s) error: %v", f, err)
			}
			if _, _, err := d2compiler.Compile("", strings.NewReader(got), nil); err != nil {
				t.Errorf("MermaidToD2(%s) produced invalid D2: %v\n%s", f, err, got)
			}
		})
	}
}

func TestMermaidToD2UnsupportedType(t *testing.T) {
	_, err := MermaidToD2("pie title Pets\n \"Dogs\" : 50")
	if err == nil {
		t.Fatal("MermaidToD2(pie) error = nil, want unsupported-type error")
	}
	if !strings.Contains(err.Error(), "pie") {
		t.Errorf("error %q does not name the unsupported type %q", err, "pie")
	}
}
