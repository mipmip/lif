# Lif documentation

This directory is the canonical documentation for Lif, maintained at xRiErOS/lif. Start with the quick path below, then open only the reference page needed for the current task.

## Start here

- [Installation](installation.md) — build and install Lif and its three binaries.
- [Quick start](quick-start.md) — initialize a project, create work, and render the first roadmap.
- [Feature overview](feature-overview.md) — understand the product surface without reading command help.
- [Project profiles](project-profiles.md) — choose between `classic`, `todo`, `simple`, and `complex`.

## Use Lif with coding agents

- [Agent integration](agent-integration.md) — prime agents, automate session setup, and choose safe machine-readable interfaces.
- [Querying and automation](commands/querying-and-automation.md) — use GraphQL, JSON output, and structured errors.
- [Data model and file format](data-model.md) — understand the Markdown files shared by humans, scripts, and agents.

## Configure a project

- [Configuration reference](configuration.md) — configure stores, profiles, types, statuses, priorities, display, and completion policies.
- [Compatibility and upgrading](compatibility-and-upgrading.md) — upgrade guide: breaking changes with old form, new form, and migration step.
- [Fork lineage](fork-lineage.md) — understand where the project comes from and how it is attributed.

## Browse every command

- [Complete command index](commands/index.md) — map every top-level command to its canonical page.
- [Project setup](commands/project-setup.md) — `init`, `path`, `prime`, `pick`, `version`, `help`, and `completion`.
- [Inspection and search](commands/inspection-and-search.md) — `list`, `show`, and `next`.
- [Lifecycle](commands/lifecycle.md) — `create`, `start`, `complete`, `scrap`, `archive`, and `delete`.
- [Organization and relationships](commands/organization-and-relations.md) — `update`, `tag`, `order`, and `rename`.
- [Planning and reporting](commands/planning-and-reporting.md) — `roadmap`, `milestones`, `progress`, and `graph`.
- [Querying and automation](commands/querying-and-automation.md) — `graphql`, JSON output, and structured errors.
- [Validation and maintenance](commands/validation-and-maintenance.md) — `check`, `promote`, policies, and integrity checks.
- [Separate applications](commands/separate-binaries.md) — `lif-serve` and the `lif-tui` binary.

## Use visual interfaces

- [Web UI and API](web-ui-and-api.md) — run the planning workspace, board, and GraphQL service.
- [TUI companion](tui-companion.md) — distinguish the bundled `lif-tui` from the separate `bt` companion.

## Solve problems

- [Troubleshooting](troubleshooting.md) — diagnose store discovery, ready work, policies, configuration, and server startup.

## Documentation contract

`lif --help`, each `lif <command> --help`, and `lif prime` are the primary behavioral sources for this documentation. Release-specific pages and screenshots must be regenerated from the exact public build they describe.

Each detailed subject has one canonical page. The repository README presents the product and links here instead of duplicating the reference material.
