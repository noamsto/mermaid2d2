package mermaid2d2

import (
	"fmt"
	"strings"

	"oss.terrastruct.com/d2/d2graph"
)

// writeStyling emits classDef/class lines for every object carrying a D2
// class: name whose resolved Style has at least one property with a Mermaid
// equivalent. The first object to carry a given class name supplies that
// class's style (D2 bakes each class's properties directly onto every object
// referencing it, so — barring a per-object style override — they agree).
func (e *mermaidEmitter) writeStyling(b *strings.Builder, graph *d2graph.Graph) {
	classCSS := map[string][]string{}
	var classOrder []string
	for _, obj := range graph.Objects {
		for _, name := range obj.Classes {
			if _, seen := classCSS[name]; seen {
				continue
			}
			if css := d2StyleToMermaid(obj.Style); len(css) > 0 {
				classCSS[name] = css
				classOrder = append(classOrder, name)
			}
		}
	}
	for _, name := range classOrder {
		fmt.Fprintf(b, "    classDef %s %s\n", name, strings.Join(classCSS[name], ","))
	}
	for _, obj := range graph.Objects {
		if obj.IsContainer() {
			continue
		}
		var names []string
		for _, name := range obj.Classes {
			if _, ok := classCSS[name]; ok {
				names = append(names, name)
			}
		}
		if len(names) > 0 {
			fmt.Fprintf(b, "    class %s %s\n", e.id(obj), strings.Join(names, ","))
		}
	}
}

// d2StyleToMermaid inverts mermaidStyleToD2 (mermaidtod2.go): the same six
// properties, in the same order, with the "px" suffix restored on the
// numeric ones (stroke-dasharray excepted — it's conventionally unitless).
// Colors need no quoting on the way out (Mermaid classDef values are bare,
// unlike D2's leading-# comment ambiguity).
func d2StyleToMermaid(s d2graph.Style) []string {
	props := []struct {
		field   *d2graph.Scalar
		css     string
		numeric bool
	}{
		{s.Fill, "fill", false},
		{s.Stroke, "stroke", false},
		{s.StrokeWidth, "stroke-width", true},
		{s.StrokeDash, "stroke-dasharray", false},
		{s.FontColor, "color", false},
		{s.FontSize, "font-size", true},
	}
	var lines []string
	for _, p := range props {
		if p.field == nil || strings.TrimSpace(p.field.Value) == "" {
			continue
		}
		val := p.field.Value
		if p.numeric {
			val += "px"
		}
		lines = append(lines, p.css+":"+val)
	}
	return lines
}
