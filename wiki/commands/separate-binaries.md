# Separate Binaries

`lif-serve` is a standalone binary built from the same repository as `lif`, wrapping one dedicated application command. The main binary does not advertise a runnable `serve` subcommand; a hidden migration stub only reports that the command moved to a separate executable and exits with an error. The terminal UI is not separate: it is the `lif tui` subcommand of the main binary, see [Terminal UI](../tui-companion.md).

## Why one extra binary

`lif` registers the core, file-based subcommands — creation, editing, listing, [validation](validation-and-maintenance.md), and [querying](querying-and-automation.md), among others — plus `tui`. `serve` is registered separately, in its own `cmd/` entry point, so that a plain `lif` install carries no web server dependencies, and a deployment that only needs the CLI doesn't have to ship the web server:

| Binary | Wraps | Purpose |
|---|---|---|
| `lif` | every core subcommand, including `tui` | day-to-day CLI use, scripting, agent tool calls, the [terminal UI](../tui-companion.md) via `lif tui` |
| `lif-serve` | `serve` | starts the web UI and GraphQL HTTP API — see [Web UI and API](../web-ui-and-api.md) |

`lif-serve` is a thin `main()` that builds the same root command, registers its one subcommand, and defaults to running that subcommand when invoked with no arguments or with a flag as the first argument — so `lif-serve` alone is equivalent to `lif-serve serve`.

`lif-serve` still accepts the two global flags every `lif` command does:

| Flag | Description |
|---|---|
| `--lif-path string` | Path to the data directory (overrides config and `LIF_PATH`) |
| `--config string` | Path to a config file (default: searches upward for `.lif.yml`) |

`lif-serve` does not register `version`, `graphql`, `check`, or any other core command — `lif-serve version` fails with cobra's "unknown command" error. Version and diagnostic commands are only available on the main `lif` binary.

## `lif-serve`

```
$ lif-serve --help
Start the web server

Usage:
  lif-serve serve [flags]

Aliases:
  serve, server, s

Flags:
      --cors-origin strings   Allowed CORS origins (use * to allow all) (default [http://localhost:*,http://127.0.0.1:*])
  -h, --help                  help for serve
  -p, --port int              Port to listen on (default 8080)
```

The global flags in the table above apply; the full `lif-serve --help` output also lists them.

Starting it launches the web UI and GraphQL API on the configured port (default `8080`). Full route list, network binding behavior, and security posture are documented in [Web UI and API](../web-ui-and-api.md); do not run it against a store you don't want reachable from every process on the machine without first reading that page's CORS and origin-checking section.

## Building the separate binaries

Use the repository's canonical source-build pipeline; it generates the embedded frontend before compiling both entrypoints and stamps version metadata:

```
mise install
mise run setup
mise run build
```

The resulting `lif` and `lif-serve` executables are written to the repository root. See [Installation](../installation.md) for prerequisites and PATH setup.

## Related documentation

- [Web UI and API](../web-ui-and-api.md)
- [Terminal UI](../tui-companion.md)
- [Querying and Automation](querying-and-automation.md)
- [Installation](../installation.md)
