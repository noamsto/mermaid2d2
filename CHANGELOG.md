# Changelog

## v0.2.0

Bidirectional conversion between D2 and Mermaid as a Go library and the `m2d2`
CLI. Pre-1.0: the parser dependency is pinned to a fork of `sammcj/mermaid-check`
(see [#2](https://github.com/noamsto/mermaid2d2/issues/2)) and the output format
is still settling, so 1.0.0 is deferred until that fork is dropped for a tagged
upstream release and the feature surface stabilizes.

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
  diamonds, notes → `tooltip` on their target state.
- **classDiagram** — classes → `sql_table`-style `class` shapes with typed
  members and visibility; relationships → connections with UML arrowheads;
  notes → `tooltip` on their target class.
- **erDiagram** — entities → `sql_table` shapes (typed columns, PK/FK/UK
  constraints), relationships → connections with crow's-foot cardinality labels.
- **mindmap** — tree of nodes connected to their parents, with mapped shapes.

Labels containing D2 syntax characters are automatically quoted. Diagram types
with no D2 equivalent return a clear unsupported-type error.
