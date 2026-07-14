package mermaid2d2

import "testing"

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
			name: "class with no members",
			in:   "Dog: {shape: class}\n",
			want: "classDiagram\n    class Dog\n",
		},
		{
			name: "inheritance relationship",
			in:   "Dog -> Animal: {target-arrowhead: {shape: triangle; style.filled: false}}\n",
			want: "classDiagram\n    Dog --|> Animal\n",
		},
		{
			name: "association with label",
			in:   "A -> B: uses\n",
			want: "classDiagram\n    A --> B : uses\n",
		},
		{
			name: "composition with cardinalities",
			in:   "Car -> Wheel: {source-arrowhead: {shape: diamond; label: 1}; target-arrowhead: {label: 4}}\n",
			want: "classDiagram\n    Car \"1\" *-- \"4\" Wheel\n",
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
