---
myst:
  html_meta:
    description: How to get a throwaway Multipass VM or LXD container that can actually run Concierge's dev preset, including the container settings Kubernetes needs.
---

(how-to-run-in-a-throwaway-machine)=
# How to run Concierge in a throwaway machine

{ref}`explanation-what-is-concierge` says to use Concierge in virtual machines, CI runners, and dedicated test hosts — not your daily workstation. This guide covers getting one of those machines, then handing it to Concierge.

Concierge doesn't create the machine itself; that's left to whichever tool you already use for throwaway machines. This guide covers two of them.

## Multipass

A Multipass VM needs no special configuration:

```bash
multipass launch --name concierge-dev --cpus 4 --memory 8G --disk 20G noble
multipass shell concierge-dev
```

Then jump to {ref}`prepare the machine <how-to-run-in-a-throwaway-machine-prepare>` below.

## LXD container

A plain `lxc launch` container isn't enough: the `dev` preset bootstraps a Kubernetes cluster, and running a kubelet inside an unprivileged container fails in confusing ways unless the container is configured for it first.

Launch the container with nesting, privilege, and the kernel modules `kube-proxy` needs:

```bash
lxc launch ubuntu:24.04 concierge-dev \
  -c security.nesting=true \
  -c security.privileged=true \
  -c linux.kernel_modules=ip_vs,ip_vs_rr,ip_vs_wrr,ip_vs_sh,ip_tables,ip6_tables,netlink_diag,nf_nat,overlay,br_netfilter
```

Pass `/dev/kmsg` through, which `kubelet` reads directly:

```bash
lxc config device add concierge-dev kmsg unix-char source=/dev/kmsg path=/dev/kmsg
```

Relax AppArmor confinement and the cgroup device rules, since the default container profile is too strict for a nested container runtime:

```bash
lxc config set concierge-dev raw.lxc="lxc.apparmor.profile=unconfined
lxc.cgroup.devices.allow=a"
```

If you're bind-mounting a project directory into the container, remap the container's `ubuntu` user to your own UID/GID first, so files created on either side are readable from both:

```bash
lxc config set concierge-dev raw.idmap="both $(id -u) $(id -g)"
lxc config device add concierge-dev project disk source="$(pwd)" path=/home/ubuntu/project
```

Restart the container for the profile changes to take effect, then shell in:

```bash
lxc restart concierge-dev
lxc exec concierge-dev -- su -l ubuntu
```

(how-to-run-in-a-throwaway-machine-prepare)=
## Prepare the machine

Inside the VM or container, install and run Concierge as usual:

```bash
sudo snap install --classic concierge
sudo concierge prepare -p dev
```

See {ref}`how-to-set-up-a-machine` for other presets, and {ref}`reference-presets` for what each one installs.
