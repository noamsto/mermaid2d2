package mermaid2d2

import (
	"strings"
	"testing"

	mermaid "github.com/noamsto/mermaid-check"
	"github.com/noamsto/mermaid-check/ast"
	"oss.terrastruct.com/d2/d2compiler"
)

func TestClassDiagramToD2(t *testing.T) {
	tests := []struct {
		name string
		cd   *ast.ClassDiagram
		want string
	}{
		{
			name: "fields and methods with mixed visibility",
			cd: parseClassDiagram(t, "classDiagram\n"+
				"    class Animal {\n"+
				"        +String name\n"+
				"        -int age\n"+
				"        #bool alive\n"+
				"        ~int region\n"+
				"        +makeSound(String kind) int\n"+
				"    }\n"),
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
			name: "class with no members",
			cd:   parseClassDiagram(t, "classDiagram\n    class Dog"),
			want: "Dog: {shape: class}\n",
		},
		{
			// The parser drops inheritance edges, so build the AST directly to
			// exercise the arrowhead mapping.
			name: "inheritance relationship",
			cd: &ast.ClassDiagram{Statements: []ast.ClassStmt{
				&ast.Relationship{From: "Dog", To: "Animal", Type: "inheritance"},
			}},
			want: "Dog -> Animal: {target-arrowhead: {shape: triangle; style.filled: false}}\n",
		},
		{
			name: "association with label",
			cd:   parseClassDiagram(t, "classDiagram\n    A --> B : uses"),
			want: "A -> B: uses\n",
		},
		{
			name: "composition with cardinalities",
			cd:   parseClassDiagram(t, "classDiagram\n    Car \"1\" *-- \"4\" Wheel"),
			want: "Car -> Wheel: {source-arrowhead: {shape: diamond; label: 1}; target-arrowhead: {label: 4}}\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertClassConvertsTo(t, tt.cd, tt.want)
		})
	}
}

// assertClassConvertsTo checks that classDiagramToD2(cd) equals want and that the
// output is valid D2 (it compiles).
func assertClassConvertsTo(t *testing.T, cd *ast.ClassDiagram, want string) {
	t.Helper()
	got := classDiagramToD2(cd)
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
