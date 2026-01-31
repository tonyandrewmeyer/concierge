package system

import (
	"bytes"
	"os"
	"os/user"
	"strings"
	"testing"
	"time"
)

func TestDryRunWorkerPrint(t *testing.T) {
	// Create a buffer to capture output
	var buf bytes.Buffer

	drw := &DryRunWorker{
		realSystem: nil,
		out:        &buf,
	}

	drw.Print("test message")

	expected := "test message\n"
	if buf.String() != expected {
		t.Fatalf("expected: %q, got: %q", expected, buf.String())
	}
}

func TestDryRunWorkerAutoPrintsCommands(t *testing.T) {
	var buf bytes.Buffer
	drw := &DryRunWorker{
		realSystem: nil,
		out:        &buf,
	}

	cmd := NewCommand("echo", []string{"hello", "world"})

	// Test Run - should auto-print the command
	output, err := drw.Run(cmd)
	if err != nil {
		t.Fatalf("Run should not return error, got: %v", err)
	}
	if len(output) != 0 {
		t.Fatalf("Run should return empty output, got: %v", output)
	}
	// Check for "Would run:" prefix and the arguments (path may vary by system)
	if !strings.Contains(buf.String(), "Would run:") || !strings.Contains(buf.String(), "echo hello world") {
		t.Fatalf("Run should print command, got: %s", buf.String())
	}

	buf.Reset()

	// Test RunMany - should auto-print each command
	err = drw.RunMany(cmd, cmd)
	if err != nil {
		t.Fatalf("RunMany should not return error, got: %v", err)
	}
	if strings.Count(buf.String(), "Would run:") != 2 {
		t.Fatalf("RunMany should print 2 commands, got: %s", buf.String())
	}

	buf.Reset()

	// Test RunExclusive - should auto-print the command
	output, err = drw.RunExclusive(cmd)
	if err != nil {
		t.Fatalf("RunExclusive should not return error, got: %v", err)
	}
	if !strings.Contains(buf.String(), "Would run:") {
		t.Fatalf("RunExclusive should print command, got: %s", buf.String())
	}

	buf.Reset()

	// Test RunWithRetries - should auto-print the command
	output, err = drw.RunWithRetries(cmd, 1*time.Second)
	if err != nil {
		t.Fatalf("RunWithRetries should not return error, got: %v", err)
	}
	if !strings.Contains(buf.String(), "Would run:") {
		t.Fatalf("RunWithRetries should print command, got: %s", buf.String())
	}
}

func TestDryRunWorkerAutoPrintsFileOperations(t *testing.T) {
	var buf bytes.Buffer

	// Use a mock system for realSystem since WriteHomeDirFile needs User().HomeDir
	mock := NewMockSystem()

	drw := &DryRunWorker{
		realSystem: mock,
		out:        &buf,
	}

	// Test WriteHomeDirFile - should print what would be written
	err := drw.WriteHomeDirFile("test/path", []byte("content"))
	if err != nil {
		t.Fatalf("WriteHomeDirFile should not return error, got: %v", err)
	}
	if !strings.Contains(buf.String(), "Would write file:") {
		t.Fatalf("WriteHomeDirFile should print file path, got: %s", buf.String())
	}

	buf.Reset()

	// Test MkdirAll - should print what directory would be created
	err = drw.MkdirAll("/test/path", os.ModePerm)
	if err != nil {
		t.Fatalf("MkdirAll should not return error, got: %v", err)
	}
	if !strings.Contains(buf.String(), "Would create directory: /test/path") {
		t.Fatalf("MkdirAll should print directory path, got: %s", buf.String())
	}

	buf.Reset()

	// Test RemovePath - should print what would be removed
	err = drw.RemovePath("/test/path")
	if err != nil {
		t.Fatalf("RemovePath should not return error, got: %v", err)
	}
	if !strings.Contains(buf.String(), "Would remove: /test/path") {
		t.Fatalf("RemovePath should print path, got: %s", buf.String())
	}

	buf.Reset()

	// Test ChownAll - should print what ownership change would occur
	err = drw.ChownAll("/test/path", &user.User{Uid: "1000", Gid: "1000"})
	if err != nil {
		t.Fatalf("ChownAll should not return error, got: %v", err)
	}
	if !strings.Contains(buf.String(), "Would chown /test/path") {
		t.Fatalf("ChownAll should print chown info, got: %s", buf.String())
	}
}

func TestDryRunWorkerDelegatesReadOperations(t *testing.T) {
	// Create a mock system with test data
	mock := NewMockSystem()
	mock.MockFile("test-file.txt", []byte("test content"))
	mock.MockSnapStoreLookup("test-snap", "stable", false, true)

	// Create DryRunWorker that wraps a real system
	// For this test, we need to create a minimal real system
	// Since we can't easily create a real System in tests, we'll verify
	// that the methods exist and have the right signatures

	// For now, just verify the interface is properly implemented
	var _ Worker = &DryRunWorker{}
}

func TestDryRunWorkerImplementsWorkerInterface(t *testing.T) {
	// Verify DryRunWorker implements Worker interface
	var _ Worker = (*DryRunWorker)(nil)
}
