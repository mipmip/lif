# Quick Start

This page initializes a disposable project, builds a milestone/epic/feature hierarchy, and drives it through the core lifecycle. After initialization, every command uses an explicit `--lif-path`, so inherited `LIF_PATH` values cannot redirect the example.

## Prerequisites

A working `lif` binary on your `PATH`; see [Installation](installation.md) if you have not built one yet.

## Step 1: create a disposable project directory

```bash
mkdir /tmp/lif-quickstart-demo
cd /tmp/lif-quickstart-demo
```

## Step 2: initialize with an explicit profile

`lif init --profile` cannot be combined with `--lif-path`, because a profile writes `.lif.yml` while an explicit store path intentionally skips that file. Run the profiled initialization inside the disposable directory itself.

```bash
lif init --profile classic
```

The `classic` profile defines five types: `milestone`, `epic`, `feature`, `bug`, and `task`, which is what the hierarchy below uses. `lif init --help` lists the other available profiles (`complex`, `simple`, `todo`).

## Step 3: build a small hierarchy

Every following command passes `--lif-path .lif` explicitly, pointing at the directory `init` just created, rather than relying on the config file it wrote or any environment variable.

Create a milestone, then an epic underneath it, then a feature and a bug underneath the epic:

```bash
lif --lif-path .lif create "Launch v1" --type milestone
lif --lif-path .lif create "Onboarding flow" --type epic --parent <milestone-id>
lif --lif-path .lif create "Add welcome email" --type feature --parent <epic-id> --priority high
lif --lif-path .lif create "Fix typo in welcome banner" --type bug --parent <epic-id>
```

Each `create` call prints `Created <id> <filename>`; substitute the generated IDs for `<milestone-id>` and `<epic-id>` in subsequent commands.

## Step 4: list what exists

```bash
lif --lif-path .lif list
```


## Step 5: find the next item to work on

```bash
lif --lif-path .lif next
```

`next` picks the highest-priority item that is not blocked and not already in progress, completed, scrapped, or a draft; in this example that is the high-priority feature created in step 3.

## Step 6: start it

```bash
lif --lif-path .lif start lif-quickstart-demo-l5py
```

This marks the item's `status` as `in-progress` and prints its full contents.

## Step 7: complete it

```bash
lif --lif-path .lif complete lif-quickstart-demo-l5py --summary "Welcome email shipped"
```

`--summary` records what changed; `complete` also accepts `--commit` to record a git ref alongside the completion.

## Step 8: view the roadmap

```bash
lif --lif-path .lif roadmap --format tty --include-done
```


`--include-done` is required because the feature you completed would otherwise be hidden. The terminal view renders the milestone, epic, and both children as a styled tree with a legend, confirming the hierarchy end to end.

## Cleanup

The example store lives entirely under `/tmp/lif-quickstart-demo`; remove it when you are done with `rm -rf /tmp/lif-quickstart-demo`.

## Related documentation

- [Installation](installation.md)
- [Project Setup](commands/project-setup.md)
- [Lifecycle](commands/lifecycle.md)
- [Inspection and Search](commands/inspection-and-search.md)
- [Planning and Reporting](commands/planning-and-reporting.md)
- [Organization and Relations](commands/organization-and-relations.md)
- [Project Profiles](project-profiles.md)
- [Feature Overview](feature-overview.md)
- [Configuration](configuration.md)
