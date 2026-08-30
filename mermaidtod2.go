package mermaid2d2

import (
	"fmt"
	"strconv"
	"strings"

	mermaid "github.com/sammcj/mermaid-check"
	"github.com/sammcj/mermaid-check/ast"
)

// MermaidToD2 parses Mermaid source and emits equivalent D2.
//
// Flowchart nodes, edges, and subgraphs map onto D2 shapes, connections, and
// containers; sequence diagrams map onto a D2 sequence_diagram, with loop/alt/
// opt blocks becoming labeled groups; state diagrams map onto nodes and
// connections, with composite states becoming containers; class diagrams map
// onto D2 class shapes, with relationships becoming connections whose arrowheads
// encode the UML relation type; ER diagrams map onto D2 sql_table shapes, with
// attributes as typed columns and relationships as connections; mindmaps map
// onto a D2 tree of nodes connected to their parents; C4 diagrams (Context,
// Container, Component, Dynamic, Deployment) map onto D2 nodes and containers,
// with Rel(...) relationships becoming connections. Diagram types with no D2
// equivalent (gantt, pie, journey, ...) return an error rather than
// being mangled. Sequence notes become D2 notes; state and class notes render
// as a tooltip attribute on their target node.
func MermaidToD2(src string) (string, error) {
	diagram, err := mermaid.Parse(src)
	if err != nil {
		return "", fmt.Errorf("mermaid2d2: parse mermaid: %w", err)
	}
	switch d := diagram.(type) {
	case *ast.Flowchart:
		return flowchartToD2(d), nil
	case *ast.SequenceDiagram:
		return sequenceToD2(d), nil
	case *ast.StateDiagram:
		return stateDiagramToD2(d), nil
	case *ast.ClassDiagram:
		return classDiagramToD2(d), nil
	case *ast.ERDiagram:
		return erDiagramToD2(d), nil
	case *ast.MindmapDiagram:
		return mindmapToD2(d), nil
	case *ast.C4Diagram:
		return c4DiagramToD2(d), nil
	default:
		return "", fmt.Errorf("mermaid2d2: unsupported diagram type %q: only flowchart, sequence, state, class, er, mindmap, and C4 are supported", diagram.GetType())
	}
}

// d2Container is a node in the D2 output tree: the root board or a subgraph.
type d2Container struct {
	path     string // dotted D2 path; "" for the root board
	label    string // container label, or "" to omit
	children []*d2Container
	edges    []d2Edge // edges whose innermost common scope is this container
}

type d2Edge struct {
	from, to string // Mermaid (global) node ids
	arrow    string // resolved D2 arrow
	label    string
	style    string // D2 connection style attribute, or "" for a plain edge
}

func flowchartToD2(fc *ast.Flowchart) string {
	e := &flowEmitter{
		containerOf: map[string]string{},
		labelOf:     map[string]string{},
		shapeOf:     map[string]string{},
		classOf:     map[string][]string{},
		byPath:      map[string]*d2Container{},
		usedSlugs:   map[string]bool{},
		classStyles: map[string][]string{},
	}
	root := &d2Container{}
	e.byPath[""] = root
	e.walk(fc.Statements, root)

	// Membership of both endpoints is only known after the full walk, so edges
	// are placed into their scope now: the innermost container enclosing both.
	for _, ed := range e.edges {
		scope := commonContainer(e.containerOf[ed.from], e.containerOf[ed.to])
		c := e.byPath[scope]
		c.edges = append(c.edges, ed)
	}

	var b strings.Builder
	if dir := d2Direction(fc.Direction); dir != "" {
		fmt.Fprintf(&b, "direction: %s\n", dir)
	}
	e.emitClasses(&b)
	e.emit(&b, root, 0)
	return b.String()
}

