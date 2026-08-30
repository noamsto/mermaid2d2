package mermaid2d2

import (
	"strings"
	"testing"

	mermaid "github.com/sammcj/mermaid-check"
	"github.com/sammcj/mermaid-check/ast"
	"oss.terrastruct.com/d2/d2compiler"
)

func TestClassDiagramToD2(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want string
	}{
		{
			name: "fields and methods with mixed visibility",
			src: "classDiagram\n" +
				"    class Animal {\n" +
				"        +String name\n" +
				"        -int age\n" +
				"        #bool alive\n" +
				"        ~int region\n" +
				"        +makeSound(String kind) int\n" +
				"    }\n",
			want: "Animal: {\n" +
				"  shape: class\n" +
				"  +name: String\n" +
				"  -age: int\n" +
				"  \\#alive: bool\n" +
				"  ~region: int\n" +
				"  +makeSound(String kind): int\n" +
				"}\n",
		},
		{
			name: "static and abstract member classifiers",
			src: "classDiagram\n" +
				"    class Widget {\n" +
				"        +String colour$\n" +
				"        +draw()*\n" +
				"        +make() Widget$\n" +
				"        #int count$\n" +
				"    }\n",
			want: "Widget: {\n" +
				"  shape: class\n" +
				"  \"+colour$\": String\n" +
				"  \"+draw()*\"\n" +
				"  \"+make()$\": Widget\n" +
				"  \"#count$\": int\n" +
				"}\n",
		},
		{
			name: "class with no members",
			src:  "classDiagram\n    class Dog",
			want: "Dog: {shape: class}\n",
		},
		{
			name: "inheritance relationship",
			src:  "classDiagram\n    Dog --|> Animal",
			want: "Dog -> Animal: {target-arrowhead: {shape: triangle; style.filled: false}}\n",
		},
		{
			name: "left-pointing inheritance keeps the superclass first",
			src:  "classDiagram\n    Animal <|-- Dog",
			want: "Animal <-> Dog: {source-arrowhead: {shape: triangle; style.filled: false}; target-arrowhead: {shape: none}}\n",
		},
		{
			name: "aggregation diamond is hollow, composition filled",
			src:  "classDiagram\n    Team o-- Player\n    House *-- Room",
			want: "Team <-> Player: {source-arrowhead: {shape: diamond; style.filled: false}; target-arrowhead: {shape: none}}\n" +
				"House <-> Room: {source-arrowhead: {shape: diamond; style.filled: true}; target-arrowhead: {shape: none}}\n",
		},
		{
			name: "left-pointing association keeps written order",
			src:  "classDiagram\n    A <-- B",
			want: "A <-> B: {source-arrowhead: {shape: arrow}; target-arrowhead: {shape: none}}\n",
		},
		{
			name: "undirected association has no arrowheads",
			src:  "classDiagram\n    A -- B",
			want: "A -- B\n",
		},
		{
			name: "two-way association keeps both arrowheads",
			src:  "classDiagram\n    A <--> B",
			want: "A <-> B\n",
		},
		{
			name: "association with label",
			src:  "classDiagram\n    A --> B : uses",
			want: "A -> B: uses\n",
		},
		{
			name: "composition with cardinalities",
			src:  "classDiagram\n    Car \"1\" *-- \"4\" Wheel",
			want: "Car <-> Wheel: {source-arrowhead: {shape: diamond; style.filled: true; label: 1}; target-arrowhead: {shape: none; label: 4}}\n",
		},
		{
			name: "note becomes a tooltip on its target class",
			src:  "classDiagram\n    class Dog\n    note for Dog \"some note\"",
			want: "Dog: {shape: class}\n" +
				"Dog.tooltip: some note\n",
		},
		{
			name: "note text needing d2 quoting",
			src:  "classDiagram\n    class Dog\n    note for Dog \"50% done #tag\"",
			want: "Dog: {shape: class}\n" +
				"Dog.tooltip: \"50% done #tag\"\n",
		},
		{
			name: "standalone note becomes a floating node",
			src:  "classDiagram\n    note \"a floating note\"",
			want: "note_1: a floating note\n",
		},
		{
			name: "multiple standalone notes get unique ids",
			src: "classDiagram\n" +
				"    note \"first\"\n" +
				"    note \"second\"\n",
			want: "note_1: first\n" +
				"note_2: second\n",
		},
		{
			name: "targeted note still renders as a tooltip alongside a standalone note",
			src: "classDiagram\n" +
				"    class Dog\n" +
				"    note for Dog \"dog note\"\n" +
				"    note \"floating note\"\n",
			want: "Dog: {shape: class}\n" +
				"Dog.tooltip: dog note\n" +
				"note_1: floating note\n",
		},
		{
			name: "standalone note skips an id already taken by a class name",
			src: "classDiagram\n" +
				"    class note_1\n" +
				"    note \"floating note\"\n",
			want: "note_1: {shape: class}\n" +
				"note_2: floating note\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertClassConvertsTo(t, tt.src, tt.want)
		})
	}
}

// assertClassConvertsTo checks that converting src equals want and that the
// output is valid D2 (it compiles).
func assertClassConvertsTo(t *testing.T, src, want string) {
	t.Helper()
	got := classDiagramToD2(parseClassDiagram(t, src), src)
	if got != want {
		t.Errorf("classDiagramToD2\n got:\n%s\nwant:\n%s", got, want)
	}
	if _, _, err := d2compiler.Compile("", strings.NewReader(got), nil); err != nil {
		t.Errorf("classDiagramToD2 produced invalid D2: %v\n%s", err, got)
	}
}

func parseClassDiagram(t *testing.T, src string) *ast.ClassDiagram {
	t.Helper()
	d, err := mermaid.Parse(src)
	if err != nil {
		t.Fatalf("mermaid.Parse(%q) error: %v", src, err)
	}
	cd, ok := d.(*ast.ClassDiagram)
	if !ok {
		t.Fatalf("mermaid.Parse(%q) = %T, want *ast.ClassDiagram", src, d)
	}
	return cd
}
