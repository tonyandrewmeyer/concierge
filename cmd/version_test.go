package cmd

import (
	"runtime/debug"
	"testing"
)

func TestResolveVersion(t *testing.T) {
	tests := []struct {
		name          string
		ldflagVersion string
		ldflagCommit  string
		bi            *debug.BuildInfo
		ok            bool
		wantVersion   string
		wantCommit    string
	}{
		{
			name:          "ldflags injected by goreleaser take precedence",
			ldflagVersion: "v1.0.2",
			ldflagCommit:  "a8143a89",
			bi: &debug.BuildInfo{
				Main: debug.Module{Version: "v1.0.5"},
				Settings: []debug.BuildSetting{
					{Key: "vcs.revision", Value: "deadbeef"},
				},
			},
			ok:          true,
			wantVersion: "v1.0.2",
			wantCommit:  "a8143a89",
		},
		{
			name:          "go install populates Main.Version and vcs.revision",
			ldflagVersion: "dev",
			ldflagCommit:  "dev",
			bi: &debug.BuildInfo{
				Main: debug.Module{Version: "v1.0.2"},
				Settings: []debug.BuildSetting{
					{Key: "vcs.revision", Value: "a8143a89a4c2168a153a9260f18f455b966ad490"},
				},
			},
			ok:          true,
			wantVersion: "v1.0.2",
			wantCommit:  "a8143a89a4c2168a153a9260f18f455b966ad490",
		},
		{
			name:          "in-tree go build reports (devel) and a dirty revision",
			ldflagVersion: "dev",
			ldflagCommit:  "dev",
			bi: &debug.BuildInfo{
				Main: debug.Module{Version: "(devel)"},
				Settings: []debug.BuildSetting{
					{Key: "vcs.revision", Value: "abc123"},
					{Key: "vcs.modified", Value: "true"},
				},
			},
			ok:          true,
			wantVersion: "dev",
			wantCommit:  "abc123-dirty",
		},
		{
			name:          "go install at a version with no recorded vcs revision",
			ldflagVersion: "dev",
			ldflagCommit:  "dev",
			bi: &debug.BuildInfo{
				Main: debug.Module{Version: "v1.0.2"},
			},
			ok:          true,
			wantVersion: "v1.0.2",
			wantCommit:  "dev",
		},
		{
			name:          "build info unavailable leaves defaults in place",
			ldflagVersion: "dev",
			ldflagCommit:  "dev",
			bi:            nil,
			ok:            false,
			wantVersion:   "dev",
			wantCommit:    "dev",
		},
		{
			name:          "ldflag commit injected without a version still wins",
			ldflagVersion: "v1.0.2",
			ldflagCommit:  "dev",
			bi: &debug.BuildInfo{
				Main: debug.Module{Version: "v9.9.9"},
				Settings: []debug.BuildSetting{
					{Key: "vcs.revision", Value: "deadbeef"},
				},
			},
			ok:          true,
			wantVersion: "v1.0.2",
			wantCommit:  "dev",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVersion, gotCommit := resolveVersion(tt.ldflagVersion, tt.ldflagCommit, tt.bi, tt.ok)
			if gotVersion != tt.wantVersion {
				t.Errorf("version: got %q, want %q", gotVersion, tt.wantVersion)
			}
			if gotCommit != tt.wantCommit {
				t.Errorf("commit: got %q, want %q", gotCommit, tt.wantCommit)
			}
		})
	}
}
