# mermaid2d2

Convert between [D2](https://d2lang.com) and [Mermaid](https://mermaid.js.org)
diagram syntax, as a Go library and a CLI (`m2d2`).

## Status

| Direction | Status |
|---|---|
| **D2 → Mermaid** | Implemented — graph/flowchart diagrams (nodes, containers, edges, direction) |
| **Mermaid → D2** | Implemented — flowchart (nodes, subgraphs, edges, direction) and sequence (participants, messages, loop/alt/opt groups) |

D2 → Mermaid maps D2's node/container/edge graph onto a Mermaid `flowchart`:
containers become `subgraph`s, connections become edges, and the board
direction sets the orientation. D2 features with no flowchart equivalent (SQL
tables, class shapes, grids, styling) are dropped.

Mermaid → D2 maps a Mermaid `flowchart` onto D2 shapes, connections, and
containers (subgraphs become containers; a node's subgraph membership is
preserved via qualified edge endpoints), and a `sequenceDiagram` onto a D2
`sequence_diagram` (loop/alt/opt/par blocks become labeled groups; `Note`s
become D2 notes scoped to their first participant). Node shapes
with a D2 equivalent are mapped (rhombus → `diamond`, hexagon, circle, cylinder,
stadium → `oval`); shapes without one fall back to the default rectangle. Dotted
(`-.->`) and thick (`==>`) links carry their line style onto the D2 connection.
A `stateDiagram-v2` maps onto D2 nodes and connections (composite states become
containers, `[*]` start/end become sentinel circle nodes, choices become
diamonds). A `classDiagram` maps onto D2 `class` shapes, with relationships
becoming connections whose arrowheads encode the UML relation type (inheritance,
composition, aggregation, …). An `erDiagram` maps onto D2 `sql_table` shapes
(attributes become typed columns; PK/FK/UK keys become column constraints;
relationship cardinalities become arrowhead labels). State and class notes have
no D2 equivalent and are dropped; diagram types other than flowchart, sequence,
state, class, and er return an error.

## CLI

```sh
go install github.com/noamsto/mermaid2d2/cmd/m2d2@latest

m2d2 diagram.d2                 # -> Mermaid on stdout (target inferred from .d2)
m2d2 diagram.d2 -o diagram.mmd  # write to a file
cat diagram.d2 | m2d2 -to mermaid -   # read from stdin
```

The target format is inferred from the input extension (`.d2` → mermaid,
`.mmd`/`.mermaid` → d2) unless `-to` is given.

## Library

```go
import "github.com/noamsto/mermaid2d2"

out, err := mermaid2d2.D2ToMermaid(src)
d2, err := mermaid2d2.MermaidToD2(src)
```

## Development

The repo ships a Nix flake with a Go devShell, formatting, and pre-commit hooks:

```sh
direnv allow   # or: nix develop
go test ./...
```
