# Planning and reporting

This page documents the read-only commands that summarize a project's structure and status: roadmap views, milestone rollups, aggregate progress, and relationship graphs.

## `lif roadmap`

`lif roadmap [id]` displays a roadmap of milestones, epics, and their child items, walking the parent/child hierarchy rather than any single flat list. With no ID argument it renders the entire roadmap, rooted at every top-rank container (e.g. milestones). With an ID argument naming a milestone, epic, or feature, it scopes the output to that item's subtree only; `--status` and `--no-status` cannot be combined with an ID argument, since scoping to one subtree and filtering the top-level milestone list are mutually exclusive concerns.

`--depth <n>` limits how many levels below the roadmap's root are rendered, following the `tree -L n` convention where the root itself never counts. Without an ID argument the root is the roadmap as a whole, so `--depth 1` lists milestones only; with an ID argument the root is that item, so `--depth 1` lists its direct children. `--status <name>` (repeatable) and `--no-status <name>` (repeatable) filter which milestones appear by status, and `--include-done` includes completed items that are otherwise omitted.

`--view` chooses the layout used for terminal output: `tree` (default) nests items under their containers, `table` lists them flat and sortable. The `--view` choice only affects terminal rendering — Markdown output always uses the fixed Markdown template regardless of `--view`.

`--format` chooses the output mode explicitly: `tty` for the colored terminal tree/table, or `markdown` for a plain Markdown document with headings per milestone/epic and a list per leaf item. Left unset, the format is auto-detected from whether stdout is a terminal. In Markdown mode, `--no-links` renders item IDs as plain text instead of Markdown links, `--link-prefix <url>` sets their URL prefix, and `--tags` includes tags. `--max-width <n>` caps terminal output. `--json` instead returns the roadmap domain model as structured data and bypasses terminal/Markdown rendering.

The `--json` document has a `milestones` array. Each entry holds the `milestone` item, an ordered `containers` array whose entries carry either an `epic` group or a `feature` group (the keys are slot names for rank 2 and rank 3, whatever those types are called in the project), and an `other` array of items without a container. An `unscheduled` group of the same shape appears when items sit outside every milestone, and a roadmap rooted at one epic or feature adds a `root` group.

```
lif roadmap
lif roadmap --view table --format tty
lif roadmap lif-xkih --depth 1
lif roadmap --format markdown --tags --no-links > ROADMAP.md
lif roadmap --status in-progress --no-status scrapped
```

## `lif milestones`

`lif milestones` lists every item on the top container rank (e.g. milestones), each annotated with how many of its descendants — through any number of parent levels, such as an epic's tasks — are completed. Completed and scrapped milestones are hidden by default; `--all` includes them. Descendants hidden by a status/archive policy do not contribute to the completed/total counts, so a hidden subtree under a visible milestone cannot inflate or deflate its progress figure.

`--view` chooses the arrangement: `table` (default) lists milestones flat and sortable, `tree` nests them. `--tags` renders each milestone's tags. `--max-width <n>` caps the rendered width (`0` disables the cap; otherwise falls back to the `display.max_width` config value, or 110).

`--json` returns an array with one entry per milestone, each shaped `{ "item": { … }, "completed": <n>, "total": <n> }`.

```
lif milestones
lif milestones --all --view tree --tags
```

## `lif progress`

`lif progress` shows counts by status across every configured status, plus a percent-complete figure computed as `completed / (total - scrapped)`, truncated toward zero. An `<id>` argument scopes the counts to a single item's descendants (for example a milestone or epic) instead of the whole workspace; the root item's own status is not counted, since it is the container, not an item of work. `--json` returns the per-status counts plus the derived `completed`, `total`, and `percent` fields, and adds a `root` field (the resolved full ID) when scoped; the plain-text form additionally renders a fixed-width bar under the counts.

```
lif progress
lif progress lif-xkih
lif progress --json
```


## `lif graph`

`lif graph [id]` prints parent and blocking relationships. The default Graphviz DOT output can be piped into tools such as `dot -Tpng`; `--format ascii` prints a terminal edge list, `--format mermaid` a Mermaid flowchart, and `--format json` returns `nodes` and `edges`.

`--format mermaid` writes a `flowchart LR` that pastes into any Markdown document that renders Mermaid, which is what makes a blocking chain readable in a document rather than only on a terminal. Node handles keep the item id, hyphen included, so a handle in the diagram can be looked up in the store. Since `lif.prefix` is free-form configuration, a character that Mermaid cannot carry in an identifier — a space or a quote, for instance — becomes an underscore in the handle only; the label always shows the id as written. Label text is escaped only where Mermaid needs it, which is less than it looks: a double quote would close the label, a `<` would be drawn as markup because labels are rendered as HTML, and `#` and `&` each open an entity that a title may spell out literally. Brackets, parentheses and braces are left as written — inside a quoted label Mermaid takes them verbatim, and their HTML entity form is actively wrong there, since Mermaid resolves the `#91;` inside `&#91;` and leaves the ampersand behind. Edges carry their relation as the arrow label, so `parent` and `blocks` stay distinguishable, and each status becomes a `classDef` with its configured colour, which is the DOT output's `fillcolor` expressed the way Mermaid allows. Status names are configuration too, so they pass through the same handle guard as the ids — a name Mermaid could not carry as an identifier would otherwise split the `class` and `classDef` lines apart. One limit is Mermaid's own rather than this command's: the renderer refuses a diagram above 500 edges by default, and a whole store easily passes that, so scope the graph with an item id, a `--depth` or a `--relation` instead of asking for everything. Combined with `--relation blocks`, this is the direct route from the store to a dependency diagram:

```
lif graph lif-xkih --format mermaid --relation blocks --depth 0
```

Without an ID the command includes the complete store. Naming one item scopes the graph to its neighborhood: `--depth 1` includes its direct relationships, larger values widen the traversal by hops, and `--depth 0` walks its whole connected component. `--relation parent` and `--relation blocks` can be repeated to restrict edge kinds. Broken links and self-links are omitted here and reported by `lif check`.

```
lif graph
lif graph lif-xkih --format ascii
lif graph lif-xkih --depth 2 --relation parent
lif graph --format json
lif graph lif-xkih --format mermaid --relation blocks
```

## Related documentation

- [Organization and relations](organization-and-relations.md)
- [Querying and automation](querying-and-automation.md)
