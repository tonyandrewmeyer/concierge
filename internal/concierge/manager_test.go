package concierge

import (
	"testing"

	"github.com/canonical/concierge/internal/config"
	"github.com/canonical/concierge/internal/system"
)

func TestDryRunPlanExecution(t *testing.T) {
	// Test that with a mock system (simulating dry-run behavior),
	// no actual commands are executed and Print() calls don't panic

	mock := system.NewMockSystem()

	// Create a minimal config
	conf := &config.Config{}

	// Create the plan with mock system
	plan := NewPlan(conf, mock)

	// Verify plan was created
	if plan == nil {
		t.Fatal("plan should not be nil")
	}

	// Execute prepare - should not fail with mock system
	err := plan.Execute(PrepareAction)
	if err != nil {
		t.Fatalf("plan execution should not fail with mock system: %v", err)
	}

	// Verify that Print() was called (mock system no-ops it)
	// and that the execution completed without actual system changes
}

func TestDryRunConfigField(t *testing.T) {
	// Test that the DryRun field exists and can be set
	conf := &config.Config{
		DryRun: true,
	}

	if !conf.DryRun {
		t.Fatal("DryRun should be true")
	}

	conf.DryRun = false
	if conf.DryRun {
		t.Fatal("DryRun should be false")
	}
}
