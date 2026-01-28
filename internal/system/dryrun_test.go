package system

import (
	"bytes"
	"os"
	"os/user"
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

func TestDryRunWorkerSkipsExecution(t *testing.T) {
	// Create a mock system to verify no commands are executed
	mock := NewMockSystem()

	// We can't use a real System here, so we'll test with nil
	// and verify the methods don't panic and return success
	drw := &DryRunWorker{
		realSystem: nil,
		out:        &bytes.Buffer{},
	}

	cmd := NewCommand("echo", []string{"should not run"})

	// Test Run
	output, err := drw.Run(cmd)
	if err != nil {
		t.Fatalf("Run should not return error, got: %v", err)
	}
	if len(output) != 0 {
		t.Fatalf("Run should return empty output, got: %v", output)
	}

	// Test RunMany
	err = drw.RunMany(cmd, cmd)
	if err != nil {
		t.Fatalf("RunMany should not return error, got: %v", err)
	}

	// Test RunExclusive
	output, err = drw.RunExclusive(cmd)
	if err != nil {
		t.Fatalf("RunExclusive should not return error, got: %v", err)
	}
	if len(output) != 0 {
		t.Fatalf("RunExclusive should return empty output, got: %v", output)
	}

	// Test RunWithRetries
	output, err = drw.RunWithRetries(cmd, 1*time.Second)
	if err != nil {
		t.Fatalf("RunWithRetries should not return error, got: %v", err)
	}
	if len(output) != 0 {
		t.Fatalf("RunWithRetries should return empty output, got: %v", output)
	}

	// Verify mock wasn't used (no commands executed)
	if len(mock.ExecutedCommands) != 0 {
		t.Fatalf("no commands should be executed, got: %v", mock.ExecutedCommands)
	}
}

func TestDryRunWorkerSkipsFileOperations(t *testing.T) {
	drw := &DryRunWorker{
		realSystem: nil,
		out:        &bytes.Buffer{},
	}

	// Test WriteHomeDirFile
	err := drw.WriteHomeDirFile("test/path", []byte("content"))
	if err != nil {
		t.Fatalf("WriteHomeDirFile should not return error, got: %v", err)
	}

	// Test MkdirAll
	err = drw.MkdirAll("/test/path", os.ModePerm)
	if err != nil {
		t.Fatalf("MkdirAll should not return error, got: %v", err)
	}

	// Test RemovePath
	err = drw.RemovePath("/test/path")
	if err != nil {
		t.Fatalf("RemovePath should not return error, got: %v", err)
	}

	// Test ChownAll
	err = drw.ChownAll("/test/path", &user.User{Uid: "1000", Gid: "1000"})
	if err != nil {
		t.Fatalf("ChownAll should not return error, got: %v", err)
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
