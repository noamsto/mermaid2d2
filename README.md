# mermaid2d2

Convert between [D2](https://d2lang.com) and [Mermaid](https://mermaid.js.org)
diagram syntax, as a Go library and a CLI (`m2d2`).

## Status

| Direction | Status |
|---|---|
| **D2 → Mermaid** | Implemented — graph/flowchart diagrams (nodes, containers, edges, direction) |
| **Mermaid → D2** | Not yet implemented (`ErrNotImplemented`) |

D2 → Mermaid maps D2's node/container/edge graph onto a Mermaid `flowchart`:
containers become `subgraph`s, connections become edges, and the board
direction sets the orientation. D2 features with no flowchart equivalent (SQL
tables, class shapes, grids, styling) are dropped.

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
```

## Development

The repo ships a Nix flake with a Go devShell, formatting, and pre-commit hooks:

```sh
direnv allow   # or: nix develop
go test ./...
```
