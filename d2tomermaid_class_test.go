package mermaid2d2

import (
	"testing"

	mermaid "github.com/sammcj/mermaid-check"
)

func TestClassDiagramFromD2Classes(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "fields and methods with mixed visibility",
			in: "Animal: {\n" +
				"  shape: class\n" +
				"  +name: String\n" +
				"  -age: int\n" +
				"  \\#alive: bool\n" +
				"  ~region: int\n" +
				"  +makeSound(String kind): int\n" +
				"}\n",
			want: "classDiagram\n" +
				"    class Animal {\n" +
				"        +String name\n" +
				"        -int age\n" +
				"        #bool alive\n" +
				"        ~int region\n" +
				"        +makeSound(String kind) int\n" +
				"    }\n",
		},
		{
			name: "static and abstract member classifiers",
			in: "Widget: {\n" +
				"  shape: class\n" +
				"  \"+colour$\": String\n" +
				"  \"+draw()*\"\n" +
				"  \"+make()$\": Widget\n" +
				"  \"#count$\": int\n" +
				"}\n",
			want: "classDiagram\n" +
				"    class Widget {\n" +
				"        +String colour$\n" +
				"        #int count$\n" +
				"        +draw()*\n" +
				"        +make() Widget$\n" +
				"    }\n",
		},
		{
			name: "class with no members",
			in:   "Dog: {shape: class}\n",
			want: "classDiagram\n    class Dog\n",
		},
		{
			name: "inheritance relationship",
			in: "Dog: {shape: class}\n" +
				"Animal: {shape: class}\n" +
				"Dog -> Animal: {target-arrowhead: {shape: triangle; style.filled: false}}\n",
			want: "classDiagram\n    class Dog\n    class Animal\n    Dog --|> Animal\n",
		},
		{
			name: "association with label",
			in: "A: {shape: class}\n" +
				"B: {shape: class}\n" +
				"A -> B: uses\n",
			want: "classDiagram\n    class A\n    class B\n    A --> B : uses\n",
		},
		{
			name: "composition with cardinalities",
			in: "Car: {shape: class}\n" +
				"Wheel: {shape: class}\n" +
				"Car -> Wheel: {source-arrowhead: {shape: diamond; label: 1}; target-arrowhead: {label: 4}}\n",
			want: "classDiagram\n    class Car\n    class Wheel\n    Car \"1\" *-- \"4\" Wheel\n",
		},
		{
			name: "realization",
			in: "Dog: {shape: class}\n" +
				"Animal: {shape: class}\n" +
				"Dog -> Animal: {style.stroke-dash: 5; target-arrowhead: {shape: triangle; style.filled: false}}\n",
			want: "classDiagram\n    class Dog\n    class Animal\n    Dog ..|> Animal\n",
		},
		{
			name: "dependency",
			in: "A: {shape: class}\n" +
				"B: {shape: class}\n" +
				"A -> B: {style.stroke-dash: 5}\n",
			want: "classDiagram\n    class A\n    class B\n    A ..> B\n",
		},
		{
			name: "aggregation",
			in: "Car: {shape: class}\n" +
				"Wheel: {shape: class}\n" +
				"Car -> Wheel: {source-arrowhead: {shape: diamond; style.filled: false}}\n",
			want: "classDiagram\n    class Car\n    class Wheel\n    Car o-- Wheel\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			graph := compileD2(t, tt.in)
			got := classDiagramFromD2(graph)
			if got != tt.want {
				t.Errorf("classDiagramFromD2(%q)\n got:\n%s\nwant:\n%s", tt.in, got, tt.want)
			}
		})
	}
}

func TestClassDiagramFromD2SanitizesIdentifiers(t *testing.T) {
	graph := compileD2(t, "\"Order-Item\": {shape: class}\ncustomer: {shape: class}\ncustomer -> \"Order-Item\"\n")
	got := classDiagramFromD2(graph)
	if _, err := mermaid.Parse(got); err != nil {
		t.Errorf("classDiagramFromD2 produced unparsable Mermaid: %v\n%s", err, got)
	}
}
