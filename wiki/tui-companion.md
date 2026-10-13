# TUI Companion

This page explains the layers of terminal tooling around Lif data: the `lif` CLI itself, the bundled `lif-tui` binary, and the separate `bt` companion project. Read this before assuming any terminal UI is Lif's primary interface.

## Layer 1: `lif` CLI is the source of truth

Every terminal UI described on this page is a client of the `lif` CLI and its underlying `.lif` data directory; none of them hold state of their own. If a terminal UI and the `lif` CLI ever disagree about an item's fields, the item files on disk (as read by `lif show`/`lif list`) are authoritative, not the UI's cached rendering. Command and data details live in [`commands/separate-binaries.md`](commands/separate-binaries.md), [`feature-overview.md`](feature-overview.md), and [`configuration.md`](configuration.md).

## Layer 2: `lif-tui`, the bundled terminal UI

`lif-tui` is a separate binary built from `cmd/lif-tui` in this repository; running it with no arguments opens the interactive terminal interface for browsing and managing items, the same `tui` view that used to ship inside the main binary. The `tui` subcommand is not part of the main `lif` binary: running `lif tui` prints `The "tui" command has moved to a separate binary: lif-tui` and `Please install and use "lif-tui" instead.`, then exits with a non-zero status, so `lif-tui` is required to use that view at all. It is the terminal UI carried over from the upstream project; it is kept for compatibility and receives no new features here. Build and install instructions for `lif-tui` are in [`commands/separate-binaries.md`](commands/separate-binaries.md); general setup is in [`installation.md`](installation.md).

## Layer 3: `bt`, the separate PO cockpit

A separate project, maintained independently of this repository, provides the `bt` executable: a keyboard-first, mouse-friendly Product Owner cockpit that works on item data. It ships and versions independently, with its own README, changelog and releases, and it is licensed Apache-2.0 like Lif. Its repository and name are listed on the [fork lineage](fork-lineage.md) page.

Which data names a given `bt` release reads, and which Lif release it targets, is stated in its own documentation; this repository does not verify it. Before pointing `bt` at a project that uses the current names (`.lif/`, `.lif.yml`), read its release notes.

## Choosing between the two

Use `lif-tui` when you only need the bundled view and have no use for the companion's capabilities. For day-to-day Product Owner work, evaluate `bt` from its own project: it is where active terminal UI development happens, and `lif-tui` will not receive equivalent features. Neither is endorsed by or affiliated with the original upstream project that Lif descends from; see [`fork-lineage.md`](fork-lineage.md) for how Lif positions itself relative to that upstream.

## Related documentation

- [fork-lineage.md](fork-lineage.md)
- [installation.md](installation.md)
- [feature-overview.md](feature-overview.md)
- [project-profiles.md](project-profiles.md)
- [commands/separate-binaries.md](commands/separate-binaries.md)
- [configuration.md](configuration.md)
- [web-ui-and-api.md](web-ui-and-api.md)
