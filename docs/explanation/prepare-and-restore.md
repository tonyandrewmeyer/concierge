---
myst:
  html_meta:
    description: Why concierge restore is a strict opposite of concierge prepare — and why that means Concierge is only for throwaway machines.
---

(explanation-prepare-and-restore)=
# `prepare` and `restore`

Concierge has two mirror-image commands: `prepare` provisions the machine
according to your configuration; `restore` undoes what `prepare` did.

## They are opposites, not deltas

`concierge restore` does not observe the machine and revert changes it can
detect. It computes what `prepare` **would install** from the same
configuration, and removes exactly that set. This makes restore predictable
and cheap, but it has a consequence:

> If the machine already had one of Concierge's snaps, packages, or
> configuration files before you ran `prepare`, `restore` will remove it
> anyway.

For that reason, Concierge is intended for **throwaway** machines — VMs, CI
runners, dedicated test hosts — not for the workstation you use for
everything else.

## Why declarative

A `concierge.yaml` file captures exactly what a machine needs to be a
charm-development environment, in a form that can be committed to version
control, diffed, and shared between developers and CI. Two developers using
the same config get the same setup; a CI job using that config gets the same
setup as a developer's laptop.

## Why an all-or-nothing restore

The alternative — a restore that tries to preserve pre-existing state — would
need to snapshot every changed file, every installed package, every Juju
controller before touching them, and remember which of them existed
beforehand. That's a substantial amount of machinery to build and maintain,
and it hides bugs when the snapshot and reality drift apart.

The all-or-nothing restore is straightforward: whatever the config says
`prepare` installs, `restore` removes. The trade-off is the constraint above:
don't run Concierge on a machine you care about.
