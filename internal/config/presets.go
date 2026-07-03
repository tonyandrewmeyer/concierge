package config

import (
	"errors"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/canonical/concierge/presets"
	"gopkg.in/yaml.v3"
)

// ValidPresets returns the sorted list of available preset names.
func ValidPresets() []string {
	entries, err := presets.FS.ReadDir(".")
	if err != nil {
		return nil
	}

	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".yaml") {
			names = append(names, strings.TrimSuffix(e.Name(), ".yaml"))
		}
	}
	sort.Strings(names)
	return names
}

// Preset returns a configuration preset by name.
func Preset(preset string) (*Config, error) {
	filename := preset + ".yaml"
	data, err := presets.FS.ReadFile(filename)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("unknown preset '%s'", preset)
		}
		return nil, fmt.Errorf("failed to read preset '%s': %w", preset, err)
	}
	return loadPreset(data)
}

// fixNilMapEntries walks the given path of nested map keys and replaces any
// nil-valued entries in the map found there with empty maps, so that
// unmarshalling into a typed map does not silently drop bare YAML keys like
// "charmcraft:" or "ingress:".
func fixNilMapEntries(raw map[string]any, path ...string) {
	cur := any(raw)
	for _, key := range path {
		asMap, ok := cur.(map[string]any)
		if !ok {
			return
		}
		cur = asMap[key]
	}

	entries, ok := cur.(map[string]any)
	if !ok {
		return
	}
	for name, val := range entries {
		if val == nil {
			entries[name] = map[string]any{}
		}
	}
}

// fixNilYAMLEntries fixes bare YAML keys that would otherwise be silently
// dropped when unmarshalling into typed maps.
func fixNilYAMLEntries(raw map[string]any) {
	fixNilMapEntries(raw, "host", "snaps")
	fixNilMapEntries(raw, "providers", "k8s", "features")
}

// unmarshalYAMLConfig parses YAML config data into a Config. It first
// decodes into a generic map so that bare-key nil entries can be fixed up,
// then re-encodes and decodes that fixed-up data into the typed Config.
func unmarshalYAMLConfig(data []byte) (*Config, error) {
	var raw map[string]any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	fixNilYAMLEntries(raw)

	fixed, err := yaml.Marshal(raw)
	if err != nil {
		return nil, err
	}

	conf := &Config{}
	if err := yaml.Unmarshal(fixed, conf); err != nil {
		return nil, err
	}
	return conf, nil
}

// loadPreset parses YAML preset data into a Config.
func loadPreset(data []byte) (*Config, error) {
	conf, err := unmarshalYAMLConfig(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse preset: %w", err)
	}
	return conf, nil
}
