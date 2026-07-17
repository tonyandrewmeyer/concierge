---
myst:
  html_meta:
    description: What Concierge is designed to do — make charm-development machine setup declarative and reproducible — and what it deliberately isn't.
---

(explanation-what-is-concierge)=
# What Concierge is for

Charm development needs a familiar-looking machine: the right craft tools
(`charmcraft`, `snapcraft`, `rockcraft`), the right providers (LXD,
Kubernetes), and a Juju controller bootstrapped and ready to use. Getting a
machine into that shape by hand takes many commands and a lot of tacit
knowledge, and every developer's machine ends up subtly different.

Concierge exists to make that setup **declarative** and **reproducible**.
You describe the machine you want in a YAML file — or pick one of the
[built-in presets](../reference/presets) — and Concierge does the work.

## What Concierge is not

Concierge is a **provisioner**, not a runtime. Once it has prepared a machine,
you interact with the tools it installed (Juju, Charmcraft, snapd, …)
directly. Concierge doesn't sit in the loop.

Concierge is also not a **charm-development tutorial**. It gets you a machine
that is ready to develop charms; learning how to develop charms happens
elsewhere (see the [Ops
documentation](https://canonical.com/juju/docs/ops/latest/) for that).

## Where Concierge fits

The typical use is on a throwaway machine — a VM, a CI runner, a cloud
instance — that exists to build and test charms. Because
[`concierge restore`](prepare-and-restore) reverses everything `prepare` did,
Concierge is also a natural fit for CI pipelines that need to leave the runner
in a clean state between jobs.
