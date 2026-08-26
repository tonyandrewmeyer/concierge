# AGENTS.md

This file provides guidance to AI agents when working with code in this repository.

## Build and Test Commands

```bash
# Build the binary
go build

# Build a snapshot release with goreleaser (output in ./dist)
goreleaser build --clean --snapshot

# Run directly
go run .

# Run unit tests
go test ./...

# Run all integration tests (requires LXD)
spread -v lxd:

# Run specific integration test
spread -v lxd:ubuntu-24.04:tests/juju-model-defaults

# Run integration tests on local machine
spread -v github-ci:
```

Note: The binary must be run with `sudo` for most operations since it installs system packages and configures providers.

## Gotchas

- **Never call `exec.Command()` directly.** Build commands with `system.NewCommand(executable, []string{arg1, arg2})`, passing each argument as a separate slice element (no string concatenation) — the binary runs as root, so this avoids command injection.
- During `prepare`, the merged configuration (including all overrides) is saved to `~/.cache/concierge/concierge.yaml`; `restore` reads this file to undo exactly what was provisioned.
- **Refreshing a snap to a different channel may require stopping it first.** See `internal/providers/lxd.go` (`workaroundRefresh()`) for the pattern.
