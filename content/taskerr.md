---
title: "Taskerr"
description: "A terminal-based task manager with a fast TUI, CLI commands, colored tags, filtering, and local database storage."
date: 2026-05-21
draft: false
tags:
  - Go
  - CLI
  - TUI
  - Task Management
  - SQLite
repo: "https://github.com/mustafmst/taskerr"
install: "curl -fsSL https://raw.githubusercontent.com/mustafmst/taskerr/main/scripts/install.sh | bash"
---

# Taskerr

Taskerr is a lightweight task manager for people who live in the terminal. It combines a keyboard-first text user interface with practical CLI commands, colored tags, local storage, and simple filtering.

![Taskerr main view](/images/taskerr/taskerr_main.png)

## Why Taskerr

Most task managers either pull you out of the terminal or grow into systems with more structure than you need. Taskerr keeps the workflow local, fast, and direct: open the TUI for interactive planning, or use the CLI when you only need to add or inspect tasks quickly.

It is designed around plain task capture, tag-based organization, and fast keyboard navigation instead of project-management overhead.

## Features

- Interactive two-panel TUI with tags on the left and tasks on the right.
- CLI commands for adding tasks, listing tasks, and managing tags.
- Colored tags for lightweight organization.
- OR/AND tag filtering for broad or narrow task views.
- Hide or show completed tasks.
- Add, edit, complete, and delete tasks from the TUI.
- Task statistics with completion rate, recent completions, top tags, and 7-day activity.
- Auto-refresh when the task database changes outside the TUI.
- SQLite by default, with MySQL and PostgreSQL support.
- Configuration through `~/.taskerr` and `TASKERR_` environment variables.

## Screenshots

### Main View

![Taskerr main task view](/images/taskerr/taskerr_main.png)

### Add Task

![Taskerr add task modal](/images/taskerr/taskerr_add.png)

### Tag Filtering

![Taskerr tag filtering view](/images/taskerr/taskerr_filter.png)

### Statistics

![Taskerr statistics modal](/images/taskerr/taskerr_stats.png)

## TUI Workflow

Run `taskerr` with no arguments to launch the interactive terminal interface.

```bash
taskerr
```

The TUI opens with a two-panel layout. Tags are shown on the left, tasks are shown on the right, and the footer lists the most important keyboard shortcuts.

| Key | Action |
| --- | --- |
| `TAB` | Switch between tags and tasks |
| `j` / `k` | Move down or up |
| `n` | Add a new task |
| `e` | Edit the selected task |
| `d` | Delete the selected task or tag |
| `SPACE` | Toggle task completion or tag selection |
| `m` | Toggle tag filter mode between OR and AND |
| `h` | Hide or show completed tasks |
| `s` | Open task statistics |
| `q` | Quit |

## CLI Usage

Use the CLI for quick task capture and tag management without opening the TUI.

```bash
# Add a task
taskerr add "Buy groceries"

# Add a task with tags
taskerr add "Finish report" --tags=work,urgent

# List tasks
taskerr ls

# Create and list tags
taskerr tag create work
taskerr tag list

# Delete a tag
taskerr tag delete work

# Attach or remove a tag from an existing task
taskerr task tag 1 work
taskerr task untag 1 work
```

## Installation

Install from the repository with the provided script:

```bash
curl -fsSL https://raw.githubusercontent.com/mustafmst/taskerr/main/scripts/install.sh | bash
```

Requirements:

- `git`
- `go` 1.25.1 or newer
- `make`

The installer builds Taskerr and places the binary at `~/.local/bin/taskerr`.

## Configuration

Taskerr reads configuration from `~/.taskerr` and from environment variables prefixed with `TASKERR_`.

```bash
export TASKERR_DB_PROVIDER=sqlite
export TASKERR_DB_CONNECTION=/path/to/database.db
```

Supported database providers:

- `sqlite`
- `mysql`
- `postgres`

By default, Taskerr uses SQLite and stores data in `~/.taskerr.db`.

## Roadmap

These items are planned and are not part of the current feature set:

- Due dates and deadline indicators.
- Task priorities.
- CLI commands for completing, editing, and deleting tasks.
- Task search.
- Shell completion scripts.

## Development Note

Taskerr started as a personal utility: a simple terminal-based task manager that fit the author's workflow without becoming overly complicated. AI tools were used for much of the implementation, with manual review and modifications by the author.

## Source

Taskerr is available on GitHub: [github.com/mustafmst/taskerr](https://github.com/mustafmst/taskerr).
