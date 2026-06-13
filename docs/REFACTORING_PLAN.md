# Recall CLI — Refactoring Plan & Checklist

**Goal:** Make the existing product follow 12-factor principles, be unit-testable,
modular, and easy to read/maintain — **without adding features and without
over-engineering**. The product's job is narrow: *save a command once, recall it
easily.* Every change below serves that job or removes weight that doesn't.

> Rule for this effort: **No new features.** Fix the foundation first. Advanced
> feature ideas are parked at the bottom for a later cycle.

---

## Guiding principles (so we don't over-engineer)

- Prefer **deleting** code over abstracting it.
- Add an interface/indirection only when a test or a real second implementation needs it.
- Keep `cmd/recall` thin (flag parsing + output); put logic in `internal/*`.
- One source of truth for every path and default.
- Convention: **data → stdout, human messages/prompts → stderr.**

---

## Phase 1 — Paths & config (12-Factor III: config in the environment)

The single highest-value change. Removes duplication and unblocks clean testing.

- [ ] Create `internal/paths` package as the **single source of truth** for:
      data dir, config dir, db path, git sources dir.
- [ ] Honor environment overrides with sensible defaults:
      - `RECALL_DB_PATH` → overrides db file location
      - `XDG_DATA_HOME` → defaults to `~/.local/share`, data under `…/recall`
      - `XDG_CONFIG_HOME` → defaults to `~/.config`, config under `…/recall`
- [ ] Replace hardcoded db path in `storage.go`, `config.go` (`DefaultConfig`),
      and `gitops.go` with calls into `internal/paths`.
- [ ] Remove the filesystem side effect from `config.LoadConfig` (no `MkdirAll`
      on read). Directory creation happens only on write (`config init`) or store open.
- [ ] Unit tests: env override resolution (set `RECALL_DB_PATH` / `XDG_*`, assert paths).

## Phase 2 — Extract logic out of the `main` package (modularity + testability)

- [ ] Move `findBestMatch` + `fuzzyScore` from `cmd/recall/run.go` into
      `internal/search`. Keep it pure: input `[]storage.Command` + query → ranked result(s).
- [ ] Replace the manual bubble sort with `sort.Slice`.
- [ ] In `run.go`, replace direct `os.Stderr`/`fmt.Println` with `cmd.ErrOrStderr()`/`cmd.OutOrStdout()`.
- [ ] Remove `os.Exit(exitCode)` from inside `RunE`. Return a typed
      "exit-code" error and translate it to an exit status once, in `main`.
- [ ] Apply the same exit-code handling to the root TUI path in `main.go`.
- [ ] Unit tests for `internal/search`: exact match, substring, fuzzy ordering, tag filter, no-match.

## Phase 3 — Storage simplification & correctness (de-over-engineer)

- [ ] **Delete the unused normalized tag tables**: `tags`, `command_tags`,
      `migrateTagsToTable`, `syncTags`, and the related calls. They are written
      but never read. Keep the single `commands.tags` column (matches the use case).
- [ ] Stop ignoring errors on `db.Exec` calls; return/wrap them.
- [ ] Wrap any multi-statement write in a transaction (after the tag-table removal,
      `Upsert`/`Update` become single statements — verify no remaining multi-step writes).
- [ ] Keep schema creation in one place; document that `IF NOT EXISTS` makes startup idempotent.
- [ ] Confirm existing storage tests still pass; add a test asserting tags round-trip
      through the single column.

## Phase 4 — Shell layer testability

- [ ] Make `shell.Execute` injectable enough to test: extract shell resolution
      (`$SHELL` → `/bin/sh`) and allow the command runner to be substituted in tests
      (function var or small interface — whichever is lighter).
- [ ] Make `GetLastCommand` take the history file path / reader as a parameter
      (default resolves from `$HISTFILE` → `~/.zsh_history` → `~/.bash_history`).
- [ ] Remove leftover `// TODO`, `// MVP`, and rhetorical comments in `history.go`;
      replace with one accurate doc comment on supported shells.
- [ ] Unit tests: zsh extended-history line parsing, bash plain line, empty file,
      missing file error.

## Phase 5 — Output consistency & polish

- [ ] Document and apply the stream convention (data→stdout, messages→stderr)
      across all commands; fix any command that prints status to stdout.
- [ ] Replace hand-rolled JSON in `list.go` with `encoding/json` (marshal a small
      DTO struct so field names stay stable).
- [ ] Fix tab-indented `Example:` strings in `add.go`, `delete.go`, `edit.go`
      (use plain spaces so `--help` aligns).
- [ ] Replace magic strings with named constants where it aids readability
      (`"local"` source default, `"options:"` placeholder prefix). Don't over-do it.

## Phase 6 — Tests, CI, and a short architecture note

- [ ] Ensure `go vet ./...`, `gofmt -l`, and `go test ./...` are clean and run in CI.
- [ ] Add/extend tests so every `internal/*` logic package (paths, search, storage,
      shell, placeholders, gitops) has meaningful coverage. TUI stays mostly untested
      (acceptable — it's presentation).
- [ ] Add a brief `CLAUDE.md` / `ARCHITECTURE.md`: package responsibilities, the
      stdout/stderr convention, and the env vars that configure the app.

---

## Final verification checklist (run after all phases)

**12-Factor**
- [ ] No hardcoded data/config paths remain (grep for `.local/share`, `.config`).
- [ ] App location is fully configurable via env (`RECALL_DB_PATH`, `XDG_*`).
- [ ] No filesystem side effects on config/read paths.

**Modular**
- [ ] `cmd/recall` contains only flag wiring + output formatting; no search/sort logic.
- [ ] Every path/default has exactly one source of truth.

**Unit-testable**
- [ ] `internal/search`, `internal/shell`, `internal/paths` all have tests.
- [ ] No `os.Exit` inside any `RunE`; exit codes flow through one place in `main`.
- [ ] No direct `os.Stdout`/`os.Stderr` writes inside command logic (use `cmd` writers).

**Readable / maintainable**
- [ ] Dead code removed (normalized tag tables gone).
- [ ] No ignored errors on DB writes.
- [ ] `--json` uses `encoding/json`; no hand-rolled escaping.
- [ ] `--help` output is aligned (no stray tabs).
- [ ] No leftover TODO/MVP/rhetorical comments.

**Not broken / not over-engineered**
- [ ] `go build`, `go vet`, `go test ./...` all green.
- [ ] Manual smoke test: `add`, `list`, `run`, `save`, `delete`, `sync`, TUI, `init zsh/bash`.
- [ ] No new abstractions introduced "just in case" — each interface has ≥1 test or real user.

---

## Parked for later: advanced features (do NOT build until the above is done)

These are explicitly out of scope for the refactor. Listed only so they aren't forgotten.

- Multi-shell history support (fish, PowerShell) for `recall save`.
- Workspace/project-scoped command suggestions surfaced more prominently.
- Encrypted or secret-aware fields for commands containing credentials.
- `recall export` / richer `recall import` formats.
- Fuzzy-search quality improvements (e.g. `sahilm/fuzzy`, already a transitive dep).
- Usage analytics / "frecency" ranking instead of raw usage count.
