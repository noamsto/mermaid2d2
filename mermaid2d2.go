// Package mermaid2d2 converts between Mermaid and D2 diagram syntax.
//
// Conversion is source-to-source in both directions and driven by two
// functions: [MermaidToD2] and [D2ToMermaid]. Neither renders anything; the
// output is diagram source for the other tool to lay out.
//
// Mermaid flowchart, sequenceDiagram, stateDiagram-v2, classDiagram,
// erDiagram, mindmap, and the C4 family convert to D2. Diagram types with no
// D2 graph analog (gantt, pie, journey, xychart, gitGraph) return an error
// rather than a mangled approximation.
//
// The reverse direction infers the Mermaid diagram type from the D2 shapes
// present: sql_table shapes become an erDiagram, class shapes a classDiagram,
// and anything else a flowchart.
//
// Both directions are lossy where the formats disagree, and each function
// documents what it drops.
package mermaid2d2
