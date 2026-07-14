package mermaid2d2

import "testing"

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