type flowEmitter struct {
	containerOf map[string]string   // node id -> container path (first mention wins)
	labelOf     map[string]string   // node id -> label
	shapeOf     map[string]string   // node id -> D2 shape keyword (non-default shapes only)
	classOf     map[string][]string // node id -> applied class names
	nodeOrder   []string            // node ids in first-seen order
	byPath      map[string]*d2Container
	usedSlugs   map[string]bool
	edges       []d2Edge
	classStyles map[string][]string // classDef name -> D2 style lines (mappable only)
	classOrder  []string            // classDef names in first-seen order
}

// note records a node's container on first mention; Mermaid binds a node to the
// subgraph where it first appears, and later references do not move it.
func (e *flowEmitter) note(id, path string) {
	if _, ok := e.containerOf[id]; ok {
		return
	}
	e.containerOf[id] = path
	e.nodeOrder = append(e.nodeOrder, id)
}

func (e *flowEmitter) walk(stmts []ast.Statement, c *d2Container) {
	for _, s := range stmts {
		switch v := s.(type) {
		case *ast.NodeDef:
			e.note(v.ID, c.path)
			if v.Label != "" && v.Label != v.ID {
				e.labelOf[v.ID] = v.Label
			}
			if shape := d2Shape(v.Shape); shape != "" {
				e.shapeOf[v.ID] = shape
			}
		case *ast.Link:
			e.note(v.From, c.path)
			e.note(v.To, c.path)
			e.edges = append(e.edges, d2Edge{
				from:  v.From,
				to:    v.To,
				arrow: d2Arrow(v.Arrow, v.BiDir),
				label: strings.TrimSpace(v.Label),
				style: d2EdgeStyle(v.Arrow),
			})
		case *ast.ClassDef:
			if lines := mermaidStyleToD2(v.Styles); len(lines) > 0 {
				if _, seen := e.classStyles[v.Name]; !seen {
					e.classOrder = append(e.classOrder, v.Name)
				}
				e.classStyles[v.Name] = lines
			}
		case *ast.ClassAssignment:
			for _, id := range v.NodeIDs {
				e.note(id, c.path)
				e.classOf[id] = append(e.classOf[id], v.ClassName)
			}
		case *ast.Subgraph:
			// The subgraph id is the D2 container id. The quoted-title form has
			// no id, so fall back to a slug of the title.
			id := v.ID
			if id == "" {
				id = e.slug(v.Title)
			} else {
				e.usedSlugs[id] = true
			}
			path := id
			if c.path != "" {
				path = c.path + "." + id
			}
			label := ""
			if v.Title != "" && v.Title != id {
				label = v.Title
			}
			child := &d2Container{path: path, label: label}
			e.byPath[path] = child
			c.children = append(c.children, child)
			e.walk(v.Statements, child)
		}
	}
}

func (e *flowEmitter) emit(b *strings.Builder, c *d2Container, depth int) {
	indent := strings.Repeat("  ", depth)

	for _, id := range e.nodeOrder {
		if e.containerOf[id] != c.path {
			continue
		}
		lbl, hasLbl := e.labelOf[id]
		var attrs []string
		if shape, ok := e.shapeOf[id]; ok {
			attrs = append(attrs, "shape: "+shape)
		}
		if class := e.nodeClass(id); class != "" {
			attrs = append(attrs, "class: "+class)
		}
		switch {
		case hasLbl && len(attrs) > 0:
			fmt.Fprintf(b, "%s%s: %s {%s}\n", indent, id, d2Label(lbl), strings.Join(attrs, "; "))
		case len(attrs) > 0:
			fmt.Fprintf(b, "%s%s: {%s}\n", indent, id, strings.Join(attrs, "; "))
		case hasLbl:
			fmt.Fprintf(b, "%s%s: %s\n", indent, id, d2Label(lbl))
		}
	}

	for _, child := range c.children {
		tail := strings.TrimPrefix(child.path, c.path+".")
		if c.path == "" {
			tail = child.path
		}
		if child.label != "" {
			fmt.Fprintf(b, "%s%s: %s {\n", indent, tail, d2Label(child.label))
		} else {
			fmt.Fprintf(b, "%s%s: {\n", indent, tail)
		}
		e.emit(b, child, depth+1)
		fmt.Fprintf(b, "%s}\n", indent)
	}

	for _, ed := range c.edges {
		from, to := e.rel(ed.from, c.path), e.rel(ed.to, c.path)
		line := fmt.Sprintf("%s%s %s %s", indent, from, ed.arrow, to)
		if ed.label != "" {
			line += ": " + d2Label(ed.label)
		}
		if ed.style != "" {
			line += " {" + ed.style + "}"
		}
		fmt.Fprintln(b, line)
	}
}

