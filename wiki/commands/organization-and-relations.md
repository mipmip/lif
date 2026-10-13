# Organization and relations

This page documents the commands that reshape existing items: editing fields and body text, tagging, ordering siblings, and renaming IDs or slugs. It complements the creation and lifecycle commands documented elsewhere in this wiki.

## `lif update`

`lif update <id> [id...]` (alias `u`) changes one or more properties of one or more existing items. Every flag applies to every item named in the call. It requires at least one change-producing flag; running it with none returns a validation error listing the accepted flags. Every ID is resolved, and `--status`'s required fields and `--parent`'s type and cycle rules are checked for every item, before the first item is written; naming the same item twice is an error.

Field flags are `--status`, `--type`, `--priority`, `--title`, each validated against the project's configured statuses, types, and priorities before anything is written; an empty `--title` is rejected. `-p`/`--priority` accepts an empty string to clear the priority.

Body flags come in two mutually exclusive groups. The first group replaces the whole body: `-d`/`--body` (use `-` to read from stdin) or `--body-file` to read from a file. The second group edits the existing body in place: `--body-replace-old` together with `--body-replace-new` (both required together) finds and replaces one span of text, and `--body-append` (use `-` for stdin) appends text to the end. `--body-replace-old` and `--body-append` can be combined in the same call and both apply; neither can be combined with `--body`/`--body-file`.

Relationships use paired add/remove flags rather than full replacement: `--parent` sets a parent item ID and `--remove-parent` clears it (mutually exclusive with each other); `--blocking`/`--remove-blocking` and `--blocked-by`/`--remove-blocked-by` each accept a repeatable item ID to add or remove one relationship at a time. Tags follow the same pattern: `--tag`/`--remove-tag` (repeatable).

`--set key=value` (repeatable) writes an arbitrary extra front matter key, and `--unset key` (repeatable) removes one; both count as a change even though they bypass the typed fields above, and a key that names a managed field (`--set status=…`) is rejected with a pointer to the matching flag. `--if-match <etag>` makes the update conditional: it fails instead of writing if the item's current etag does not match, which is the command's only optimistic-locking guard — `update` does not otherwise detect concurrent edits. An etag belongs to one item, so `--if-match` accepts exactly one ID.

For each item, all of the call's field changes, body changes, relationship changes, and extra-key operations land in a single write under one etag, so a status change and any extra front matter fields required by policy land together rather than being split across two calls. `--json` returns the resulting item in the same shape as `lif show --json`, not a wrapped success envelope: one object for a single ID, an array of objects for several. Updating an item whose file is under `archive/` changes it in place; it does not move the file back to the active directory.

```
lif update lif-w1z6 --status in-progress --tag backend
lif update lif-w1z6 --body-replace-old "Initial" --body-replace-new "Updated" --body-append "

More context."
lif update lif-w1z6 --blocked-by lif-xkih --parent lif-e1ta
```

## `lif tag`

`lif tag <id> [id...]` adds and removes tags on one or more items in a single invocation. `--tag` (repeatable) adds tags, `--remove-tag` (repeatable) removes them; any tag not named by either flag is left untouched, and at least one of the two flags is required. Every given ID is resolved before the first item is written, so a typo'd or missing ID among several fails the whole call before any file changes — but once writing starts, each item is written independently, so a failure partway through (for example an etag conflict on a later ID) leaves the items already written in their new tagged state.

`--json` prints the written items as JSON: one item object for a single ID, an array of item objects for several.

```
lif tag lif-w1z6 lif-e1ta --tag backend --tag urgent
lif tag lif-w1z6 --remove-tag urgent
```

## `lif order`

`lif order <id>` sets an item's manual order key relative to its siblings under the same parent, using a fractional index. A move only ever writes the one item file being placed, never the neighboring siblings, and ordering is scoped strictly per parent: `--after`/`--before` must name a sibling that shares the target item's parent.

`--json` prints the placed item as a JSON object.

Exactly one placement flag is required, and the four are mutually exclusive: `--first` (before all siblings), `--last` (after all siblings), `--after <id>` (immediately after that sibling), or `--before <id>` (immediately before that sibling). `--after`/`--before` require the named sibling to already carry an explicit order value — placing relative to a sibling that has never been ordered fails, and so does `--last` while no sibling carries an order yet. This means the first placement in a sibling group that has no order values must be `--first`; after that `--last`, `--after` and `--before` can be used to place and interleave the rest. Siblings without an order value sort after every ordered sibling.

```
lif order lif-w1z6 --first
lif order lif-e1ta --last
lif order lif-xkih --before lif-w1z6
```

## `lif rename`

`lif rename` renames items in exactly one of three modes, chosen by which flags and positional arguments are given; supplying flags for more than one mode is rejected.

Slug mode changes only the filename's human-readable slug, leaving the item's ID untouched: `lif rename <id> --slug "new-slug"` sets it explicitly, `lif rename <id> --no-slug` clears it so the filename becomes `<id>.md` (`<id>--.md` when the ID contains a hyphen, as every prefixed ID does, so the ID reads back unchanged), and `lif rename <id> --reslug` regenerates the slug from the item's current title.

ID mode changes the item's ID itself and cascades: `lif rename <id> <new-id>` gives the item a full new ID, and `lif rename <id> --suffix <new-suffix>` keeps the project's configured prefix and only replaces the suffix (it refuses to run against an ID that does not already start with that prefix). Both forms update every reference to the old ID found in the `parent`, `blocking` and `blocked_by` front matter of other items as part of the same rename. Text inside item bodies is not rewritten: an ID that a body mentions keeps its old value. A new ID that already exists is rejected as a collision.

Prefix mode, `lif rename --prefix "new-prefix-"`, rebrands every item ID in the project to the new prefix in one pass, updates every cross-reference in the same relationship fields, and writes the new prefix to `lif.prefix` in the config file. It additionally requires `--yes` or an interactive y/N confirmation before applying, and refuses to run while a `lif-serve` process is active or while active git worktrees exist for the project.

`--dry-run` prints the planned changes for any of the three modes without writing anything; combine it with `--json` to get the plan as structured data. Without `--dry-run`, a successful rename reports how many items changed.

```
lif rename lif-w1z6 --slug "endpoint-handler"
lif rename lif-w1z6 --reslug
lif rename lif-w1z6 bew-1234
lif rename lif-w1z6 --suffix k7x2
lif rename --prefix "bew-" --dry-run
```

## Related documentation

- [Planning and reporting](planning-and-reporting.md)
