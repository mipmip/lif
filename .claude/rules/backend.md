---
paths:
  - "**/*.go"
  - "internal/**"
  - "pkg/**"
  - "cmd/**"
---

# Go Backend

## Package Layering

The codebase follows a strict dependency direction:

```
cmd/ → internal/commands/ → internal/graph/ → pkg/itemgraph/ → pkg/itemcore/ → pkg/item/
                          → internal/agent/
                          → internal/worktree/
                          → internal/terminal/
```

- `pkg/item/` — pure data model, no I/O, no dependencies beyond stdlib
- `pkg/itemcore/` — stateful engine with mutex-protected in-memory store + disk persistence + file watchers
- `internal/graph/` — GraphQL resolvers; the API layer that glues everything together
- `internal/agent/` — manages Claude Code subprocess lifecycle
- `internal/worktree/` — manages git worktree lifecycle
- `internal/terminal/` — manages PTY sessions

Do not introduce upward dependencies (e.g., `pkg/` importing `internal/`).

## Resolver Pattern

GraphQL resolvers in `internal/graph/schema.resolvers.go` are the API boundary. They should:

- Validate input (etags, IDs, relationships) before mutating state
- Normalize short IDs to full IDs via `Core.NormalizeID()` before use
- Use helper methods on `itemgraph.CoreResolver` (e.g., `validateETag`, `SetParent`) rather than inlining validation logic
- Return user-facing error messages (these surface in the GraphQL response)
- After mutations, notify subscribers so subscriptions push updates to the frontend

## Agent Manager

The agent manager (`internal/agent/manager.go`) owns the lifecycle of Claude Code subprocesses:

- Sessions are keyed by item ID (or `__central__` for the main workspace)
- One `AgentChatStore` per session on the frontend, one `Session` + optional `runningProcess` on the backend
- JSONL conversation files are persisted to `.lif/.conversations/<itemID>.jsonl`
- The store format uses `entry` structs with `type: "message"` or `type: "meta"` — see `store.go`
- Images are saved to `.lif/.conversations/images/` and referenced by UUID in JSONL entries
- The manager provides pub/sub channels for real-time subscription updates — always notify subscribers after state changes, even for "empty" states (see GraphQL Subscriptions rule in CLAUDE.md)
- Agent processes are spawned with `--print` and `--output-format stream-json` for structured output parsing

## Worktree Manager

The worktree manager (`internal/worktree/worktree.go`) handles git worktree lifecycle:

- Worktrees are created under a `lif/` branch prefix
- Optional post-creation setup commands run asynchronously with status tracking (`SetupRunning` → `SetupDone`/`SetupFailed`)
- `DetectItemIDs` scans for item ID changes by diffing the worktree's `.lif/` dir against the base branch

## Concurrency

- `itemcore.Core` and `agent.Manager` are accessed concurrently from GraphQL resolvers. Both use `sync.RWMutex` — read operations take `RLock`, writes take `Lock`.
- Subscription channels are buffered (size 1) with non-blocking sends to avoid blocking the mutation path.
- File watchers run in background goroutines — be careful about accessing shared state from watcher callbacks.

## Error Handling

- Wrap errors with context using `fmt.Errorf("doing X: %w", err)` — don't return bare errors from internal functions.
- CLI commands should use `cmd.PrintErrln()` for user-facing errors, not `log.Fatal` (which kills the process).
- GraphQL resolver errors become user-visible error messages — keep them clear and actionable.
