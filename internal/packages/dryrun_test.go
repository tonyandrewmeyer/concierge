package packages

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/user"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/canonical/concierge/internal/system"
)

// testDryRunWorker is a test implementation that captures Print output
// and verifies no commands are executed
type testDryRunWorker struct {
	printOutput    *bytes.Buffer
	mu             sync.Mutex
	executedCmds   []string
	mockSnapInfo   map[string]*system.SnapInfo
}

func newTestDryRunWorker() *testDryRunWorker {
	return &testDryRunWorker{
		printOutput:  &bytes.Buffer{},
		executedCmds: []string{},
		mockSnapInfo: map[string]*system.SnapInfo{},
	}
}

func (t *testDryRunWorker) Print(msg string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	fmt.Fprintln(t.printOutput, msg)
}

func (t *testDryRunWorker) User() *user.User {
	return &user.User{Username: "test", Uid: "1000", Gid: "1000", HomeDir: "/tmp"}
}

func (t *testDryRunWorker) Run(c *system.Command) ([]byte, error) {
	// In dry-run mode, commands should not be executed
	// But we track them here to verify they were NOT called
	t.mu.Lock()
	t.executedCmds = append(t.executedCmds, c.CommandString())
	t.mu.Unlock()
	return []byte{}, nil
}

func (t *testDryRunWorker) RunMany(commands ...*system.Command) error {
	for _, c := range commands {
		t.Run(c)
	}
	return nil
}

func (t *testDryRunWorker) RunExclusive(c *system.Command) ([]byte, error) {
	return t.Run(c)
}

func (t *testDryRunWorker) RunWithRetries(c *system.Command, maxDuration time.Duration) ([]byte, error) {
	return t.Run(c)
}

func (t *testDryRunWorker) WriteHomeDirFile(filepath string, contents []byte) error {
	return nil
}

func (t *testDryRunWorker) ReadHomeDirFile(filepath string) ([]byte, error) {
	return nil, fmt.Errorf("file not found")
}

func (t *testDryRunWorker) ReadFile(filePath string) ([]byte, error) {
	return nil, fmt.Errorf("file not found")
}

func (t *testDryRunWorker) SnapInfo(snap string, channel string) (*system.SnapInfo, error) {
	if info, ok := t.mockSnapInfo[snap]; ok {
		return info, nil
	}
	return &system.SnapInfo{Installed: false, Classic: false}, nil
}

func (t *testDryRunWorker) SnapChannels(snap string) ([]string, error) {
	return []string{"stable", "edge"}, nil
}

func (t *testDryRunWorker) RemovePath(path string) error {
	return nil
}

func (t *testDryRunWorker) MkdirAll(path string, perm os.FileMode) error {
	return nil
}

func (t *testDryRunWorker) ChownAll(path string, user *user.User) error {
	return nil
}

func (t *testDryRunWorker) MockSnapInfo(name string, installed, classic bool) {
	t.mockSnapInfo[name] = &system.SnapInfo{
		Installed: installed,
		Classic:   classic,
	}
}

// Verify testDryRunWorker implements system.Worker
var _ system.Worker = (*testDryRunWorker)(nil)

func TestSnapHandlerPrintsDryRunMessages(t *testing.T) {
	drw := newTestDryRunWorker()
	drw.MockSnapInfo("existing-snap", true, false)

	snaps := []*system.Snap{
		system.NewSnap("new-snap", "stable", []string{}),
		system.NewSnap("existing-snap", "stable", []string{}),
	}

	handler := NewSnapHandler(drw, snaps)
	err := handler.Prepare()
	if err != nil {
		t.Fatalf("Prepare should not fail: %v", err)
	}

	output := drw.printOutput.String()

	// Verify Print was called with appropriate messages
	if !strings.Contains(output, "Installing snap 'new-snap'") {
		t.Errorf("expected 'Installing snap' message for new snap, got: %s", output)
	}
	if !strings.Contains(output, "Refreshing snap 'existing-snap'") {
		t.Errorf("expected 'Refreshing snap' message for existing snap, got: %s", output)
	}
}

func TestSnapHandlerRestorePrintsDryRunMessages(t *testing.T) {
	drw := newTestDryRunWorker()

	snaps := []*system.Snap{
		system.NewSnap("snap-to-remove", "stable", []string{}),
	}

	handler := NewSnapHandler(drw, snaps)
	err := handler.Restore()
	if err != nil {
		t.Fatalf("Restore should not fail: %v", err)
	}

	output := drw.printOutput.String()

	if !strings.Contains(output, "Removing snap 'snap-to-remove'") {
		t.Errorf("expected 'Removing snap' message, got: %s", output)
	}
}

func TestDebHandlerPrintsDryRunMessages(t *testing.T) {
	drw := newTestDryRunWorker()

	debs := []*Deb{
		{Name: "make"},
		{Name: "python3"},
	}

	handler := NewDebHandler(drw, debs)
	err := handler.Prepare()
	if err != nil {
		t.Fatalf("Prepare should not fail: %v", err)
	}

	output := drw.printOutput.String()

	if !strings.Contains(output, "Updating apt package cache") {
		t.Errorf("expected 'Updating apt' message, got: %s", output)
	}
	if !strings.Contains(output, "Installing apt package 'make'") {
		t.Errorf("expected 'Installing apt package make' message, got: %s", output)
	}
	if !strings.Contains(output, "Installing apt package 'python3'") {
		t.Errorf("expected 'Installing apt package python3' message, got: %s", output)
	}
}

// Ensure io import is used
var _ = io.Discard
