# Installation

This page covers the four ways to install Lif and explains the difference between its three binaries. If you already run an install from an earlier release, follow the [upgrade guide](compatibility-and-upgrading.md) first.

## Origin

Lif is maintained at [xRiErOS/lif](https://github.com/xRiErOS/lif), and its Go module path is `github.com/xRiErOS/lif`. See [Fork lineage](fork-lineage.md) for the project's origin and attribution.

## Installation channels

All four channels install the same three binaries described below. Pick one:

### Homebrew (macOS)

```sh
brew install xRiErOS/lif/lif
```

Upgrade with `brew upgrade --cask lif`.

### Install script (Linux, macOS)

```sh
curl -fsSL https://raw.githubusercontent.com/xRiErOS/lif/main/install.sh | sh
```

Configurable via environment variables: `LIF_VERSION` (default: latest release), `LIF_BIN_DIR` (default: `$HOME/.local/bin`), `LIF_REPO` (default: `xRiErOS/lif`).

### go install (Linux, macOS, Windows)

```sh
go install github.com/xRiErOS/lif/cmd/lif@latest
go install github.com/xRiErOS/lif/cmd/lif-tui@latest
```

A `lif-serve` built this way embeds no web assets — the repository tracks only a placeholder (`internal/web/dist/.gitkeep`) so the `//go:embed` directive in `internal/web/embed.go` compiles, but the actual frontend build is not part of the module source. Requesting `/` returns 404. Use Homebrew, the install script, or a release archive for a working `lif-serve`.

A binary installed this way still reports the correct release version (`lif version` shows the release tag, not `dev`): `internal/version/buildinfo.go` reads it from `runtime/debug.ReadBuildInfo()` when the build did not receive `-ldflags` (which `go install` never does). Commit and build date stay `unknown` for a `go install pkg@version` install, since module proxy downloads carry no `.git` history to stamp from; they are only populated when building from a local git checkout (`go build`/`go install` run inside the cloned repository).

### GitHub Releases

Download the `linux`/`darwin`/`windows` × `amd64`/`arm64` (`i386` on Linux only) archive for your platform from the [Releases page](https://github.com/xRiErOS/lif/releases), verify against the accompanying `checksums.txt`, and extract `lif`, `lif-serve`, `lif-tui` onto your `PATH`.

## Prerequisites (build from source only)

Install [mise](https://mise.jdx.dev/getting-started.html) first. The repository's `mise.toml` declares the Go, Node.js, and pnpm toolchains the build needs and installs them; the current Go module requires Go 1.26.

## The three binaries

The fork builds three separate binaries, one per entrypoint under `cmd/`: `lif` (`cmd/lif`, the CLI you use for day-to-day issue tracking), `lif-serve` (`cmd/lif-serve`, the standalone web UI and GraphQL API server), and `lif-tui` (`cmd/lif-tui`, the interactive terminal UI carried over from upstream).

Each binary is independent; you only need `lif-serve` if you want the web UI, and `lif-tui` if you want the retained terminal UI, alongside the required `lif` CLI.

## Building from source

Clone the fork, let mise install the declared toolchain, install dependencies and generated code, then run the repository's canonical build:

```sh
git clone https://github.com/xRiErOS/lif.git
cd lif

mise install
mise run setup
mise run build
```

The build pipeline compiles the Svelte frontend into `internal/web/dist` before building the Go entrypoints, and stamps version, commit, and build date into the binaries. A plain `go build ./cmd/lif` also compiles, because the repository tracks a placeholder in `internal/web/dist` that satisfies the embed directive, but it produces an unstamped binary, and a `lif-serve` built that way embeds no web assets and serves 404 for the browser UI.

`mise run build` writes the stamped `lif`, `lif-serve`, and `lif-tui` binaries to the repository root. Install them in a user-owned directory on `PATH`:

```sh
install -d "$HOME/.local/bin"
install -m 755 lif lif-serve lif-tui "$HOME/.local/bin/"
export PATH="$HOME/.local/bin:$PATH"
```

Persist the `PATH` addition in your shell configuration if `$HOME/.local/bin` is not already present.

## Verifying the build

Confirm that the installed CLI is reachable and carries the version, commit, and build date injected by the build task:

```sh
lif version
lif --help
```

A plain local `go build` does not inject those linker values (it also has no embedded frontend, see the `go install` section above); use `mise run build` for a fully stamped binary with the web UI, or one of the release-artifact channels above.

## License

The fork is distributed under the Apache License 2.0, the same license as the upstream project; see `LICENSE` in the repository root.

## Related documentation

- [Quick Start](quick-start.md)
- [Configuration](configuration.md)
- [Separate Binaries](commands/separate-binaries.md)
- [Web UI and API](web-ui-and-api.md)
- [TUI Companion](tui-companion.md)
- [Fork Lineage](fork-lineage.md)
- [Compatibility and Upgrading](compatibility-and-upgrading.md)
- [Troubleshooting](troubleshooting.md)
