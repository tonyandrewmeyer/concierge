---
myst:
  html_meta:
    description: How to adapt a built-in Concierge preset into a custom concierge.yaml — adding snaps, changing the Juju version, disabling providers, or adding new ones.
---

(how-to-write-a-custom-config)=
# Write a custom config

When the built-in [presets](../reference/presets) don't fit your needs, provide
your own configuration. Concierge reads a YAML file named `concierge.yaml`
from the current directory when you pass `-c concierge.yaml`, or when you run
`sudo concierge prepare` with no preset.

The best starting point is the preset that most closely matches what you want.

## Start from a preset

Copy the preset you want to adapt. If you installed Concierge from the snap,
the preset files are not on disk locally; fetch them from the repository:

```bash
curl -o concierge.yaml \
  https://raw.githubusercontent.com/canonical/concierge/main/presets/dev.yaml
```

Or, if you built Concierge from source:

```bash
cp $(go env GOPATH)/src/github.com/canonical/concierge/presets/dev.yaml concierge.yaml
```

Edit the file to suit your needs, then run:

```bash
sudo concierge prepare -c concierge.yaml
```

## Common adaptations

### Add or remove a snap

Snaps live under `host.snaps` as a map keyed by snap name. Add an entry to
install a snap; remove one to skip it. Add a `channel:` if you need a specific
track. Adapted from `dev.yaml`:

```yaml
host:
  snaps:
    charmcraft:
    jq:
    astral-uv:
      channel: latest/beta
```

### Change the Juju version

Set `juju.channel`, and optionally pin an `agent-version`:

```yaml
juju:
  channel: 3.6/stable
  agent-version: "3.6.11"
```

### Turn off Kubernetes

Start from `dev.yaml` and either delete the `k8s:` block from `providers:`
or disable it explicitly:

```yaml
providers:
  k8s:
    enable: false
```

### Add a provider that presets don't cover

To add a Google cloud provider on top of an existing preset, extend the
`providers:` block and provide credentials — see
[Provide cloud credentials](provide-credentials):

```yaml
providers:
  google:
    enable: true
    bootstrap: true
    credentials-file: /home/ubuntu/google-credentials.yaml
```

## Full reference

Every field is documented in the [configuration schema](../reference/configuration).

## Try it without side effects

Before running `prepare` for real, use [dry-run mode](preview-changes) to see
exactly what Concierge would do.
