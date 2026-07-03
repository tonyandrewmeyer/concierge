# Exploration: replace `spf13/viper` with direct `gopkg.in/yaml.v3`

## Summary

This branch removes `github.com/spf13/viper` entirely and replaces every
call site with direct `gopkg.in/yaml.v3` unmarshalling plus small stdlib
helpers (`os.LookupEnv`, `os.ReadFile`) for the behaviour viper was
providing beyond plain YAML decoding. **The diff is complete** — every
viper call site in the tree is handled, not just `presets.go`.

The originating audit (`non-roadmap/tweak-dependabot/audit/concierge.md`,
2026-06-22) scoped this as a ~30 LOC change confined to `presets.go`. That
undersold it: viper is also the backbone of `internal/config/config.go`,
providing config-file search, env-var-driven CLI flag overrides, and a
package-level global config store. The actual diff touches both files and
is closer to ~150 LOC net, not ~30. It is still a mechanical,
non-load-bearing swap — no behaviour had to be dropped — but the team
should recalibrate the cost estimate before deciding.

## Call sites

**Touched:**
- `internal/config/presets.go` — `loadPreset` used a fresh `viper.New()`
  instance to parse preset YAML into `Config`, with `fixNilMapEntries`/
  `fixNilYAMLEntries` patching bare YAML keys (e.g. `charmcraft:`) so they
  don't get silently dropped as nil. Replaced with a shared
  `unmarshalYAMLConfig` helper: decode into `map[string]any`, run the same
  nil-fixup logic (now walking real nested maps instead of viper's dotted
  key `Get`/`Set`), re-marshal, then decode into the typed `Config` via
  `yaml.v3`.
- `internal/config/config.go` — this was the larger surface, not called
  out by the audit:
  - `init()` set up the global viper singleton (`SetConfigType`,
    `SetConfigName`, `AddConfigPath`, `SetEnvPrefix`,
    `SetEnvKeyReplacer`, `AutomaticEnv`). Replaced with two constants
    (`envPrefix = "CONCIERGE"`, `defaultConfigFileName =
    "concierge.yaml"`) and no init-time global state.
  - `parseConfig` used `viper.ReadConfig`/`viper.ReadInConfig` +
    `viper.ConfigFileNotFoundError` to locate/parse the config file and
    fall back to the `dev` preset when none is found. Replaced with
    `os.ReadFile` + `errors.Is(err, os.ErrNotExist)`, then the same
    `unmarshalYAMLConfig` helper used by presets.
  - `envOrFlagBool` / `envOrFlagString` / `envOrFlagSlice` used
    `viper.GetBool`/`viper.GetString` to read env-var overrides for CLI
    flags. Replaced with `os.LookupEnv` + `strconv.ParseBool` for the
    bool case, preserving the exact existing semantics (see behaviour
    deltas below).
  - `bindFlags` used `viper.BindEnv` + `viper.IsSet`/`viper.Get` to apply
    an environment-variable value onto any unset flag. Replaced with a
    direct `os.LookupEnv(flagToEnvVar(f.Name))` check — same effect,
    since (see below) viper's config-file store was never actually
    populated at the point `bindFlags` ran, so it was only ever reading
    from the environment anyway.
  - `flagToEnvVar` used `viper.GetEnvPrefix()`; replaced with the new
    `envPrefix` constant.
- `internal/config/config_format.go` — `Config` and all nested structs
  used `mapstructure:"..."` tags for viper's `Unmarshal`. Replaced with
  `yaml:"..."` tags of the same key names, since `yaml.v3` doesn't read
  `mapstructure` tags — see behaviour delta below re: the runtime status
  cache file.
- `internal/config/config_test.go` — three tests drove the global viper
  singleton directly (`viper.Reset()`, `SetConfigFile`, `ReadInConfig`,
  `Unmarshal`) to build a `Config` from a temp YAML file. Rewritten to
  call the already-unexported `parseConfig(path)` helper directly,
  matching the pattern the other two config tests in the same file
  already used. `TestFlagToEnvVar` no longer needs `viper.SetEnvPrefix`
  since the prefix is now a compile-time constant.

**Not touched (no viper usage, confirmed by tree-wide grep):**
- `providers/google.go`, `internal/juju/juju.go`,
  `internal/concierge/manager.go` — already used `gopkg.in/yaml.v3`
  directly for credentials/status marshalling; audit correctly marked
  these `keep-justified` and out of scope here.
- `cmd/*.go` — consume `internal/config.NewConfig`; no direct viper
  reference, no changes needed.

A repo-wide `grep -rn viper` after these changes returns no matches.

## Behaviour deltas

1. **Config-file search is narrower.** Viper's `SetConfigName("concierge")`
   + `AddConfigPath(".")` would, in principle, search `.` for any
   supported-extension file named `concierge.*` (yaml, yml, json, toml,
   ...). The replacement only looks for the literal file `concierge.yaml`
   in the current directory. This matches the one form documented in
   `README.md` (`concierge.yaml`); no test or docs reference another
   extension. Flagging as a delta in case anyone relies on
   `concierge.yml`/`.json`/`.toml` undocumented.
