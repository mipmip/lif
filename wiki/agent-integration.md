# Agent Integration

This page covers how to wire a coding agent up to `lif` — the generic instruction, the two integrations Lif currently ships support for, and the safe read/write patterns an agent should follow once it's connected. It supplements the exact CLI syntax documented per-command elsewhere in this wiki.

## The one rule that matters

Never let an agent rely on stale or remembered knowledge of a project's Lif setup. Every project can configure its own types, statuses, priorities, required-fields policy, and commit-field behavior in `.lif.yml`, and `lif prime` is the single command that reports what is actually configured in *this* project, right now:

```bash
lif prime
```

`lif prime` renders a full agent-facing usage guide, generated from a template filled in with the running project's real configuration (its types, statuses, priorities, hierarchy ranks, and any `require_fields_on` policy) — it is not static prose. An agent should treat `lif prime`'s output as authoritative over any generic knowledge about Lif from training data or a previous session, because the previous project it worked in may have had different types, a different commit-field policy, or no policy at all. When no Lif config file can be found from the current directory, `lif prime` silently exits with no output, which is the intended way to detect "this isn't a Lif project" without an error.

The minimal, tool-agnostic way to get an agent to run it is one line in the project's `AGENTS.md`, `CLAUDE.md`, or equivalent instruction file:

```
**IMPORTANT**: before you do anything else, run the `lif prime` command and heed its output.
```

## Claude Code

Claude Code's `SessionStart` and `PreCompact` hooks can invoke `lif prime` automatically so the agent gets current project context at the start of a session and again after context compaction, add this to the project's `.claude/settings.json`:

```json
{
  "hooks": {
    "SessionStart": [
      { "hooks": [{ "type": "command", "command": "lif prime" }] }
    ],
    "PreCompact": [
      { "hooks": [{ "type": "command", "command": "lif prime" }] }
    ]
  }
}
```

A packaged Claude Code plugin for this exists at `extras/claude/plugins/lif-prime/` (the same hook configuration, plus a plugin manifest), published through the `lif` marketplace defined in `.claude-plugin/marketplace.json`; its plugin ID is `lif-prime@lif`. It saves a project from hand-editing `.claude/settings.json`. As of the current state that plugin is explicitly a work in progress and not ready for general use — its own README says so — so the manual hook configuration above is the supported path until it lands.

## OpenCode

