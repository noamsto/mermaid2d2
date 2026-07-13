package mermaid2d2

import (
	"testing"
)

func TestMermaidToD2Mindmap(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "root with children and a shape",
			in:   "mindmap\n  root((Origin))\n    Ideas\n    Tools",
			want: "n1: Origin {shape: circle}\n" +
				"n2: Ideas\n" +
				"n1 -> n2\n" +
				"n3: Tools\n" +
				"n1 -> n3\n",
		},
		{
			name: "nested grandchild",
			in:   "mindmap\n  Root\n    Branch\n      Leaf",
			want: "n1: Root\n" +
				"n2: Branch\n" +
				"n1 -> n2\n" +
				"n3: Leaf\n" +
				"n2 -> n3\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertConvertsTo(t, tt.in, tt.want)
		})
	}
}
