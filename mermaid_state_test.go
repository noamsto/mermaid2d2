package mermaid2d2

import (
	"strings"
	"testing"

	mermaid "github.com/sammcj/mermaid-check"
	"github.com/sammcj/mermaid-check/ast"

	"oss.terrastruct.com/d2/d2compiler"
)

func TestStateDiagramToD2(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "declared state and labeled transition",
			in:   "stateDiagram-v2\n    state \"Named\" as s1\n    s1 --> s2: go",
			want: "s1: Named\n" +
				"s1 -> s2: go\n",
		},
		{
			name: "start and end sentinels",
			in:   "stateDiagram-v2\n    [*] --> A\n    A --> [*]",
			want: "start: {shape: circle}\n" +
				"start -> A\n" +
				"end: {shape: circle}\n" +
				"A -> end\n",
		},
		{
			name: "multiple ends share one sentinel",
			in:   "stateDiagram-v2\n    A --> B\n    B --> [*]\n    A --> [*]",
			want: "A -> B\n" +
				"end: {shape: circle}\n" +
				"B -> end\n" +
				"A -> end\n",
		},
		{
			name: "choice node becomes diamond",
			in:   "stateDiagram-v2\n    state if_state <<choice>>\n    [*] --> if_state\n    if_state --> A: yes\n    if_state --> B: no",
			want: "if_state: {shape: diamond}\n" +
				"start: {shape: circle}\n" +
				"start -> if_state\n" +
				"if_state -> A: yes\n" +
				"if_state -> B: no\n",
		},
		{
			name: "fork fans out",
			in:   "stateDiagram-v2\n    state fork_state <<fork>>\n    [*] --> fork_state\n    fork_state --> s1\n    fork_state --> s2",
			want: "fork_state\n" +
				"start: {shape: circle}\n" +
				"start -> fork_state\n" +
				"fork_state -> s1\n" +
				"fork_state -> s2\n",
		},
		{
			name: "note becomes a tooltip on its target state",
			in:   "stateDiagram-v2\n    state \"Named\" as s1\n    s1 --> s2: go\n    note right of s1: hello",
			want: "s1: Named\n" +
				"s1 -> s2: go\n" +
				"s1.tooltip: hello\n",
		},
		{
			name: "note text needing d2 quoting",
			in:   "stateDiagram-v2\n    state \"Named\" as s1\n    note right of s1: 50% done #tag",
			want: "s1: Named\n" +
				"s1.tooltip: \"50% done #tag\"\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertStateConvertsTo(t, parseState(t, tt.in), tt.want)
		})
	}
}

// TestStateDiagramToD2Composite exercises the composite-container path via a
// real parse (mermaid-check v0.1.3+ populates State.IsComposite/Nested).
func TestStateDiagramToD2Composite(t *testing.T) {
	sd := parseState(t, "stateDiagram-v2\n"+
		"    [*] --> First\n"+
		"    state \"Big Room\" as First {\n"+
		"        [*] --> s2\n"+
		"        state \"Second\" as s2\n"+
		"        note left of s2: inner note\n"+
		"        s2 --> s3\n"+
		"        s3 --> [*]\n"+
		"    }\n"+
		"    First --> [*]")
	want := "start: {shape: circle}\n" +
		"start -> First\n" +
		"First: Big Room {\n" +
		"  start: {shape: circle}\n" +
		"  start -> s2\n" +
		"  s2: Second\n" +
		"  s2.tooltip: inner note\n" +
		"  s2 -> s3\n" +
		"  end: {shape: circle}\n" +
		"  s3 -> end\n" +
		"}\n" +
		"end: {shape: circle}\n" +
		"First -> end\n"
	assertStateConvertsTo(t, sd, want)
}

func parseState(t *testing.T, src string) *ast.StateDiagram {
	t.Helper()
	d, err := mermaid.Parse(src)
	if err != nil {
		t.Fatalf("mermaid.Parse(%q) error: %v", src, err)
	}
	sd, ok := d.(*ast.StateDiagram)
	if !ok {
		t.Fatalf("mermaid.Parse(%q) = %T, want *ast.StateDiagram", src, d)
	}
	return sd
}

// assertStateConvertsTo checks that stateDiagramToD2 equals want and that the
// output is valid D2 (it compiles).
func assertStateConvertsTo(t *testing.T, sd *ast.StateDiagram, want string) {
	t.Helper()
	got := stateDiagramToD2(sd)
	if got != want {
		t.Errorf("stateDiagramToD2\n got:\n%s\nwant:\n%s", got, want)
	}
	if _, _, err := d2compiler.Compile("", strings.NewReader(got), nil); err != nil {
		t.Errorf("stateDiagramToD2 produced invalid D2: %v\n%s", err, got)
	}
}
