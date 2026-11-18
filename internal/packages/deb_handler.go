package packages

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/canonical/concierge/internal/system"
)

// NewDeb constructs a new Deb instance.
func NewDeb(name string) *Deb {
	return &Deb{Name: name}
}

// Deb is a simple representation of a package installed from the Ubuntu archive.
type Deb struct {
	Name string
}

// NewDebHandler constructs a new instance of a DebHandler.
func NewDebHandler(system system.Worker, debs []*Deb) *DebHandler {
	return &DebHandler{
		Debs:   debs,
		system: system,
	}
}

// DebHandler can install or remove a set of debs.
type DebHandler struct {
	Debs   []*Deb
	system system.Worker
}

// aptEnv contains environment variables that prevent apt/dpkg (and tools
// hooked into them, such as needrestart) from blocking on interactive
// prompts during unattended package operations.
var aptEnv = []string{
	"DEBIAN_FRONTEND=noninteractive",
	"NEEDRESTART_MODE=a",
}

// aptCommand constructs an apt-get command that runs non-interactively so
// package operations never hang waiting for user input.
func aptCommand(args ...string) *system.Command {
	cmd := system.NewCommand("apt-get", append([]string{"-y"}, args...))
	cmd.Env = aptEnv
	return cmd
}

// Prepare updates the apt cache and installs a set of debs from the archive.
func (h *DebHandler) Prepare() error {
	if len(h.Debs) == 0 {
		return nil
	}

	err := h.updateAptCache()
	if err != nil {
		return fmt.Errorf("failed to update apt cache: %w", err)
	}

	for _, deb := range h.Debs {
		err := h.installDeb(deb)
		if err != nil {
			return fmt.Errorf("failed to install deb: %w", err)
		}
	}
	return nil
}

// Restore removes a set of debs from the machine.
func (h *DebHandler) Restore() error {
	for _, deb := range h.Debs {
		err := h.removeDeb(deb)
		if err != nil {
			return fmt.Errorf("failed to remove deb: %w", err)
		}
	}

	cmd := aptCommand("autoremove")

	_, err := system.RunExclusive(h.system, cmd)
	if err != nil {
		return fmt.Errorf("failed to install apt package: %w", err)
	}

	return nil
}

// installDeb uses `apt` to install the package on the system from the archives.
func (h *DebHandler) installDeb(d *Deb) error {
	cmd := aptCommand("install",
		"-o", "Dpkg::Options::=--force-confdef",
		"-o", "Dpkg::Options::=--force-confold",
		d.Name)

	_, err := system.RunExclusive(h.system, cmd)
	if err != nil {
		return fmt.Errorf("failed to install apt package '%s': %w", d.Name, err)
	}

	h.logInstalled(d)
	return nil
}

// logInstalled logs that a deb was installed, including the installed version
// when it can be found.
func (h *DebHandler) logInstalled(d *Deb) {
	// Skip the lookup when the log line wouldn't be shown.
	if !slog.Default().Enabled(context.Background(), slog.LevelInfo) {
		return
	}

	attrs := []any{"package", d.Name}

	// In a dry run nothing was installed, so there is no version to report.
	if !system.IsDryRun(h.system) {
		if version, err := h.installedVersion(d); err != nil {
			slog.Debug("Could not look up apt package version", "package", d.Name, "error", err)
		} else if version != "" {
			attrs = append(attrs, "version", version)
		}
	}

	slog.Info("Installed apt package", attrs...)
}

// installedVersion returns the version of the deb that dpkg has installed.
func (h *DebHandler) installedVersion(d *Deb) (string, error) {
	cmd := system.NewCommand("dpkg-query", []string{"--show", "--showformat=${Version}", d.Name})
	cmd.ReadOnly = true

	output, err := h.system.Run(cmd)
	if err != nil {
		return "", fmt.Errorf("failed to query version of apt package '%s': %w", d.Name, err)
	}

	return strings.TrimSpace(string(output)), nil
}

// Remove uninstalls the deb from the system with `apt`.
func (h *DebHandler) removeDeb(d *Deb) error {
	cmd := aptCommand("remove", d.Name)

	_, err := system.RunExclusive(h.system, cmd)
	if err != nil {
		return fmt.Errorf("failed to remove apt package '%s': %w", d.Name, err)
	}

	slog.Info("Removed apt package", "package", d.Name)
	return nil
}

// updateAptCache is a helper method to update the host's package cache.
func (h *DebHandler) updateAptCache() error {
	cmd := aptCommand("update")

	_, err := system.RunExclusive(h.system, cmd)
	if err != nil {
		return fmt.Errorf("failed to update apt package lists: %w", err)
	}

	return nil
}
