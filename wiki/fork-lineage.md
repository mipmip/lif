# Fork Lineage

This page explains where this codebase comes from and how Lif relates to the original project. It exists so that anyone evaluating, contributing to, or reporting issues against Lif understands the attribution and continuity that apply to it.

## Origin

Lif originated as a fork of [hmans/beans](https://github.com/hmans/beans), an agentic-first issue tracker created by Hendrik Mans that stores tasks as Markdown files alongside a project's code.

The original idea, the core data model (one Markdown file per work item with YAML front matter, stored in a data directory), the initial command surface, and the initial GraphQL query engine all originate from that upstream project, and Lif is grateful for that foundation.

## Project home

Lif is maintained at [xRiErOS/lif](https://github.com/xRiErOS/lif). Report issues, request features and send pull requests there.

## Names

The fork was first published under the name beans, like the upstream project, and is now called Lif. The domain object, formerly called a bean, is called an *item*; the type names `milestone`, `epic`, `feature`, `bug` and `task` are unchanged. Existing projects keep working because the former data names are still read in place; see [Legacy names](configuration.md#legacy-names). Every breaking rename, with its migration step, is listed in the [upgrade guide](compatibility-and-upgrading.md).

## License and attribution

Both the upstream project and Lif are distributed under the Apache License, Version 2.0; the license text is unchanged and is included in this repository's `LICENSE` file.

Apache-2.0 permits forking, modification, and independent redistribution, provided the license and any existing copyright and attribution notices are preserved, which Lif does.

Any code, documentation, or design carried over from `hmans/beans` remains attributed to that project and to Hendrik Mans; new work added in Lif is layered on top of it under the same license.

## Why this fork exists

Lif is an independent, actively maintained continuation of the upstream codebase, developed outside the upstream repository. It is not a hostile fork; it exists so that development, bug fixes, and new features can continue on an independent schedule and roadmap, separate from upstream's own release cadence.

## What has changed so far

Development has focused on richer planning (bundled project profiles, scoped roadmaps, order and rename workflows), safer automation (completion policies, structured JSON and error output), worktree-aware storage, hardening existing behavior (data-integrity edge cases in item storage and rename operations, path-resolution precedence between the `--lif-path` flag, project configuration and the `LIF_PATH` environment variable), and the Lif rename with its item vocabulary.

The on-disk data format (Markdown files with YAML front matter) and the overall command-line surface remain consistent with the upstream design described above; see [Compatibility and Upgrading](compatibility-and-upgrading.md) for how to verify compatibility for a specific version.

Commit history follows the Conventional Commits convention (`feat:`, `fix:`, `perf:`, `refactor:`, with a `!` marker for breaking changes), which is also what upstream uses and what the release changelog is generated from.

## Related projects

The terminal UI is part of Lif itself: `lif tui` is a subcommand of the main binary. It is the former standalone companion [xRiErOS/beans-tui](https://github.com/xRiErOS/beans-tui) (executable `bt`), absorbed into this repository; it replaced the separate binary that carried the upstream terminal UI. The beans-tui repository keeps its own history. See [Terminal UI](tui-companion.md).

## How to reference Lif

When filing issues, requesting features, or citing behavior, reference [xRiErOS/lif](https://github.com/xRiErOS/lif), rather than the upstream repository, so that reports land with the maintainers who are actually responsible for this codebase.

When citing the underlying design or crediting the original author, reference [hmans/beans](https://github.com/hmans/beans) and Hendrik Mans as the source of the original concept and implementation.

## Related documentation

- [Compatibility and Upgrading](compatibility-and-upgrading.md)
- [Installation](installation.md)
- [Terminal UI](tui-companion.md)
