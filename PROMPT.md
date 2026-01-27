Project Prompt: Build "Recall" 🧠
Role: You are a Senior Tooling Engineer specializing in Go and Terminal User Interfaces (TUIs).

Objective: Build RecallCLI (also known as RecallHub or RecallCmd), a modern, beautiful, and "Git-Ops" native command manager for DevOps engineers and Developers. It is designed to replace tools like pet and history by focusing onUX, Team Sync, and Context Awareness.

1. The Core Philosophy
most snippet managers are just "searchable text files". Recall is a Product.

Project Name: RecallCLI 🧠 (Your external memory for the Terminal. Don't memorize it, just Recall it.)
The prompt outlines the tech stack (Go/Bubbletea), key differentiators (Workspaces, Git-Ops), and the core "Magic" workflows.

Hybrid: It supports both "Ghost Text" (speed) and "Dashboard" (discovery).
Team Native: It syncs via Git (standard Pull Requests) so teams share "Spells" (commands).
2. Tech Stack
Language: Go (Golang) - for performance and single-binary distribution.
TUI Framework: Bubbletea (Model-View-Update architecture).
Styling: Lipgloss.
Database: SQLite (embedded) - for structured relationships (Tags, Workspaces, Usage Stats).
Shell Integration: Zsh & Bash scripts (for Ghost Text & Keybindings).
3. Key Features & Requirements
A. The Dashboard (TUI)
Access: Triggered via Ctrl+Space (or similar).
Layout:
Sidebar: List of "Books" (Global, Project-Specific, Team-Backend).
Main Pane: List of commands with icons + fuzzy search.
Detail Pane: Full Markdown rendering of the command description, usage examples, and flag explanations.
B. "Magical" Workflows
Magic Save (recall save):
When run, it grabs the last executed command from shell history instantly.
Opens a TUI modal to add a description and tags.
Smart Placeholders:
Support {{branch}}, {{file}}, {{options:dev,prod}}.
When the command runs, pop up a list picker or file browser instead of a raw text input.
C. Context Awareness (Workspaces)
Detect the current directory.
If inside /work/billing-service, automatically filter the dashboard to show the "Billing" tag/workspace first.
D. Team Sync (Git-Ops)
Config allows multiple "Sources":
sources:
  - name: "personal"
    path: "~/.local/share/recall/personal.db"
  - name: "team-ops"
    git: "git@gitlab.com:my-org/ops-runbooks.git" # Auto-pulls on startup
4. Interaction Model
Ghost Text: Hook into Zsh autosuggestions. If a user types kubectl logs, suggest their most used styled command in faint text. Right arrow to complete.
Dashboard: If they don't remember the start, Ctrl+Space opens the full UI to browse.
5. Definition of Done (MVP)
 A working Go CLI with cobra.
 A Bubbletea TUI list with fuzzy filtering.
 SQLite storage layer.
 recall save works to grab shell history.
 recall sync pulls from a sample Git repo.
6. Future Roadmap (Not in MVP) 🚀
These features are explicitly out of scope for the initial build but should be kept in mind for architecture:

AI Command Generation: Natural language to command generation (e.g., "how to crop video").
Runbooks: Chaining multiple commands into an interactive sequence/script.
Secret Management: Integration with 1Password/Vault to inject secrets at runtime ({{op://...}}).
Interactive Wizards: UI-driven form builders for complex command flags.
