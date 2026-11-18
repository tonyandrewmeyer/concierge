package packages

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/canonical/concierge/internal/system"
)

func TestDebHandlerCommands(t *testing.T) {
	type test struct {
		testFunc func(d *DebHandler)
		expected []string
	}

	tests := []test{
		{
			func(d *DebHandler) { _ = d.Prepare() },
			[]string{
				"DEBIAN_FRONTEND=noninteractive NEEDRESTART_MODE=a apt-get -y update",
				"DEBIAN_FRONTEND=noninteractive NEEDRESTART_MODE=a apt-get -y install -o Dpkg::Options::=--force-confdef -o Dpkg::Options::=--force-confold cowsay",
				"dpkg-query --show '--showformat=${Version}' cowsay",
				"DEBIAN_FRONTEND=noninteractive NEEDRESTART_MODE=a apt-get -y install -o Dpkg::Options::=--force-confdef -o Dpkg::Options::=--force-confold python3-venv",
				"dpkg-query --show '--showformat=${Version}' python3-venv",
			},
		},
		{
			func(d *DebHandler) { _ = d.Restore() },
			[]string{
				"DEBIAN_FRONTEND=noninteractive NEEDRESTART_MODE=a apt-get -y remove cowsay",
				"DEBIAN_FRONTEND=noninteractive NEEDRESTART_MODE=a apt-get -y remove python3-venv",
				"DEBIAN_FRONTEND=noninteractive NEEDRESTART_MODE=a apt-get -y autoremove",
			},
		},
	}

	debs := []*Deb{
		NewDeb("cowsay"),
		NewDeb("python3-venv"),
	}

	for _, tc := range tests {
		system := system.NewMockSystem()
		tc.testFunc(NewDebHandler(system, debs))

		if !reflect.DeepEqual(tc.expected, system.ExecutedCommands) {
			t.Fatalf("expected: %v, got: %v", tc.expected, system.ExecutedCommands)
		}
	}
}

func TestDebHandlerLogsInstalledVersion(t *testing.T) {
	logs := captureLogs(t)

	sys := system.NewMockSystem()
	sys.MockCommandReturn("dpkg-query --show '--showformat=${Version}' cowsay", []byte("3.03+dfsg2-8\n"), nil)

	err := NewDebHandler(sys, []*Deb{NewDeb("cowsay")}).Prepare()
	if err != nil {
		t.Fatal(err)
	}

	expected := `msg="Installed apt package" package=cowsay version=3.03+dfsg2-8`
	if !strings.Contains(logs.String(), expected) {
		t.Fatalf("expected log to contain %q, got: %s", expected, logs.String())
	}
}

func TestDebHandlerLogsWithoutVersionWhenLookupFails(t *testing.T) {
	logs := captureLogs(t)

	sys := system.NewMockSystem()
	sys.MockCommandReturn("dpkg-query --show '--showformat=${Version}' cowsay", nil, fmt.Errorf("dpkg-query failed"))

	err := NewDebHandler(sys, []*Deb{NewDeb("cowsay")}).Prepare()
	if err != nil {
		t.Fatalf("a failed version lookup should not fail the install: %v", err)
	}

	output := logs.String()
	if !strings.Contains(output, `msg="Installed apt package" package=cowsay`) || strings.Contains(output, "version=") {
		t.Fatalf("expected an install log line without a version, got: %s", output)
	}
}
