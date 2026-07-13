package mermaid2d2

import (
	"fmt"
	"strings"

	"github.com/sammcj/mermaid-check/ast"
)

// stateDiagramToD2 emits D2 for a Mermaid state diagram.
//
// States become D2 nodes (composite states become containers, recursing into
// their nested statements) and transitions become connections. D2 has no native
// start/end markers, so Mermaid's [*] is rendered with sentinel circle nodes:
// per scope, a single `start` node fans out to every start target and a single
// `end` node collects every end source. Choice nodes become diamonds; fork and
// join nodes become plain nodes (D2 has no bar shape). Notes and comments are
// dropped, matching how the other converters treat features with no D2 analog.
func stateDiagramToD2(d *ast.StateDiagram) string {
	var b strings.Builder
	emitStateStmts(&b, d.Statements, 0)
	return b.String()
}

// emitStateStmts renders one scope of state statements. The start/end sentinels
// are declared lazily on first use so each scope (the root board or a composite
// container) gets its own uniquely keyed pair.
func emitStateStmts(b *strings.Builder, stmts []ast.StateStmt, depth int) {
	indent := strings.Repeat("  ", depth)
	var startDone, endDone bool
	for _, s := range stmts {
		switch v := s.(type) {
		case *ast.State:
			if v.IsComposite {
				if v.Description != "" && v.Description != v.ID {
					fmt.Fprintf(b, "%s%s: %s {\n", indent, v.ID, d2Label(v.Description))
				} else {
					fmt.Fprintf(b, "%s%s: {\n", indent, v.ID)
				}
				emitStateStmts(b, v.Nested, depth+1)
				fmt.Fprintf(b, "%s}\n", indent)
				continue
			}
			if v.Description != "" && v.Description != v.ID {
				fmt.Fprintf(b, "%s%s: %s\n", indent, v.ID, d2Label(v.Description))
				continue
			}
			fmt.Fprintf(b, "%s%s\n", indent, v.ID)
		case *ast.Transition:
			line := fmt.Sprintf("%s%s -> %s", indent, v.From, v.To)
			if lbl := strings.TrimSpace(v.Label); lbl != "" {
				line += ": " + d2Label(lbl)
			}
			fmt.Fprintln(b, line)
		case *ast.StartState:
			if !startDone {
				fmt.Fprintf(b, "%sstart: {shape: circle}\n", indent)
				startDone = true
			}
			fmt.Fprintf(b, "%sstart -> %s\n", indent, v.To)
		case *ast.EndState:
			if !endDone {
				fmt.Fprintf(b, "%send: {shape: circle}\n", indent)
				endDone = true
			}
			fmt.Fprintf(b, "%s%s -> end\n", indent, v.From)
		case *ast.Choice:
			fmt.Fprintf(b, "%s%s: {shape: diamond}\n", indent, v.ID)
		case *ast.Fork:
			fmt.Fprintf(b, "%s%s\n", indent, v.ID)
		case *ast.Join:
			fmt.Fprintf(b, "%s%s\n", indent, v.ID)
		}
	}
}
