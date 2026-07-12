package mermaid2d2

import (
	"fmt"
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
// encode the UML relation type. Diagram types with no D2 equivalent (gantt, pie,
// journey, ...) return an error rather than being mangled. Mermaid features
// without a D2 equivalent (notes) are dropped.
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
	default:
		return "", fmt.Errorf("mermaid2d2: unsupported diagram type %q: only flowchart, sequence, state, and class are supported", diagram.GetType())
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
		byPath:      map[string]*d2Container{},
		usedSlugs:   map[string]bool{},
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
	e.emit(&b, root, 0)
	return b.String()
}

type flowEmitter struct {
	containerOf map[string]string // node id -> container path (first mention wins)
	labelOf     map[string]string // node id -> label
	shapeOf     map[string]string // node id -> D2 shape keyword (non-default shapes only)
	nodeOrder   []string          // node ids in first-seen order
	byPath      map[string]*d2Container
	usedSlugs   map[string]bool
	edges       []d2Edge
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
		shape, hasShape := e.shapeOf[id]
		switch {
		case hasLbl && hasShape:
			fmt.Fprintf(b, "%s%s: %s {shape: %s}\n", indent, id, lbl, shape)
		case hasShape:
			fmt.Fprintf(b, "%s%s: {shape: %s}\n", indent, id, shape)
		case hasLbl:
			fmt.Fprintf(b, "%s%s: %s\n", indent, id, lbl)
		}
	}

	for _, child := range c.children {
		tail := strings.TrimPrefix(child.path, c.path+".")
		if c.path == "" {
			tail = child.path
		}
		if child.label != "" {
			fmt.Fprintf(b, "%s%s: %s {\n", indent, tail, child.label)
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
			line += ": " + ed.label
		}
		if ed.style != "" {
			line += " {" + ed.style + "}"
		}
		fmt.Fprintln(b, line)
	}
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
			fmt.Fprintf(&b, "%s: %s\n", p.ID, p.Alias)
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
				fmt.Fprintf(b, "%s%s -> %s: %s\n", indent, v.From, v.To, txt)
				continue
			}
			fmt.Fprintf(b, "%s%s -> %s\n", indent, v.From, v.To)
		case *ast.Loop:
			e.group(b, depth, "loop", v.Label, v.Statements)
		case *ast.Opt:
			e.group(b, depth, "opt", v.Label, v.Statements)
		case *ast.Break:
			e.group(b, depth, "break", v.Label, v.Statements)
		case *ast.Alt:
			for _, c := range v.Conditions {
				e.group(b, depth, "alt", c.Label, c.Statements)
			}
		case *ast.Par:
			for _, br := range v.Branches {
				e.group(b, depth, "par", br.Label, br.Statements)
			}
		case *ast.Critical:
			e.group(b, depth, "critical", v.Label, v.Statements)
			for _, o := range v.Options {
				e.group(b, depth, "option", o.Label, o.Statements)
			}
		}
	}
}

// group emits one block as a labeled D2 group. label falls back to the block
// kind when the Mermaid block has no description.
func (e *seqEmitter) group(b *strings.Builder, depth int, kind, label string, stmts []ast.SeqStmt) {
	if strings.TrimSpace(label) == "" {
		label = kind
	}
	e.counts[kind]++
	indent := strings.Repeat("  ", depth)
	fmt.Fprintf(b, "%s%s_%d: {\n", indent, kind, e.counts[kind])
	fmt.Fprintf(b, "%s  label: %s\n", indent, label)
	e.emit(b, stmts, depth+1)
	fmt.Fprintf(b, "%s}\n", indent)
}
