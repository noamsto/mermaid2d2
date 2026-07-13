# mermaid2d2

Convert between [D2](https://d2lang.com) and [Mermaid](https://mermaid.js.org)
diagram syntax, as a Go library and a CLI (`m2d2`).

## Coverage

### Mermaid → D2

| Mermaid | D2 target | Mapped features |
|---|---|---|
| `flowchart` / `graph` | shapes, connections, containers | subgraphs → containers (membership preserved via qualified edge endpoints); node shapes (rhombus → `diamond`, hexagon, circle, cylinder, stadium → `oval`); dotted (`-.->`) / thick (`==>`) edges → stroke style; direction |
| `sequenceDiagram` | `sequence_diagram` | participants + aliases; messages; `loop`/`alt`/`opt`/`par`/`critical` → labeled groups; `Note` → D2 note |
| `stateDiagram-v2` | nodes + connections | composite states → containers; `[*]` → sentinel start/end circles; transitions → connections; choice → `diamond` |
| `classDiagram` | `class` shapes | typed members with visibility; relationships → connections with UML arrowheads (inheritance, composition, aggregation, …) |
| `erDiagram` | `sql_table` shapes | typed columns; PK/FK/UK → column constraints; crow's-foot cardinalities → arrowhead labels |
| `mindmap` | node tree | nodes connected to their parent; mapped node shapes |
| pie, gantt, journey, C4, … | — | unsupported → error |

### D2 → Mermaid

| D2 | Mermaid target | Mapped features |
|---|---|---|
| nodes, containers, connections, direction | `flowchart` | containers → `subgraph`s; connections → edges; board direction → orientation |
| sql_table, class shapes, grids, styling | — | dropped (no `flowchart` equivalent) |

**Limitations.** Labels containing D2 syntax characters are quoted automatically.
State/class notes, sequence note side (left/right) and multi-participant spans,
and mindmap icons are dropped.

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
