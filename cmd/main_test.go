package cmd

import "testing"

// versionString is what `concierge --version` prints. The interesting case is
// a `go install <module>@<version>` build: the module proxy has no VCS
// checkout to stamp, so there is a real version and no commit.
func TestVersionString(t *testing.T) {
	for name, tc := range map[string]struct {
		version string
		commit  string
		want    string
	}{
		"ldflags set both, as goreleaser does": {
			version: "v1.0.2", commit: "abc1234", want: "v1.0.2 (abc1234)",
		},
		"go install from the proxy: a version, no commit": {
			version: "v1.8.0", commit: "", want: "v1.8.0",
		},
		"nothing resolved at all": {
			version: "dev", commit: "dev", want: "dev",
		},
		"go build from a checkout stamps both": {
			version: "v1.7.1-0.20260831152542-38b0dc549391", commit: "38b0dc54",
			want: "v1.7.1-0.20260831152542-38b0dc549391 (38b0dc54)",
		},
	} {
		t.Run(name, func(t *testing.T) {
			prevVersion, prevCommit := version, commit
			version, commit = tc.version, tc.commit
			t.Cleanup(func() { version, commit = prevVersion, prevCommit })

			if got := versionString(); got != tc.want {
				t.Errorf("versionString() = %q, want %q", got, tc.want)
			}
		})
	}
}
