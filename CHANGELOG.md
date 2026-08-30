# Changelog

Pre-1.0: the output format is still settling, so 1.0.0 is deferred until the
feature surface stabilizes.

## v0.5.0

Rendering fixes across four diagram types. Each was found by putting the
examples side by side in the README, and each changes what `m2d2` emits for
input it already accepted.

### Fixed
- **flowchart** — a node first mentioned outside a subgraph and then used
  inside its body stayed outside the D2 container, so the group box wrapped
  only part of the graph. Membership now follows the subgraph whose body
  mentions the node; between two subgraphs the first still wins.
- **sequenceDiagram** — each `alt`/`else`, `par`/`and` and `critical`/`option`
  branch became its own top-level group, losing the frame that ties them
  together. Branches now nest inside one outer group per block.
- **classDiagram** — UML markers landed on the wrong end, and for composition
  and aggregation did not render at all:
  - `Entity <|-- Order` drew its triangle at `Order`. The parser records the
    classes in written order and the relation type but not which side the
    glyph was on, so the operator is now recovered from the source line.
  - D2 draws an arrowhead only on the end its arrow points at, so a
    `source-arrowhead` on a `->` connection was silently ignored — every
    `*--` and `o--` lost its diamond. The arrow is flipped to `<-` when the
    marker belongs at the source.
  - D2 defaults a triangle arrowhead to filled and a diamond to hollow, so
    composition carried aggregation's notation. Fill is now always explicit.
  - Left-pointing associations (`A <-- B`) pointed the wrong way; the same
    change reverses them.

  Because D2 lays out along arrow direction, a superclass now sits below its
  subclass where Mermaid draws it above. Correct arrowheads cost the vertical
  order.
- **stateDiagram-v2** — `[*]` was a full-size circle with `start` or `end`
  written inside it. It is now a small unlabelled dot, and the terminal state a
  ring, as UML draws them. Neither carries an explicit colour, so
  `d2 --dark-theme` still recolours them.

### Added

- Mermaid `style A fill:#f9f` becomes `style.*` attributes on the D2 node. A
  `style` naming something that is not a drawn node is dropped rather than
  inventing one, since Mermaid also allows it to name an edge.
- Mermaid's `A:::hot` shorthand now converts. It was silently dropping the
  whole statement, edge included — a `mermaid-check` parser bug, fixed
  upstream in sammcj/mermaid-check#19 and picked up with the v0.4.1 bump.
  Emitting the shorthand as a class assignment meant no change was needed
  here.
- The reverse direction still drops inline `style.*`; a D2 node styled
  directly comes back as a plain Mermaid node.

### Docs
- The README examples are Mermaid fences GitHub renders itself, paired with the
  D2 `m2d2` produces and its render. `generate.sh` no longer needs mermaid-cli,
  and renders with `d2 --dark-theme` so the SVGs follow the reader's theme.
- `TestREADMEExamplesInSync` fails when a conversion change lands without a
  `generate.sh` run, since the README quotes generated D2 verbatim.

## v0.4.0

- classDiagram static (`$`) and abstract (`*`) member classifiers survive both
  directions. D2 has no notion of either, so they ride along in the member
  name; `d2compiler` drops a member whose key ends in a classifier, so such a
  key is quoted.

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