OpenCode integrates through a plugin that injects `lif prime`'s output into the chat system prompt on session start and again on session compaction. The plugin ships in-repo at `.opencode/plugin/lif-prime.ts` (it exports `LifPrimePlugin`); to use it, copy that file into a project's `.opencode/plugin/` directory (or `~/.opencode/plugin/` for a personal, cross-project install), no build step required. The plugin only injects context when both `lif` is on `PATH` and the target directory holds a Lif config file (`.lif.yml`, or a legacy config file name, see [Legacy names](configuration.md#legacy-names)) — otherwise it silently does nothing, mirroring `lif prime`'s own silent no-op.

## Agents without a supported hook

Any agent that can run a shell command and read `AGENTS.md`/`CLAUDE.md` can use `lif` even without a dedicated hook or plugin: put the one-line instruction from above in that file, and the agent will run `lif prime` itself at the start of its work. This is the same fallback the Claude Code and OpenCode integrations reduce to when the hook or plugin isn't installed — nothing about `lif prime`'s output depends on which integration triggered it.

## Read workflows

Once primed, an agent's normal work loop is: find work, show it, act on it.

```bash
lif next                              # single highest-priority ready item
lif next --type bug --tag cli         # narrowed the same way list is narrowed
lif list --json --ready               # every ready item, not just the top one
lif list --json --is-blocked          # items blocked directly or via a blocked ancestor
lif show --json <id> [id...]          # full details, one or many IDs
lif progress                          # status counts + percent-complete across the workspace
lif progress <id>                     # same, scoped to one item's descendants
lif milestones                        # milestones with descendant completion counts
```

`lif list -S "<keywords>"` (full-text search over title, slug and body, with Bleve query syntax: fuzzy `~`, wildcard `*`, phrase `"..."`, field-scoped `slug:`/`title:`/`body:`) is the way to check whether an item for some piece of work already exists before creating a duplicate one. `list`, `show`, `next`, `progress`, `milestones`, `roadmap` and every mutating command accept `--json` for machine-readable output (`graph` uses `--format json`; `path`, `prime` and `pick` print plain text only). Without `--json`, `lif show` in particular follows stdout — styled on a terminal, raw unpadded/unwrapped Markdown into a pipe or file, so piped `show` output is safe to feed to a parser even without `--json`.

## Write workflows

```bash
lif create --json "Title" -t task -d "Description..." -s todo
lif start <id> [id...]                                   # mark in-progress, show it
lif update --json <id> --blocked-by <other-id>            # THIS item is blocked by another
lif update --json <id> --body-append "## Notes\n\nmore"   # append to body
lif update --json <id> --body-replace-old "- [ ] x" --body-replace-new "- [x] x"
lif tag <id> [id...] --tag cli --remove-tag stale         # merge tags, one call over n items
lif complete <id> [id...] --summary "what was done"       # ONLY once every todo is checked off
lif scrap <id> [id...] --reason "why"                     # --reason is required
```

`complete`/`scrap`/`start`/`update`/`tag`/`delete` all take multiple IDs in one call; every ID is resolved and policy-checked before the first write, so an unknown ID, a repeated ID, or a policy violation in the batch leaves every item in it untouched — there is no partial application to clean up after. A batch call that does fail after writes have already started (mid-batch I/O failure, not a validation failure) returns `{"success": false, "items": [...], "count": n, ...}` naming exactly which items were written before the failure.

`lif complete` and `lif scrap` are the only commands that both change `status` and append a body section (`## Summary of Changes`, `## Reason for Scrapping`) in one atomic write — do not hand-append those sections and then call `--status completed` separately.

## Concurrency

Update calls should not always assume they hold the only writer. Use etags for optimistic locking when a write follows a read that could have gone stale:

```bash
ETAG=$(lif show <id> --etag-only)
lif update <id> --if-match "$ETAG" --status in-progress
```

A mismatched `--if-match` fails the write and reports the current etag rather than silently overwriting a concurrent change.

## JSON and error handling

Every command failure surfaces in exactly one place, and which place depends on whether that command reaches its JSON output path:

- Without `--json`: two lines to stderr, empty stdout — `Error: <message>` followed by `Run '<command> --help' for usage.`.
- With `--json`: one JSON error document to stdout, empty stderr — `{"success": false, "error": "<message>", "code": "<CODE>"}`. Errors that occur before JSON handling, such as loading an unreadable config or missing store, still surface on stderr.

Successful JSON shapes carry domain data, not a wrapper. `list`, `show`, `progress`, and `roadmap` return domain data directly; `check` has its own `success` plus diagnostic fields; and the item-mutating commands (`create`, `update`, `start`, `complete`, `scrap`, `delete`, `tag`, `order`) return the affected item for a single-item call and a bare array for a multi-item one. `init`, `archive`, and a `next` with no match return a message envelope, having no item to name. Agents should branch on process exit status first, parse the command-specific success shape second, and use `error` plus `code` for a JSON failure — not the absence of a `success` key, which a successful item document never has.

Error codes: `NOT_FOUND`, `NO_LIF_DIR`, `INVALID_STATUS`, `FILE_ERROR`, `VALIDATION_ERROR`, `CONFLICT`, `POLICY_VIOLATION`.

## GraphQL for bulk/relational reads

`lif graphql` (aliased `query`) runs a GraphQL query or mutation against the same data `list`/`show`/`update` operate on, and is the better tool once a task needs several fields across many items, or needs to traverse relationships in one round trip instead of one `show` per item:

```bash
# Actionable items with full detail in one call
lif graphql --json '{ items(filter: { excludeStatus: ["completed", "scrapped"], isBlocked: false }) { id title status type body } }'

# An item plus its immediate relationships
lif graphql --json '{ item(id: "<id>") { title body parent { title } children { id title status } } }'

# A single-field mutation
lif graphql --json 'mutation { updateItem(id: "<id>", input: { status: "completed" }) { id status etag } }'

# Print the full schema before writing a non-trivial query
lif graphql --schema
```

`lif graphql --schema` is always the ground truth for field names and argument shapes; do not assume a field exists (or still has its old name) without checking it there first, since the schema is generated from the same Go types the CLI itself uses, not maintained separately. The query root exposes `item(id)` and `items(filter)`; the mutation root exposes `createItem`, `updateItem`, `deleteItem`, `archiveItem`, and the same parent/blocking mutations the CLI flags map to. `ItemFilter` on the query side covers everything `lif list`'s flags cover (status/type/priority include and exclude lists, tag include/exclude, parent/blocking/blocked-by presence and target-ID filters, `isBlocked`/`isExplicitlyBlocked`/`isImplicitlyBlocked`, and `search` using the same Bleve syntax as `--search`).

## Lifecycle safety

- **Never skip `lif complete`'s precondition.** Only run it once every todo item in the item's body is checked off; if work was deferred, create a follow-up item for it first rather than completing early.
- **`lif delete` is destructive and cascades.** It removes incoming references (as parent or via blocking) from every item that named the deleted one, after a confirmation the agent should not bypass with `-f`/`--json` (which implies `-f`) unless the user explicitly asked for an unconditional delete.
- **`lif archive` is not destructive but is one-directional from the CLI's perspective.** It only ever moves `completed`/`scrapped` items into `<lif-path>/archive/`; run it only when the user asks for it, not as a routine part of completing work, since archived items stop showing up in a plain directory listing even though every `lif`/GraphQL query still sees them.
- **A required-fields policy blocks the transition, not the item.** If a project's `.lif.yml` has `lif.require_fields_on`, an unqualified transition into a gated status fails with `POLICY_VIOLATION` naming the missing field(s); `lif prime`'s output for the current project states exactly which fields gate which statuses, including whether the configured commit field is one of them — check there rather than guessing a policy from a different project.
- **Prefer `--blocked-by` when creating new dependent work.** It lets the new item describe its own dependency without mutating the item it depends on, and keeps write scope to the item actually being changed.
- **Include item files in the same commit as the code change they track.** Both new and modified item files, including ones marked completed or scrapped as part of the change, belong in the same commit as the code — not a separate housekeeping commit.
- **Any exploratory or example invocation that writes data must target a throwaway store**, e.g. `lif --lif-path /tmp/scratch-lif create ...`, never the project's real `.lif` directory or a shared `LIF_PATH`.

## Related documentation

- [wiki/data-model.md](data-model.md)
- [wiki/configuration.md](configuration.md)
- [wiki/commands/lifecycle.md](commands/lifecycle.md)
- [wiki/commands/inspection-and-search.md](commands/inspection-and-search.md)
- [wiki/commands/querying-and-automation.md](commands/querying-and-automation.md)
- [wiki/commands/validation-and-maintenance.md](commands/validation-and-maintenance.md)
- [wiki/web-ui-and-api.md](web-ui-and-api.md)
