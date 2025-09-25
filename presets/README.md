# Concierge Presets

This directory contains YAML versions of the built-in presets defined in `internal/config/presets.go`. These files can be used as custom configuration files or as templates for creating new configurations.

## Available Presets

### machine.yaml
A configuration preset designed to be used when testing machine charms. This preset:
- Enables LXD with bootstrap
- Installs common packages (python3-pip, python3-venv)
- Installs development snaps including charmcraft, jq, yq, and snapcraft
- Uses standard Juju model defaults for testing

### k8s.yaml
A configuration preset designed to be used when testing Kubernetes charms. This preset:
- Enables LXD (without bootstrap for building charms)
- Enables and bootstraps K8s with specific features and constraints
- Installs common packages and development snaps including rockcraft
- Configures K8s with load-balancer, local-storage, and network features

### microk8s.yaml
A configuration preset designed to be used when testing Kubernetes charms with MicroK8s. This preset:
- Enables LXD (without bootstrap for building charms)
- Enables and bootstraps MicroK8s with essential addons
- Installs common packages and development snaps including rockcraft
- Configures MicroK8s with hostpath-storage, dns, rbac, and metallb addons

### dev.yaml
A comprehensive development preset that combines both LXD and K8s capabilities. This preset:
- Enables and bootstraps both LXD and K8s
- Installs all development tools (charmcraft, rockcraft, snapcraft, jhack)
- Includes jhack with proper snap connections for Juju integration
- Ideal for developers working on both machine and Kubernetes charms

### crafts.yaml
A minimal preset focused on building artifacts without Juju. This preset:
- Disables Juju entirely
- Enables LXD for container-based building
- Installs all craft tools (charmcraft, rockcraft, snapcraft)
- Useful for CI/CD workflows focused only on artifact building

## Usage

These preset files can be used in several ways:

1. **As custom configuration files:**
   ```bash
   concierge prepare --config presets/dev.yaml
   ```

2. **As templates for new configurations:**
   Copy any preset file and modify it according to your needs.

3. **For reference:**
   Compare with the built-in presets to understand the configuration format.