// emitClasses writes the top-level D2 `classes` block for every classDef that
// mapped to at least one style. Empty style maps are omitted (D2 rejects them).
func (e *flowEmitter) emitClasses(b *strings.Builder) {
	if len(e.classOrder) == 0 {
		return
	}
	b.WriteString("classes: {\n")
	for _, name := range e.classOrder {
		fmt.Fprintf(b, "  %s: {\n    style: {\n", name)
		for _, line := range e.classStyles[name] {
			fmt.Fprintf(b, "      %s\n", line)
		}
		b.WriteString("    }\n  }\n")
	}
	b.WriteString("}\n")
}

// nodeClass renders a node's D2 class attribute value, keeping only classes that
// carry a style (a bare `class: x` to an undefined class is pointless). Returns
// "" when the node has none; a single name, or the `[a; b]` array form.
func (e *flowEmitter) nodeClass(id string) string {
	var names []string
	for _, n := range e.classOf[id] {
		if _, ok := e.classStyles[n]; ok {
			names = append(names, n)
		}
	}
	switch len(names) {
	case 0:
		return ""
	case 1:
		return names[0]
	default:
		return "[" + strings.Join(names, "; ") + "]"
	}
}

// mermaidStyleToD2 maps Mermaid classDef CSS properties onto D2 style lines in a
// fixed order, dropping properties with no D2 equivalent. Colors are quoted (a
// bare leading # is a D2 comment); numeric properties have any px unit stripped
// and are dropped when the remaining value is not a number.
func mermaidStyleToD2(css map[string]string) []string {
	props := []struct {
		css, d2 string
		numeric bool
	}{
		{"fill", "fill", false},
		{"stroke", "stroke", false},
		{"stroke-width", "stroke-width", true},
		{"stroke-dasharray", "stroke-dash", true},
		{"color", "font-color", false},
		{"font-size", "font-size", true},
	}
	var lines []string
	for _, p := range props {
		val := strings.TrimSpace(css[p.css])
		if val == "" {
			continue
		}
		if !p.numeric {
			lines = append(lines, p.d2+": "+d2Label(val))
			continue
		}
		fields := strings.Fields(strings.ReplaceAll(strings.TrimSuffix(val, "px"), ",", " "))
		if len(fields) == 0 {
			continue
		}
		if _, err := strconv.ParseFloat(fields[0], 64); err != nil {
			continue
		}
		lines = append(lines, p.d2+": "+fields[0])
	}
	return lines
}

// rel renders a node id relative to a container scope, qualifying it with the
// path segments that lie below scope (e.g. scope "one" -> "a", scope "" -> "one.a").
func (e *flowEmitter) rel(id, scope string) string {
	cp := e.containerOf[id]
	r := strings.TrimPrefix(cp, scope)
	r = strings.TrimPrefix(r, ".")
	if r == "" {
		return id
	}
	return r + "." + id
}

// slug reduces a subgraph title to a unique D2 identifier.
func (e *flowEmitter) slug(title string) string {
	s := sanitizeID(title)
	orig := s
	for i := 2; e.usedSlugs[s]; i++ {
		s = fmt.Sprintf("%s_%d", orig, i)
	}
	e.usedSlugs[s] = true
	return s
}

