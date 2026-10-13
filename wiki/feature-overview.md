# Feature Overview

This page maps what `lif` and its companion binaries actually do, grouped by capability rather than by command name. It is the starting point for understanding the whole toolset before diving into the per-command reference pages.

## Issue tracking on plain Markdown files

Every item is a Markdown file with YAML front matter, stored under a `.lif` directory inside the project it tracks. `lif create` writes a new item, `lif show` renders one, `lif update` changes its properties, and `lif delete` removes it outright; `lif complete`, `lif start`, and `lif scrap` move an item through its status lifecycle without needing to touch the file by hand. Because items are files next to the code they describe, they read, diff, and merge like any other text in version control, and an editor or a coding agent can open one directly with no separate database to query.

## A configurable hierarchy of types

Items are organized by type, and each type occupies a hierarchy rank: rank 1 is the topmost container, higher-numbered ranks nest under it, and the leaf rank (the highest number) holds the actual units of work. The built-in default hierarchy is `milestone` → `epic` → `feature`, with `bug` and `task` as leaves, but this table is fully configurable per project through `.lif.yml`, either by overriding individual built-in entries or by writing an exclusive type table that replaces the defaults entirely. `lif init --profile` writes one of four ready-made hierarchy shapes at project creation time; see [Project Profiles](project-profiles.md) for what each profile contains and when to pick it.

## Roadmap and progress reporting

`lif roadmap` generates a Markdown roadmap document from the current milestones, epics, and their child items, either for the whole project or scoped to one item's subtree. `lif milestones` lists every top-container item with a rollup of how many of its descendants are complete, and `lif progress` reports counts by status across the whole workspace or a single parent's subtree, including a percent-complete figure. A type can opt out of both `lif roadmap` and `lif milestones` by setting `roadmap: false`, which is how the `bucket` type in the `simple` and `complex` profiles stays a genuine parking lot instead of cluttering release-facing views.

## Querying and filtering

`lif list` and `lif next` share the same filtering vocabulary — `--type`, `--tag`, `--parent`, and `--sort` — so a query can move between "show me everything matching this" and "show me the single highest-priority ready item" without rewriting it. For anything the built-in flags cannot express, `lif graphql` runs an arbitrary GraphQL query or mutation directly against the Lif data, which is the same mechanism a coding agent uses to pull exactly the fields it needs while keeping token use low.

## Structure and relationships

Items can be nested (a feature under an epic, a task under a feature), reordered relative to their siblings with `lif order`, and connected through blocking and parent/child links. `lif graph` renders those relationships as Graphviz DOT, a terminal edge list, or JSON. `lif tag` adds or removes tags for cross-cutting grouping that does not depend on the type hierarchy. `lif rename` covers three distinct operations behind one command: renaming an item's slug, renaming a single item's ID with all references to it updated, and renaming the project-wide ID prefix across every item at once.

## Validation and integrity

`lif check` validates both configuration and data integrity in one pass: that colors and the default type resolve correctly, that no item carries a type the configuration no longer defines, that links do not point at items that do not exist, that no item links to itself, and that no cycle exists in the blocking or parent relationships. `--fix` automatically removes broken links and self-references; cycles always require manual intervention because there is no safe automatic resolution for one. `--strict` turns policy warnings into failures for use in a pre-commit hook or CI check.

## Completion policy: required fields

A project can require specific front-matter fields to be filled in before an item may enter a given status, configured under `lif.require_fields_on` in `.lif.yml`. The most common use is gating the `completed` status on a `commit` field, so `lif complete` refuses to finish an item until the commit that resolved it has been recorded; `lif prime` reflects whatever policy a project has configured so an agent knows the rule before it hits the refusal.

## Priming coding agents

`lif prime` prints a ready-to-paste prompt that tells an AI coding agent how to use the `lif` CLI on this specific project: its configured types and their hierarchy ranks, its statuses and priorities, and any required-fields completion policy currently in force. Because the prompt is generated from the live configuration rather than a static document, it stays accurate after a project changes its type table or its policy without anyone having to hand-edit onboarding instructions. See [Agent Integration](agent-integration.md) for the full agent workflow this supports.

## Git-aware worktrees

Lif configuration includes a `worktree` section covering the base ref new work branches from, a setup command to run after a worktree is created, a run command exposed as a workspace action, and an integration strategy for merging finished work back — either squash-merging locally or pushing a branch and opening a pull request. This lets a coding agent spin up isolated git worktrees for parallel work items without every agent needing its own bespoke branching convention.

## Interfaces beyond the CLI

The core `lif` binary is deliberately narrow; the terminal UI and the web UI ship as separate binaries, `lif-tui` and `lif-serve`, so installing the tracker does not require pulling in either interface. See [TUI Companion](tui-companion.md) for the terminal UI and [Web UI and API](web-ui-and-api.md) for the web UI and its HTTP surface.

## What is deliberately out of scope

`lif` stores issues, not time tracking, not CI pipeline state, and not a general project wiki; its roadmap and progress views only ever reflect the type hierarchy and status table a project has configured, so a type marked `roadmap: false` — such as a profile's `bucket` type — is excluded from those views by design rather than by omission. The tool assumes a single-repository, file-based store per project; cross-project aggregation is not a built-in feature of the core CLI.

## Related documentation

- [Project Profiles](project-profiles.md)
- [Configuration Reference](configuration.md)
- [Command Reference: Project Setup](commands/project-setup.md)
- [Command Reference: Inspection and Search](commands/inspection-and-search.md)
- [Command Reference: Lifecycle](commands/lifecycle.md)
- [Command Reference: Organization and Relations](commands/organization-and-relations.md)
- [Command Reference: Planning and Reporting](commands/planning-and-reporting.md)
- [Command Reference: Querying and Automation](commands/querying-and-automation.md)
- [Command Reference: Validation and Maintenance](commands/validation-and-maintenance.md)
- [Command Reference: Separate Binaries](commands/separate-binaries.md)
- [Data Model](data-model.md)
- [Agent Integration](agent-integration.md)
- [TUI Companion](tui-companion.md)
- [Web UI and API](web-ui-and-api.md)
