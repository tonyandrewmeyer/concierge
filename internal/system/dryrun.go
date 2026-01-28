package system

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"sync"
	"time"
)

// DryRunWorker is a Worker implementation that outputs what would be done
// without actually executing any commands or making any changes.
type DryRunWorker struct {
	realSystem *System
	out        io.Writer
	mu         sync.Mutex
}

// NewDryRunWorker constructs a new DryRunWorker that wraps a real System
// for read operations while skipping execution operations.
func NewDryRunWorker(realSystem *System) *DryRunWorker {
	return &DryRunWorker{
		realSystem: realSystem,
		out:        os.Stdout,
	}
}

// Print outputs a message to stdout (thread-safe).
func (d *DryRunWorker) Print(msg string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	fmt.Fprintln(d.out, msg)
}

// User returns the real user - delegates to real system.
func (d *DryRunWorker) User() *user.User {
	return d.realSystem.User()
}

// Run skips actual execution and returns success.
func (d *DryRunWorker) Run(c *Command) ([]byte, error) {
	return []byte{}, nil
}

// RunMany skips actual execution and returns success.
func (d *DryRunWorker) RunMany(commands ...*Command) error {
	return nil
}

// RunExclusive skips actual execution and returns success.
func (d *DryRunWorker) RunExclusive(c *Command) ([]byte, error) {
	return []byte{}, nil
}

// RunWithRetries skips actual execution and returns success.
func (d *DryRunWorker) RunWithRetries(c *Command, maxDuration time.Duration) ([]byte, error) {
	return []byte{}, nil
}

// WriteHomeDirFile skips actual file writing and returns success.
func (d *DryRunWorker) WriteHomeDirFile(filepath string, contents []byte) error {
	return nil
}

// ReadHomeDirFile delegates to real system for accurate conditional logic.
func (d *DryRunWorker) ReadHomeDirFile(filepath string) ([]byte, error) {
	return d.realSystem.ReadHomeDirFile(filepath)
}

// ReadFile delegates to real system for accurate conditional logic.
func (d *DryRunWorker) ReadFile(filePath string) ([]byte, error) {
	return d.realSystem.ReadFile(filePath)
}

// SnapInfo delegates to real system for accurate conditional logic.
func (d *DryRunWorker) SnapInfo(snap string, channel string) (*SnapInfo, error) {
	return d.realSystem.SnapInfo(snap, channel)
}

// SnapChannels delegates to real system for accurate conditional logic.
func (d *DryRunWorker) SnapChannels(snap string) ([]string, error) {
	return d.realSystem.SnapChannels(snap)
}

// RemovePath skips actual path removal and returns success.
func (d *DryRunWorker) RemovePath(path string) error {
	return nil
}

// MkdirAll skips actual directory creation and returns success.
func (d *DryRunWorker) MkdirAll(path string, perm os.FileMode) error {
	return nil
}

// ChownAll skips actual ownership change and returns success.
func (d *DryRunWorker) ChownAll(path string, user *user.User) error {
	return nil
}
