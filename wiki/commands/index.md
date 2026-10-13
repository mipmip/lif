# Command reference

This index maps every top-level `lif` command to its canonical reference page. Run `lif <command> --help` for the complete option list of the installed version.

## Project setup

- [`init`](project-setup.md#lif-init) — initialize a repository and optionally expand a project profile.
- [`path`](project-setup.md#lif-path) — print the resolved items data directory.
- [`prime`](project-setup.md#lif-prime) — emit project-specific instructions for coding agents.
- [`pick`](project-setup.md#lif-pick) — choose an item interactively and print its ID.
- [`version`](project-setup.md#lif-version) — print build and version information.
- [`help`](project-setup.md#lif-help) — open command help.
- [`completion`](project-setup.md#lif-completion) — generate shell completion scripts.

## Inspection and search

- [`list`](inspection-and-search.md#lif-list-alias-ls) — filter, sort, render, or export items.
- [`show`](inspection-and-search.md#lif-show) — display one item and its content.
- [`next`](inspection-and-search.md#lif-next) — select the highest-priority ready item.

## Lifecycle

- [`create`](lifecycle.md#lif-create-aliases-c-new) — create one item with type, status, priority, body, tags, and relationships.
- [`start`](lifecycle.md#lif-start) — mark one or more items in progress.
- [`complete`](lifecycle.md#lif-complete) — complete items under the configured policies.
- [`scrap`](lifecycle.md#lif-scrap) — mark work as intentionally abandoned.
- [`archive`](lifecycle.md#lif-archive) — move completed or scrapped items into project memory.
- [`delete`](lifecycle.md#lif-delete-alias-rm) — permanently remove items after confirmation or explicit force.

## Organization and relationships

- [`update`](organization-and-relations.md#lif-update) — change fields, content, and relationships.
- [`tag`](organization-and-relations.md#lif-tag) — add or remove tags in batches.
- [`order`](organization-and-relations.md#lif-order) — position an item among its siblings.
- [`rename`](organization-and-relations.md#lif-rename) — rename a slug, ID, or project prefix while updating references.

## Planning and reporting

- [`roadmap`](planning-and-reporting.md#lif-roadmap) — render a scoped project roadmap for terminals or Markdown.
- [`milestones`](planning-and-reporting.md#lif-milestones) — summarize top-level planning containers and descendant progress.
- [`progress`](planning-and-reporting.md#lif-progress) — summarize project or subtree status.
- [`graph`](planning-and-reporting.md#lif-graph) — render parent and blocking relationships as DOT, ASCII, or JSON.

## Querying and automation

- [`graphql`](querying-and-automation.md#lif-graphql) — execute precise GraphQL queries and mutations.
- [JSON output and error contracts](querying-and-automation.md#json-output-contract) — integrate items with scripts and coding agents.

## Validation and maintenance

- [`check`](validation-and-maintenance.md#lif-check) — validate configuration, files, links, policies, and optional commit references.
- [`promote`](validation-and-maintenance.md#lif-promote) — turn selected review findings into items.

## Separate applications

- [`lif-serve`](separate-binaries.md#lif-serve) — run the browser workspace and GraphQL service.
- [`lif-tui`](separate-binaries.md#lif-tui) — run the original terminal UI as a separate binary.
- [`bt`](../tui-companion.md#layer-3-bt-the-separate-po-cockpit) — use the separately maintained Product Owner cockpit.

## Related documentation

- [Feature overview](../feature-overview.md)
- [Quick start](../quick-start.md)
- [Configuration reference](../configuration.md)
- [Agent integration](../agent-integration.md)
- [Troubleshooting](../troubleshooting.md)
