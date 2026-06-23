package cmd

import "runtime/debug"

// resolveVersion derives the user-facing version and commit strings,
// falling back to information embedded by the Go toolchain when the
// ldflag-injected build-time defaults ("dev") are still in place. This
// restores meaningful output for binaries built via
// `go install github.com/canonical/concierge@vX.Y.Z` (#59), where
// goreleaser's -X ldflags have not been applied.
//
// Precedence:
//  1. ldflag-injected values (release builds from goreleaser);
//  2. bi.Main.Version (module version recorded by `go install`);
//  3. vcs.revision + vcs.modified (in-tree `go build`).
func resolveVersion(ldflagVersion, ldflagCommit string, bi *debug.BuildInfo, ok bool) (string, string) {
	version := ldflagVersion
	commit := ldflagCommit
	if version != "dev" {
		return version, commit
	}
	if !ok || bi == nil {
		return version, commit
	}

	if v := bi.Main.Version; v != "" && v != "(devel)" {
		version = v
	}

	var rev, modified string
	for _, s := range bi.Settings {
		switch s.Key {
		case "vcs.revision":
			rev = s.Value
		case "vcs.modified":
			modified = s.Value
		}
	}
	if commit == "dev" && rev != "" {
		commit = rev
		if modified == "true" {
			commit += "-dirty"
		}
	}
	return version, commit
}