// commonContainer returns the innermost container path enclosing both a and b.
func commonContainer(a, b string) string {
	if a == b {
		return a
	}
	as, bs := strings.Split(a, "."), strings.Split(b, ".")
	var common []string
	for i := 0; i < len(as) && i < len(bs); i++ {
		if as[i] != bs[i] {
			break
		}
		common = append(common, as[i])
	}
	return strings.Join(common, ".")
}

// d2Arrow maps a Mermaid link arrow onto a D2 connection direction; line style
// (dotted, thick) is handled separately by d2EdgeStyle.
func d2Arrow(arrow string, bidir bool) string {
	switch {
	case bidir:
		return "<->"
	case strings.ContainsAny(arrow, ">xo"):
		return "->"
	default:
		return "--"
	}
}

// d2EdgeStyle maps a Mermaid link's line style onto a D2 connection style
// attribute: dotted (-.->) becomes a dashed stroke, thick (==>) a wide stroke.
// A solid link returns "" (the D2 default).
func d2EdgeStyle(arrow string) string {
	switch {
	case strings.Contains(arrow, "."):
		return "style.stroke-dash: 3"
	case strings.Contains(arrow, "="):
		return "style.stroke-width: 3"
	default:
		return ""
	}
}

// d2Shape maps a Mermaid node's bracket pair onto a D2 shape keyword, returning
// "" for shapes that have no distinct D2 equivalent and render as the default
// rectangle (plain `[]`, rounded `()`, subroutine `[[]]`, asymmetric `>]`).
func d2Shape(bracket string) string {
	switch bracket {
	case "{}":
		return "diamond"
	case "{{}}":
		return "hexagon"
	case "(())":
		return "circle"
	case "[()]":
		return "cylinder"
	case "([])":
		return "oval"
	default:
		return ""
	}
}

// d2Label quotes a label that would otherwise misparse as D2 — one containing a
// D2 syntax character (;{}#|"<>[]) or a newline, or with leading/trailing space.
// Safe labels are returned unchanged so the common case stays unquoted. Inside a
// D2 double-quoted string, backslash and double-quote are the escape sequences.
func d2Label(s string) string {
	if s == "" || (s == strings.TrimSpace(s) && !strings.ContainsAny(s, ";{}#|\"<>[]\n")) {
		return s
	}
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return `"` + s + `"`
}

// d2Direction maps a Mermaid flowchart orientation onto a D2 direction,
// returning "" for the D2 default (down).
func d2Direction(dir string) string {
	switch strings.ToUpper(dir) {
	case "LR":
		return "right"
	case "RL":
		return "left"
	case "BT":
		return "up"
	default: // TB, TD, ""
		return ""
	}
}

func sequenceToD2(sd *ast.SequenceDiagram) string {
	var b strings.Builder
	b.WriteString("shape: sequence_diagram\n")

	// Declare participants up front to fix actor order and carry aliases;
	// participants used only in messages are created implicitly by D2.
	declared := map[string]bool{}
	for _, s := range sd.Statements {
		p, ok := s.(*ast.Participant)
		if !ok || declared[p.ID] {
			continue
		}
		declared[p.ID] = true
		if p.Alias != "" && p.Alias != p.ID {
			fmt.Fprintf(&b, "%s: %s\n", p.ID, d2Label(p.Alias))
			continue
		}
		fmt.Fprintf(&b, "%s\n", p.ID)
	}

	e := &seqEmitter{counts: map[string]int{}}
	e.emit(&b, sd.Statements, 0)
	return b.String()
}

// seqEmitter renders sequence messages and blocks, numbering block groups per
// kind so their D2 keys stay unique (colliding keys would merge into one group).
type seqEmitter struct{ counts map[string]int }