2. **Env-var-driven flag override for undocumented flags is preserved,
   not dropped.** `bindFlags` applies `CONCIERGE_<FLAG_NAME>` to *any*
   unset flag on the command (verbose, trace, dry-run, config, preset —
   not just the documented override table in the README). This was true
   under viper (via `AutomaticEnv`, without needing an explicit
   `BindEnv`) and remains true here via direct `os.LookupEnv`. No change,
   but noting it since it's easy to assume this only covered the
   documented override flags.
3. **Config-file values can no longer "leak" into flag overrides via a
   stray top-level key.** In principle, viper's `IsSet`/`Get` in
   `bindFlags` could have been satisfied by a value from the *loaded
   config file* as well as an env var, if the file had a top-level key
   exactly matching a flag name (e.g. a stray `disable-juju: true` at the
   YAML root, outside the documented schema). In practice this never
   happened: `bindFlags(cmd)` runs *before* `parseConfig`/`Preset` in
   `NewConfig`, so at the point `bindFlags` reads viper's store, no
   config file had been loaded into it yet — the only live source was
   ever the environment. So this delta is theoretical, not a real
   regression; confirmed by reading the call order in `NewConfig`.
4. **Bool env-var override quirk preserved verbatim.**
   `envOrFlagBool`'s existing logic only overrides the flag value when
   the env var parses as **truthy** (`if v := viper.GetBool(key); v {
   value = v }` — never forces `false`). This is arguably a latent quirk
   in the *original* code (you can't use an env var to force a boolean
   override back to `false`), but it predates this branch and is
   preserved exactly rather than "fixed" as part of a mechanical swap —
   the team should decide separately whether to change this behaviour.
5. **Internal status-cache file key naming changes.** `Config` fields
   previously had no `yaml` tags for most nested fields (only
   `mapstructure` tags, plus explicit `yaml:"-"` on three runtime-only
   fields), so `internal/concierge/manager.go`'s
   `yaml.Marshal(m.config)` — used to persist `.cache/concierge/concierge.yaml`
   status between `prepare` and `restore` — was already writing that
   file using `yaml.v3`'s default lowercase-no-separator field names
   (e.g. `agentversion`, not `agent-version`), since it never went
   through viper. Adding explicit `yaml:"agent-version"`-style tags
   (required so the *user-facing* config/preset files decode correctly)
   changes those same internal cache-file key names. Marshal and
   Unmarshal both use the same struct, so this is self-consistent for
   any single binary version — but a machine that ran `prepare` with the
   old binary and then `restore` with this branch's binary (before a
   fresh `prepare`) would silently lose its cached runtime config on
   read, since the on-disk keys wouldn't match the new tags. This is an
   internal, non-user-facing file, but it's a real upgrade-boundary
   compatibility question the team should weigh in on.

No env-var override, default-value injection, or config-file-search
behaviour that's part of the *documented* contract (README's "Config
File" and "Overrides" sections) is dropped by this branch.

## Dep tree delta

- **`go.mod` direct deps:** 8 → 7 (removed `github.com/spf13/viper`).
- **`go.mod` indirect deps:** 14 → 5.
- **`go.sum` line count:** 71 → 42 (29 lines removed).
- **Transitive deps that fell out** (confirmed via `go mod tidy` diff,
  matches the audit's prediction almost exactly):
  - `github.com/fsnotify/fsnotify`
  - `github.com/go-viper/mapstructure/v2`
  - `github.com/pelletier/go-toml/v2`
  - `github.com/sagikazarmark/locafero`
  - `github.com/sourcegraph/conc`
  - `github.com/spf13/afero`
  - `github.com/spf13/cast`
  - `github.com/subosito/gotenv`
  - `go.yaml.in/yaml/v3`
  - `golang.org/x/text`
  (10 of 10 predicted transitives confirmed gone.)
- **One new indirect dep appeared:** `github.com/kr/pretty` (pulled in
  by `gopkg.in/yaml.v3`'s own `go.mod` test dependencies, via Go's module
  graph pruning rules — it is not compiled into the concierge binary,
  just listed for build-list completeness). Net transitive count still
  drops sharply (14 → 5 indirect entries).

## Test suite status

`go build ./...` — passes.
`go test ./...` — all packages pass, **no `TODO(viper-drop)` markers
needed**. The three tests that drove viper's global singleton directly
were rewritten to call the same unexported `parseConfig` helper the other
config tests already used; no test asserted viper-specific behaviour that
had to be preserved as a documented gap.

`go vet ./...` and `gofmt -l .` are both clean.

Additionally manually verified outside the committed test suite (via a
temporary, since-removed test) that:
- an absent `concierge.yaml` correctly falls back to the `dev` preset,
- a present `concierge.yaml` in the working directory is correctly
  found and parsed.

## Compare URL

https://github.com/canonical/concierge/compare/main...tonyandrewmeyer:feat/inline-spf13-viper

---

**This branch is exploratory and NOT ready to merge. The team decides
whether to accept the behaviour deltas above.**
