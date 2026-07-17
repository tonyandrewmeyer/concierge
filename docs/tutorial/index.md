---
myst:
  html_meta:
    description: A hands-on tutorial that installs Concierge, prepares a charm development machine with the dev preset, and restores it.
---

(tutorial)=
# Tutorial

This tutorial walks you through installing Concierge, preparing a machine with
a preset, checking its status, and restoring the machine to its original state.

## What you will need

- A fresh Ubuntu 24.04 or 26.04 machine (a VM, LXD container with nesting, or a
  dedicated cloud instance). Concierge makes broad changes to the system, so do
  not run it against a machine you rely on for other work.
- `sudo` privileges.
- An internet connection.

## Install Concierge

Install the snap:

```bash
sudo snap install --classic concierge
```

## Prepare a machine

Run the `dev` preset to install LXD, Juju, `charmcraft`, `snapcraft`,
`rockcraft`, and a few supporting snaps, and to bootstrap a Juju controller
onto LXD and Kubernetes:

```bash
sudo concierge prepare -p dev
```

Concierge prints each step as it runs. Depending on network speed, expect it
to take several minutes.

## Check the status

Once `prepare` finishes, ask Concierge what it did:

```bash
sudo concierge status
```

You should see the providers that were bootstrapped, the snaps and packages
that were installed, and the Juju controllers that are now available.

## Try Juju

Concierge configures Juju for you. List the controllers to confirm:

```bash
juju controllers
```

You can now use Juju as normal — start with the [Juju
documentation](https://documentation.ubuntu.com/juju/latest/).

## Restore the machine

When you're done, undo everything Concierge did:

```bash
sudo concierge restore
```

:::{important}
`concierge restore` is the literal reverse of `concierge prepare`. It removes
everything Concierge would have installed, even if the machine had some of
those things beforehand. Run it only on machines you're happy to reset.
:::

## Next steps

- Adapt one of the built-in [presets](../reference/presets) into a
  [custom config](../how-to/write-a-custom-config).
- Preview changes before they happen with [dry-run mode](../how-to/preview-changes).

```{toctree}
:hidden:

self
```