func (e *seqEmitter) emit(b *strings.Builder, stmts []ast.SeqStmt, depth int) {
	indent := strings.Repeat("  ", depth)
	for _, s := range stmts {
		switch v := s.(type) {
		case *ast.Message:
			if txt := strings.TrimSpace(v.Text); txt != "" {
				fmt.Fprintf(b, "%s%s -> %s: %s\n", indent, v.From, v.To, d2Label(txt))
				continue
			}
			fmt.Fprintf(b, "%s%s -> %s\n", indent, v.From, v.To)
		case *ast.Note:
			// A childless, edgeless leaf inside a sequence_diagram renders as a
			// note; scoping it under the first participant places it over that
			// actor's lifeline (D2 has no left/right/span note positioning, so
			// Position and any extra participants are approximated).
			e.counts["note"]++
			key := fmt.Sprintf("note_%d", e.counts["note"])
			if len(v.Participants) > 0 {
				key = v.Participants[0] + "." + key
			}
			fmt.Fprintf(b, "%s%s: %s\n", indent, key, d2Label(strings.TrimSpace(v.Text)))
		case *ast.Loop:
			e.group(b, depth, "loop", v.Label, v.Statements)
		case *ast.Opt:
			e.group(b, depth, "opt", v.Label, v.Statements)
		case *ast.Break:
			e.group(b, depth, "break", v.Label, v.Statements)
		case *ast.Alt:
			e.branches(b, depth, "alt", func(d int) {
				for i, c := range v.Conditions {
					e.branch(b, d, i, c.Label, c.Statements)
				}
			})
		case *ast.Par:
			e.branches(b, depth, "par", func(d int) {
				for i, br := range v.Branches {
					e.branch(b, d, i, br.Label, br.Statements)
				}
			})
		case *ast.Critical:
			e.counts["critical"]++
			e.block(b, depth, fmt.Sprintf("critical_%d", e.counts["critical"]), v.Label, func(d int) {
				e.emit(b, v.Statements, d)
				for i, o := range v.Options {
					e.block(b, d, fmt.Sprintf("option_%d", i+1), fallbackLabel(o.Label, "option"), func(d int) {
						e.emit(b, o.Statements, d)
					})
				}
			})
		}
	}
}

// group emits one block as a labeled D2 group. label falls back to the block
// kind when the Mermaid block has no description.
func (e *seqEmitter) group(b *strings.Builder, depth int, kind, label string, stmts []ast.SeqStmt) {
	e.counts[kind]++
	e.block(b, depth, fmt.Sprintf("%s_%d", kind, e.counts[kind]), fallbackLabel(label, kind), func(d int) {
		e.emit(b, stmts, d)
	})
}

// branches emits a multi-branch block (alt/else, par/and) as one outer group,
// so the frame tying the branches together survives the conversion; fn fills it
// with one branch child per Mermaid branch.
func (e *seqEmitter) branches(b *strings.Builder, depth int, kind string, fn func(depth int)) {
	e.counts[kind]++
	e.block(b, depth, fmt.Sprintf("%s_%d", kind, e.counts[kind]), kind, fn)
}

// branch emits one branch of a multi-branch block. Keys are scoped to the
// parent group, so they restart at 1 in every block.
func (e *seqEmitter) branch(b *strings.Builder, depth, i int, label string, stmts []ast.SeqStmt) {
	e.block(b, depth, fmt.Sprintf("case_%d", i+1), fallbackLabel(label, "case"), func(d int) {
		e.emit(b, stmts, d)
	})
}

// block writes a labeled D2 group whose body is emitted by fn at the next depth.
func (e *seqEmitter) block(b *strings.Builder, depth int, key, label string, fn func(depth int)) {
	indent := strings.Repeat("  ", depth)
	fmt.Fprintf(b, "%s%s: {\n", indent, key)
	fmt.Fprintf(b, "%s  label: %s\n", indent, d2Label(label))
	fn(depth + 1)
	fmt.Fprintf(b, "%s}\n", indent)
}

func fallbackLabel(label, kind string) string {
	if strings.TrimSpace(label) == "" {
		return kind
	}
	return label
}
