package mermaid2d2

import (
	"slices"
	"strings"
	"testing"

	"oss.terrastruct.com/d2/d2compiler"
)

func TestD2ToMermaidFlowchartStyling(t *testing.T) {
	in := "classes: {\n" +
		"  hot: {\n" +
		"    style: {\n" +
		"      fill: \"#f96\"\n" +
		"      stroke: \"#333\"\n" +
		"      stroke-width: 2\n" +
		"    }\n" +
		"  }\n" +
		"}\n" +
		"A: Start {class: hot}\n" +
		"B: End\n" +
		"A -> B\n"
	want := "flowchart TD\n" +
		"    A[\"Start\"]\n" +
		"    B[\"End\"]\n" +
		"    A --> B\n" +
		"    classDef hot fill:#f96,stroke:#333,stroke-width:2px\n" +
		"    class A hot\n"

	got, err := D2ToMermaid(in)
	if err != nil {
		t.Fatalf("D2ToMermaid(%q) error: %v", in, err)
	}
	if got != want {
		t.Errorf("D2ToMermaid(%q)\n got:\n%s\nwant:\n%s", in, got, want)
	}
}

func TestD2ToMermaidFlowchartStylingDasharrayAndFontSize(t *testing.T) {
	in := "classes: {\n" +
		"  dashed: {\n" +
		"    style: {\n" +
		"      stroke-dash: 5\n" +
		"      font-size: 20\n" +
		"    }\n" +
		"  }\n" +
		"}\n" +
		"A: Start {class: dashed}\n" +
		"B: End\n" +
		"A -> B\n"
	want := "flowchart TD\n" +
		"    A[\"Start\"]\n" +
		"    B[\"End\"]\n" +
		"    A --> B\n" +
		"    classDef dashed stroke-dasharray:5,font-size:20px\n" +
		"    class A dashed\n"

	got, err := D2ToMermaid(in)
	if err != nil {
		t.Fatalf("D2ToMermaid(%q) error: %v", in, err)
	}
	if got != want {
		t.Errorf("D2ToMermaid(%q)\n got:\n%s\nwant:\n%s", in, got, want)
	}
}

func TestD2ToMermaidFlowchartInlineStyle(t *testing.T) {
	in := "A: Start {style: {fill: \"#f96\"; stroke: \"#333\"; stroke-width: 2}}\n" +
		"B: End\n" +
		"A -> B\n"
	want := "flowchart TD\n" +
		"    A[\"Start\"]\n" +
		"    B[\"End\"]\n" +
		"    A --> B\n" +
		"    style A fill:#f96,stroke:#333,stroke-width:2px\n"

	got, err := D2ToMermaid(in)
	if err != nil {
		t.Fatalf("D2ToMermaid(%q) error: %v", in, err)
	}
	if got != want {
		t.Errorf("D2ToMermaid(%q)\n got:\n%s\nwant:\n%s", in, got, want)
	}
}

// A class already covers its objects' styles, so the class line alone carries
// them; only a per-object override needs a style line of its own beside it.
func TestD2ToMermaidFlowchartInlineStyleOverridesClass(t *testing.T) {
	in := "classes: {\n" +
		"  hot: {style: {fill: \"#f96\"; stroke: \"#333\"}}\n" +
		"}\n" +
		"A: Plain {class: hot}\n" +
		"B: Louder {class: hot; style.fill: \"#0f0\"}\n" +
		"A -> B\n"
	want := "flowchart TD\n" +
		"    A[\"Plain\"]\n" +
		"    B[\"Louder\"]\n" +
		"    A --> B\n" +
		"    classDef hot fill:#f96,stroke:#333\n" +
		"    class A hot\n" +
		"    class B hot\n" +
		"    style B fill:#0f0,stroke:#333\n"

	got, err := D2ToMermaid(in)
	if err != nil {
		t.Fatalf("D2ToMermaid(%q) error: %v", in, err)
	}
	if got != want {
		t.Errorf("D2ToMermaid(%q)\n got:\n%s\nwant:\n%s", in, got, want)
	}
}

// TestD2InlineStyleRoundTrip pins the two directions against each other: the
// style D2ToMermaid emits is the style MermaidToD2 reads back. Compared on the
// compiled graph rather than the text, since D2 accepts the same style in
// nested and flat spellings.
func TestD2InlineStyleRoundTrip(t *testing.T) {
	in := "A {style: {fill: \"#f96\"; stroke: \"#333\"; stroke-width: 2}}\n" +
		"A -> B\n"

	mmd, err := D2ToMermaid(in)
	if err != nil {
		t.Fatalf("D2ToMermaid(%q) error: %v", in, err)
	}
	back, err := MermaidToD2(mmd)
	if err != nil {
		t.Fatalf("MermaidToD2(%q) error: %v", mmd, err)
	}
	if got, want := styleOf(t, back, "A"), styleOf(t, in, "A"); !slices.Equal(got, want) {
		t.Errorf("style of A after round trip = %v, want %v\nvia:\n%s", got, want, mmd)
	}
}

// styleOf compiles src and returns the named object's Mermaid-mappable style.
func styleOf(t *testing.T, src, id string) []string {
	t.Helper()
	graph, _, err := d2compiler.Compile("", strings.NewReader(src), nil)
	if err != nil {
		t.Fatalf("compile %q: %v", src, err)
	}
	for _, obj := range graph.Objects {
		if obj.AbsID() == id {
			return d2StyleToMermaid(obj.Style)
		}
	}
	t.Fatalf("object %q not found in %q", id, src)
	return nil
}
