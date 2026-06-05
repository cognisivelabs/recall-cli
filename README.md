# Recall CLI

Your external memory for the terminal. Don't memorize commands, just recall them.

Recall is a command manager that replaces shell history search with a context-aware, team-syncable dashboard. Save commands once, find them instantly, share them with your team via Git.

## Install

```bash
go install github.com/CognisiveLabs/recall-cli/cmd/recall@latest
```

Requires Go 1.24+.

## Quick Start

```bash
# Add a command
recall add "kubectl logs -f deploy/{{service}} -n {{options:dev,staging,prod}}"

# Open the dashboard (fuzzy search, browse, execute)
recall

# Run a saved command by keyword
recall run "kubectl logs"

# Save the last command you ran
recall save
```

## Features

### TUI Dashboard

Run `recall` to open a split-pane terminal UI with fuzzy search, detail view, and keyboard shortcuts.

- **Fuzzy search** across commands, descriptions, and tags
- **Adaptive layout** — full split-pane, compact single-pane, or resize prompt depending on terminal size
- **Tag filtering** — press `t` to filter by tag
- **Inline actions** — `a` add, `e` edit, `x` delete, `enter` run

### Smart Placeholders

Save commands with placeholders that resolve at runtime:

```
kubectl logs {{service}} -n {{options:dev,staging,prod}}
git push origin {{branch}}
ssh {{user}}@{{host}}
```

| Placeholder | Behavior |
|---|---|
| `{{name}}` | Prompts for text input |
| `{{options:a,b,c}}` | Shows a picker |
| `{{branch}}` | Auto-resolves to current git branch |
| `{{cwd}}` | Auto-resolves to current directory |
| `{{dir}}` | Auto-resolves to directory basename |
| `{{user}}` | Auto-resolves to OS username |
| `{{host}}` | Auto-resolves to hostname |
| `{{home}}` | Auto-resolves to home directory |

Auto-resolving placeholders fill in silently. Interactive ones prompt you before executing.

### Workspace Awareness

Assign a workspace filter to any command (e.g. `~/work/billing-*`). When you open `recall` from a matching directory, those commands sort to the top automatically.

### Team Sync via Git

Share commands with your team through a Git repo. Add a git source to your config:

```yaml
# ~/.config/recall/config.yaml
sources:
  - name: personal
    path: ~/.local/share/recall/recall.db
  - name: team-ops
    git: git@github.com:my-org/ops-runbooks.git
```

Then run `recall sync` to pull and import. The repo uses simple YAML files:

```yaml
# commands.yaml
commands:
  - pattern: "kubectl rollout restart deploy/{{service}}"
    description: "Restart a deployment"
    tags: [k8s, ops]
  - pattern: "docker compose -f docker-compose.{{options:dev,prod}}.yml up -d"
    description: "Start services for an environment"
    tags: [docker]
```

### Usage Tracking

Commands are ranked by how often you use them. Most-used commands appear at the top. The detail pane shows usage count and last-used time.

### Shell Integration

Add to your `.zshrc` or `.bashrc`:

```bash
eval "$(recall init zsh)"   # or bash
```

This gives you:

- **Ctrl+Space** / **Ctrl+R** — open the dashboard as a shell widget
- **`recall save`** — captures the last command from shell history
- Selected commands execute directly in your shell

## Commands

| Command | Description |
|---|---|
| `recall` | Open the TUI dashboard |
| `recall add [cmd]` | Add a command (interactive or with `-d` and `-t` flags) |
| `recall save` | Save the last shell command |
| `recall run <query>` | Find, resolve, and execute a command |
| `recall list` | List all commands (table or `--json`) |
| `recall delete <id\|pattern>` | Delete a command |
| `recall sync` | Pull git sources and import commands |
| `recall config init` | Generate a starter config file |
| `recall config show` | Print current configuration |
| `recall init [shell]` | Print shell integration script |

### Aliases

- `recall ls` = `recall list`
- `recall rm` = `recall delete`

### Flags

```
recall add "cmd" -d "description" -t "tag1,tag2"   # non-interactive add
recall run "query" --dry                            # print without executing
recall run "query" -t k8s                           # filter by tag
recall list -t docker                               # filter by tag
recall list -s team-ops                             # filter by source
recall list --json                                  # JSON output
recall delete 3 -f                                  # skip confirmation
```

## Architecture

```
cmd/recall/          CLI entry point (Cobra commands)
internal/
  config/            YAML config loading (~/.config/recall/)
  gitops/            Git sync + YAML import
  placeholders/      {{placeholder}} parsing and auto-resolution
  shell/             Shell history reading, integration scripts, exec
  storage/           SQLite storage layer + tag utilities
  tui/               Bubbletea TUI (dashboard, form, resolver, styles)
  workspace/         Working directory detection and matching
```

**Tech stack:** Go, Cobra, Bubbletea, Lipgloss, SQLite.

**Data storage:** `~/.local/share/recall/recall.db` (SQLite).

**Config:** `~/.config/recall/config.yaml` (optional).

## License

MIT
