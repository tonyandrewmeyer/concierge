package providers

import (
	"fmt"
	"log/slog"
	"path"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/canonical/concierge/internal/config"
	"github.com/canonical/concierge/internal/packages"
	"github.com/canonical/concierge/internal/system"
)

// Default channel from which K8s is installed.
const defaultK8sChannel = "1.34-classic/stable"

// NewK8s constructs a new K8s provider instance.
func NewK8s(r system.Worker, config *config.Config) *K8s {
	var channel string

	if config.Overrides.K8sChannel != "" {
		channel = config.Overrides.K8sChannel
	} else if config.Providers.K8s.Channel != "" {
		channel = config.Providers.K8s.Channel
	} else {
		channel = defaultK8sChannel
	}

	return &K8s{
		Channel:              channel,
		Features:             config.Providers.K8s.Features,
		bootstrap:            config.Providers.K8s.Bootstrap,
		modelDefaults:        config.Providers.K8s.ModelDefaults,
		bootstrapConstraints: config.Providers.K8s.BootstrapConstraints,
		system:               r,
		debs: []*packages.Deb{
			{Name: "iptables"},
		},
		snaps: []*system.Snap{
			{Name: "k8s", Channel: channel},
			{Name: "kubectl", Channel: "stable"},
		},
	}
}

// K8s represents a K8s install on a given machine.
type K8s struct {
	Channel  string
	Features map[string]map[string]string

	bootstrap            bool
	modelDefaults        map[string]string
	bootstrapConstraints map[string]string

	system system.Worker
	debs   []*packages.Deb
	snaps  []*system.Snap
}

// Prepare installs and configures K8s such that it can work in testing environments.
// This includes installing the snap, enabling the user who ran concierge to interact
// with K8s without sudo, and sets up the user's kubeconfig file.
func (k *K8s) Prepare() error {
	err := k.install()
	if err != nil {
		return fmt.Errorf("failed to install K8s: %w", err)
	}

	err = k.init()
	if err != nil {
		return fmt.Errorf("failed to install K8s: %w", err)
	}

	err = k.configureFeatures()
	if err != nil {
		return fmt.Errorf("failed to enable K8s features: %w", err)
	}

	err = k.setupKubectl()
	if err != nil {
		return fmt.Errorf("failed to setup kubectl for K8s: %w", err)
	}

	slog.Info("Prepared provider", "provider", k.Name())

	return nil
}

// Name reports the name of the provider for Concierge's purposes.
func (k *K8s) Name() string { return "k8s" }

// Bootstrap reports whether a Juju controller should be bootstrapped onto the provider.
func (k *K8s) Bootstrap() bool { return k.bootstrap }

// CloudName reports the name of the provider as Juju sees it.
func (k *K8s) CloudName() string { return "k8s" }

// GroupName reports the name of the POSIX group with permission to use K8s.
func (k *K8s) GroupName() string { return "" }

// Credentials reports the section of Juju's credentials.yaml for the provider
func (m K8s) Credentials() map[string]interface{} { return nil }

// ModelDefaults reports the Juju model-defaults specific to the provider.
func (m *K8s) ModelDefaults() map[string]string { return m.modelDefaults }

// BootstrapConstraints reports the Juju bootstrap-constraints specific to the provider.
func (m *K8s) BootstrapConstraints() map[string]string { return m.bootstrapConstraints }

// Remove uninstalls K8s and kubectl.
func (k *K8s) Restore() error {
	snapHandler := packages.NewSnapHandler(k.system, k.snaps)

	err := snapHandler.Restore()
	if err != nil {
		return err
	}

	err = k.system.RemoveAllHome(".kube")
	if err != nil {
		return fmt.Errorf("failed to remove '.kube' from user's home directory: %w", err)
	}

	slog.Info("Removed provider", "provider", k.Name())

	return nil
}

// install ensures that K8s is installed.
func (k *K8s) install() error {
	var eg errgroup.Group

	// Prepare/restore package handlers concurrently
	debHandler := packages.NewDebHandler(k.system, k.debs)
	snapHandler := packages.NewSnapHandler(k.system, k.snaps)

	eg.Go(func() error {
		// In some cases, iptables is not present on the system. In those cases,
		// make sure it's installed.
		cmd := system.NewCommand("which", []string{"iptables"})
		_, err := k.system.Run(cmd)
		if err != nil {
			err := debHandler.Prepare()
			if err != nil {
				return err
			}
		}
		return nil
	})

	eg.Go(func() error {
		err := snapHandler.Prepare()
		if err != nil {
			return err
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}

// init ensures that K8s is installed, minimally configured, and ready.
func (k *K8s) init() error {
	if k.needsBootstrap() {
		cmd := system.NewCommand("k8s", []string{"bootstrap"})
		_, err := k.system.RunWithRetries(cmd, (5 * time.Minute))
		if err != nil {
			return err
		}
	}

	cmd := system.NewCommand("k8s", []string{"status", "--wait-ready"})
	_, err := k.system.RunWithRetries(cmd, (5 * time.Minute))

	return err
}

// configureFeatures iterates over the specified features, enabling and configuring them.
func (k *K8s) configureFeatures() error {
	for featureName, conf := range k.Features {
		for key, value := range conf {
			featureConfig := fmt.Sprintf("%s.%s=%s", featureName, key, value)

			cmd := system.NewCommand("k8s", []string{"set", featureConfig})
			_, err := k.system.Run(cmd)
			if err != nil {
				return fmt.Errorf("failed to set K8s feature config '%s': %w", featureConfig, err)
			}
		}

		cmd := system.NewCommand("k8s", []string{"enable", featureName})
		_, err := k.system.RunWithRetries(cmd, (5 * time.Minute))
		if err != nil {
			return fmt.Errorf("failed to enable K8s addon '%s': %w", featureName, err)
		}
	}

	return nil
}

// setupKubectl both installs the kubectl snap, and writes the relevant kubeconfig
// file to the user's home directory such that kubectl works with K8s.
func (k *K8s) setupKubectl() error {
	cmd := system.NewCommand("k8s", []string{"kubectl", "config", "view", "--raw"})
	result, err := k.system.Run(cmd)
	if err != nil {
		return fmt.Errorf("failed to fetch K8s configuration: %w", err)
	}

	return k.system.WriteHomeDirFile(path.Join(".kube", "config"), result)
}

func (k *K8s) needsBootstrap() bool {
	cmd := system.NewCommand("k8s", []string{"status"})
	output, err := k.system.Run(cmd)

	if err != nil && strings.Contains(string(output), "Error: The node is not part of a Kubernetes cluster.") {
		return true
	}

	return false
}
