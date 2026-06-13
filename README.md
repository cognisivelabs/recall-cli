# Recall CLI

Your external memory for the terminal. Don't memorize commands, just recall them.

Recall is a command manager that replaces shell history search with a context-aware, team-syncable dashboard. Save commands once, find them instantly, share them with your team via Git.

## Install

### Homebrew (macOS / Linux)

```bash
brew tap cognisivelabs/tap
brew install recall
```

### Scoop (Windows)

```powershell
scoop bucket add cognisivelabs https://github.com/cognisivelabs/scoop-bucket
scoop install recall
```

### Go Install (all platforms)

```bash
go install github.com/CognisiveLabs/recall-cli/cmd/recall@latest
```

Requires Go 1.24+. Make sure `$GOPATH/bin` (usually `~/go/bin`) is in your `PATH`.

### Download Binary

Pre-built binaries for macOS, Linux, and Windows (amd64 and arm64) are available on the [Releases](https://github.com/cognisivelabs/recall-cli/releases) page.

### From Source

```bash
git clone https://github.com/cognisivelabs/recall-cli.git
cd recall-cli
go install ./cmd/recall/
```

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

Auto-resolving placeholders fill in silently. Interactive ones prompt you before executing. You can mix both in one command:

```bash
# {{branch}} resolves automatically, {{options:...}} shows a picker
recall add "git push origin {{branch}} --env={{options:dev,staging,prod}}"
```

### Workspace Awareness

Assign a workspace filter to any command (e.g. `~/work/billing-*`). When you open `recall` from a matching directory, those commands sort to the top automatically.

```bash
# Add a command scoped to a project directory
recall add "make deploy" -d "Deploy billing service" -t "deploy"
# Then edit it in the TUI (press 'e') and set workspace to: ~/work/billing-*

# Now when you're in ~/work/billing-service and open recall,
# this command appears at the top
cd ~/work/billing-service
recall
```

### Team Sync via Git (Git-Ops)

Share commands with your team through a Git repository. This is the core team workflow — your team maintains a shared repo of commands, and every developer syncs them locally.

#### Step 1: Create a shared commands repo

Create a Git repository (e.g. `github.com/your-org/ops-runbooks`) and add YAML files with your team's commands:

```yaml
# commands.yaml
commands:
  - pattern: "kubectl rollout restart deploy/{{service}} -n {{options:dev,staging,prod}}"
    description: "Restart a deployment"
    tags: [k8s, ops]
  - pattern: "docker compose logs -f {{options:api,worker,redis}}"
    description: "Tail service logs"
    tags: [docker, debug]
  - pattern: "aws s3 sync ./build s3://{{options:staging-bucket,prod-bucket}}"
    description: "Deploy static assets to S3"
    tags: [aws, deploy]
```

You can organize commands across multiple files and directories — recall scans all `.yaml` and `.yml` files in the repo (skipping `.git/`):

```
ops-runbooks/
  k8s/
    deployments.yaml
    debugging.yaml
  docker/
    compose.yaml
  aws/
    s3.yaml
    ecs.yaml
```

Both formats are supported:

**Structured format** (with `commands:` key):
```yaml
commands:
  - pattern: "kubectl get pods -n {{namespace}}"
    description: "List pods"
    tags: [k8s]
```

**Flat list format** (just an array):
```yaml
- pattern: "kubectl get pods -n {{namespace}}"
  description: "List pods"
  tags: [k8s]
```

#### Step 2: Configure the source

Generate a config file if you don't have one:

```bash
recall config init
```

This creates `~/.config/recall/config.yaml`. Edit it to add your team's repo:

```yaml
# ~/.config/recall/config.yaml
sources:
  - name: personal
    path: ~/.local/share/recall/recall.db
  - name: team-ops
    git: git@github.com:your-org/ops-runbooks.git
```

You can add multiple git sources:

```yaml
sources:
  - name: personal
    path: ~/.local/share/recall/recall.db
  - name: team-ops
    git: git@github.com:your-org/ops-runbooks.git
  - name: team-infra
    git: git@github.com:your-org/infra-commands.git
  - name: community
    git: https://github.com/someone/awesome-commands.git
```

#### Step 3: Sync

```bash
recall sync
```

On first run, this clones each git source. On subsequent runs, it pulls the latest changes. All commands are imported into your local SQLite database with the source name as a label.

```
$ recall sync
Syncing team-ops...
Imported 12 commands from team-ops
Syncing team-infra...
Imported 5 commands from team-infra
Sync complete.
```

#### Step 4: Use synced commands

Synced commands appear in the TUI dashboard with a source badge (e.g. `⟨team-ops⟩`). You can filter by source:

```bash
recall list -s team-ops        # list only team-ops commands
recall list -s team-infra      # list only infra commands
recall run "rollout" -t k8s    # search across all sources
```

#### The team workflow

```
1. Developer adds a new command to the shared repo
2. Team reviews via Pull Request
3. PR gets merged
4. Every developer runs `recall sync` to get the update
5. New command appears in everyone's dashboard
```

This keeps your team's tribal knowledge in version control — searchable, reviewable, and always up to date.

### Usage Tracking

Commands are ranked by how often you use them. Most-used commands appear at the top. The detail pane shows usage count and last-used time.

### Shell Integration

Add to your `.zshrc` or `.bashrc`:

```bash
eval "$(recall init zsh)"   # or bash
```

This gives you:

- **Ctrl+Space** / **Ctrl+R** — open the dashboard as a shell widget
- **`recall save`** — captures the last command from shell history (via the shell wrapper, not history file)
- Selected commands execute directly in your shell

To verify it loaded:

```bash
source ~/.zshrc
# You should see: "Recall Shell Integration Loaded"
```

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
| `recall config path` | Print config file location |
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
recall save -c "explicit command"                   # save a specific command
```

## Architecture

```
cmd/recall/          CLI entry point (Cobra commands)
internal/
  config/            YAML config loading (~/.config/recall/)
  gitops/            Git sync + YAML import
  paths/             Filesystem location resolution (XDG + env overrides)
  placeholders/      {{placeholder}} parsing and auto-resolution
  search/            Fuzzy command search and ranking (pure, no I/O)
  shell/             Shell history, integration scripts, command execution
  storage/           SQLite storage layer + tag utilities
  tui/               Bubbletea TUI (dashboard, form, resolver, styles)
  workspace/         Working directory detection and matching
```

**Tech stack:** Go, Cobra, Bubbletea, Lipgloss, SQLite.

**Data locations** (all overridable via environment):

| | Linux / macOS | Windows |
|---|---|---|
| Database | `~/.local/share/recall/recall.db` | `%APPDATA%\recall\recall.db` |
| Config | `~/.config/recall/config.yaml` | `%APPDATA%\recall\config.yaml` |
| Synced repos | `~/.local/share/recall/sources/` | `%APPDATA%\recall\sources\` |

Override any location with environment variables (work on all platforms):

| Variable | Overrides |
|---|---|
| `RECALL_DB_PATH` | Database file path (full path) |
| `XDG_DATA_HOME` | Data directory root (database + synced repos) |
| `XDG_CONFIG_HOME` | Config directory root |

> **Windows upgrade note:** versions before v0.x stored data under `~\.local\share\recall\` (a Unix-style path). If you have existing data there, move it to `%APPDATA%\recall\` or set `XDG_DATA_HOME` to point at the old location.

## License

MIT
