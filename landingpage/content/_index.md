+++
date = '2026-05-21T21:25:19+02:00'
draft = false
title = 'Taskerr'
+++

# Taskerr

## Make task management simple without leaving your terminal

Taskerr is a terminal-based task manager for people who already live in the command line. It gives you a focused TUI for daily work, a CLI for scripts and quick commands, and tag-based filtering for keeping different parts of life separate without adding a full productivity system to maintain.

```bash
taskerr
taskerr add "Write release notes" --tags=work,urgent
taskerr ls
```

![Taskerr main TUI view](/images/taskerr_main.png)

## Why Taskerr?

- **Terminal-first workflow**: open the TUI with no arguments, or use CLI commands when you want speed and automation.
- **Tags that stay useful**: organize tasks with colored tags and filter them from the keyboard.
- **OR and AND filtering**: find tasks that match any selected tag, or narrow the list to tasks that match every selected tag.
- **Fast task editing**: add, edit, complete, and delete tasks without leaving the interface.
- **Built-in statistics**: see completion counts, recent activity, and top tags directly in the terminal.
- **Local by default**: SQLite works out of the box, with MySQL and PostgreSQL available through configuration.

## TUI For Daily Work

The default view is a two-panel terminal interface: tags on the left, tasks on the right. Use `TAB` to switch focus, `j` and `k` to move, and `SPACE` to toggle either a task or a tag filter.

Taskerr starts with completed tasks hidden so your active work stays visible. Press `h` to show or hide completed items, and `m` to switch between OR and AND tag filtering.

![Filtering tasks by tags in Taskerr](/images/taskerr_filter.png)

## Add And Edit Quickly

Press `n` to create a task or `e` to edit the selected task. The modal lets you select existing tags and create new comma-separated tags in the same flow.

![Adding a task in Taskerr](/images/taskerr_add.png)

## See Progress Without Leaving The App

Press `s` to open task statistics. Taskerr shows total, completed, and incomplete tasks, completion rate, recent completions, top tags, and activity over the last seven days.

![Taskerr statistics modal](/images/taskerr_stats.png)

## CLI When You Need It

Use the CLI for quick capture, scripting, or managing tasks from another terminal session.

```bash
# Add tasks
taskerr add "Buy groceries"
taskerr add "Finish report" --tags=work,urgent

# List tasks
taskerr ls

# Manage tags
taskerr tag create work
taskerr tag list
taskerr tag delete work

# Attach and remove tags
taskerr task tag <task_id> <tag_name>
taskerr task untag <task_id> <tag_name>
```

The TUI auto-refreshes when the database changes, so tasks added from the CLI can appear while the interface is open.

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/mustafmst/taskerr/main/scripts/install.sh | bash
```

Requirements: `git`, `go` 1.25.1 or newer, and `make`.

For manual installation:

```bash
git clone https://github.com/mustafmst/taskerr.git
cd taskerr
go mod tidy
make build
./build/taskerr
```

## Configuration

Taskerr reads configuration from `~/.taskerr` and environment variables prefixed with `TASKERR_`.

```bash
export TASKERR_DB_PROVIDER=sqlite
export TASKERR_DB_CONNECTION=/path/to/taskerr.db
```

SQLite is the default database provider. MySQL and PostgreSQL are also supported through the same configuration layer.

## Keyboard Shortcuts

| Key | Action |
| --- | --- |
| `TAB` | Switch between tags and tasks |
| `j` / `k` | Move down / up |
| `n` | Add a new task |
| `e` | Edit the selected task |
| `d` | Delete the selected task or tag |
| `SPACE` | Toggle task completion or tag selection |
| `m` | Toggle OR/AND tag filtering |
| `h` | Hide or show completed tasks |
| `s` | Show statistics |
| `q` | Quit |

## Project Status

Taskerr is a practical utility project built to keep terminal task management simple. The project uses Go, Cobra, Bubble Tea, Lipgloss, GORM, and Koanf.

The implementation was created with significant AI assistance and human review/modification. Treat it as a useful small tool, not a polished enterprise productivity platform.
