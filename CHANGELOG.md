# Changelog

Pre-1.0: the output format is still settling, so 1.0.0 is deferred until the
feature surface stabilizes.

## v0.3.1

- `m2d2 -version` prints the version. Release builds stamp it in; a
  `go install` build reports the module version.
- Corrected the package doc, which still described both directions as
  flowchart-only.

## v0.3.0

Completes the D2 → Mermaid direction (it emitted only flowcharts in v0.2.0),
adds C4 support in the Mermaid → D2 direction, and moves off the
`mermaid-check` fork onto upstream.

### D2 → Mermaid
- `sql_table` shapes → `erDiagram` entities (typed columns, PK/FK/UK
  constraints) and relationships (crow's-foot cardinality labels).
- `class` shapes → `classDiagram` classes (typed members, visibility) and
  relationships (UML arrowheads → inheritance/realization/composition/
  aggregation/dependency/association).
- `classes:`/`class:` styling → flowchart `classDef`/`class`.
- The output diagram type is now detected from the D2 graph. A graph mixing
  `sql_table`/`class` shapes with each other or with plain nodes/containers has
  no single Mermaid diagram type and returns an error.

### Mermaid → D2
- **C4** (`C4Context`/`C4Container`/`C4Component`/`C4Dynamic`/`C4Deployment`) —
  elements (`Person`, `System`, `Container`, `Component`, `Node`) → D2 nodes
  (`Person` → `person` shape, `Db` variants → `cylinder`, `Queue` variants →
  `queue`, `_Ext` variants → a dashed border); boundaries → containers;
  `Rel`/`BiRel`/`Rel_Back`/etc. → connections. D2 has no native C4 notation,
  so this mapping is lossy: the diagram title, sprites, tags, links, and
  `UpdateElementStyle`/`UpdateRelStyle` overrides are dropped.
- **stateDiagram-v2** — notes → `tooltip` on their target state.
- **classDiagram** — notes → `tooltip` on their target class, or a floating
  node when standalone.

### Fixed
- erDiagram/classDiagram identifiers are sanitized, and `stroke-dasharray`
  no longer carries a `px` suffix Mermaid rejects.
- A floating note no longer takes an id that collides with a same-named class.

### Dependencies
- Now depends on upstream `sammcj/mermaid-check` v0.3.1 instead of a fork; the
  three parser fixes the fork carried are upstream as of that release.
- **Behavior change** — an unclosed composite state (`state Foo {` with no
  closing brace) is now a parse error. It previously swallowed the rest of the
  diagram silently.

## v0.2.0

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
