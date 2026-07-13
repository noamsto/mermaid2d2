# Changelog

## v1.0.0

First stable release. Bidirectional conversion between D2 and Mermaid as a Go
library and the `m2d2` CLI.

### D2 → Mermaid
- Graph/flowchart diagrams: nodes, containers → `subgraph`s, connections, board
  direction.

### Mermaid → D2
- **flowchart** — nodes, `subgraph`s → containers (subgraph membership preserved
  via qualified edge endpoints), edges with direction and labels, board
  direction. Node shapes with a D2 equivalent (rhombus → `diamond`, hexagon,
  circle, cylinder, stadium → `oval`); dotted (`-.->`) and thick (`==>`) links
  carry their line style.
- **sequenceDiagram** — participants (with aliases), messages, `loop`/`alt`/
  `opt`/`par`/`critical` blocks as labeled groups, and `Note`s as D2 notes.
- **stateDiagram-v2** — states → nodes, composite states → containers,
  transitions → connections, `[*]` → sentinel start/end circles, choices →
  diamonds.
- **classDiagram** — classes → `sql_table`-style `class` shapes with typed
  members and visibility; relationships → connections with UML arrowheads.
- **erDiagram** — entities → `sql_table` shapes (typed columns, PK/FK/UK
  constraints), relationships → connections with crow's-foot cardinality labels.
- **mindmap** — tree of nodes connected to their parents, with mapped shapes.

Labels containing D2 syntax characters are automatically quoted. Diagram types
with no D2 equivalent return a clear unsupported-type error.
