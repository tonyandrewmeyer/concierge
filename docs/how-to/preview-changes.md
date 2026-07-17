---
myst:
  html_meta:
    description: How to use the --dry-run flag on concierge prepare and restore to see exactly which commands Concierge would execute without touching the system.
---

(how-to-preview-changes)=
# Preview changes with `--dry-run`

Both `concierge prepare` and `concierge restore` accept `--dry-run`. In this
mode, Concierge prints the shell commands it would run without touching the
system.

```bash
sudo concierge prepare -p dev --dry-run
sudo concierge restore --dry-run
```

Use dry-run to:

- Confirm what a preset or custom config will actually do before committing to
  it.
- Copy the printed commands into a script if you want to run them manually or
  on a machine that doesn't have Concierge installed.
- Diff two configs by running both and comparing the output.

## What dry-run does and doesn't do

Dry-run **reads** system state (for example, to decide whether a snap is
already installed) but **never modifies** anything: no packages are installed
or removed, no files are created or modified, and no Juju controllers are
bootstrapped or destroyed.

Because dry-run reads state, it will still fail if your configuration
references something that doesn't exist — for example, a missing
`credentials-file`.

The log level defaults to `error` in dry-run so the printed commands stand
out. Pass `--verbose` or `--trace` if you also want to see decision-making
detail.
