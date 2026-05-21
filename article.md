# Taskerr: a fast terminal task manager for people who already live in the shell

Most task apps want to become a destination. They want a browser tab, a mobile app, a notification strategy, and usually a subscription.

Taskerr takes the opposite route.

It is a Go-based task manager built for the terminal first: fast startup, local data, keyboard-driven navigation, and just enough structure to stay useful without becoming another system you need to manage. If your real workspace is a shell, an editor, and a few long-running terminals, Taskerr fits naturally into that environment.

It is also a small but opinionated piece of software: you can use it as a plain CLI, or launch an interactive TUI with tag filtering, task editing, statistics, and automatic refresh when the database changes outside the app.

That combination is the pitch:

- local-first
- keyboard-first
- simple enough to trust
- flexible enough to bend into unusual workflows

## Why Taskerr is interesting

Taskerr does not try to replace project management software. It is much closer to a personal operations console.

You can:

- add and manage tasks from the command line
- switch into a full-screen terminal UI for browsing and editing
- organize everything with tags
- filter tasks with OR and AND logic
- hide completed work by default
- inspect lightweight stats about progress and activity
- keep data in SQLite by default, or point it at MySQL or PostgreSQL

That last point matters more than it sounds. A lot of terminal tools are effectively single-machine toys. Taskerr starts with a local SQLite database in your home directory, but the storage layer is abstracted through GORM, so the same app can run on a different backend if your workflow grows beyond one laptop.

## The user experience: CLI when you want speed, TUI when you want context

Taskerr has two personalities.

If you run it with no arguments, it opens the TUI. You get a two-panel interface: tags on the left, tasks on the right. The task list stays focused by default, completed items are hidden by default, and nearly everything important is reachable from the keyboard.

If you run it with arguments, it behaves like a standard CLI application powered by Cobra.

That split is one of the project’s strongest design choices. A lot of tools force you to choose between “scriptable” and “pleasant to use.” Taskerr does not. You can capture tasks in one command, then review and reorganize them later in a richer interface.

## Getting started

### Install and build

The project is written in Go and ships with a small Makefile-driven workflow.

```bash
git clone https://github.com/mustafmst/taskerr.git
cd taskerr
go mod tidy
make build
```

This creates the binary at:

```bash
build/taskerr
```

You can also install it into your local bin directory with:

```bash
make install
```

### Run the TUI

Start Taskerr without arguments:

```bash
./build/taskerr
```

From there, the essential keys are:

- `TAB` to switch between the tags panel and the tasks panel
- `j` / `k` to move
- `n` to create a task
- `e` to edit the selected task
- `d` to delete the selected task or tag
- `SPACE` to toggle completion or toggle a selected tag
- `m` to switch between OR and AND tag filtering
- `h` to hide or show completed tasks
- `a` to expand the selected task card and reveal more detail
- `s` to open the statistics modal
- `q` to quit

The TUI is where Taskerr stops feeling like “just another CLI wrapper around a database.” It becomes a real terminal application.

### Use the CLI

The CLI is intentionally straightforward.

Create a task:

```bash
taskerr add "Write release notes"
```

Create a task with details and tags:

```bash
taskerr add "Ship blog post" --details="Draft, edit, publish, promote" --tags=writing,marketing
```

List tasks:

```bash
taskerr ls
```

Create tags:

```bash
taskerr tag create work
taskerr tag create urgent --color="#ff5733" --desc="Needs attention now"
```

List tags:

```bash
taskerr tag list
```

Attach or remove tags from an existing task:

```bash
taskerr task tag 12 urgent
taskerr task untag 12 urgent
```

Delete a tag:

```bash
taskerr tag delete urgent
```

## A few workflows that are better than they look

Taskerr is simple on paper, but that simplicity creates room for workflows that feel surprisingly powerful in practice.

### 1. Inbox now, organize later

Most task systems slow you down at the moment of capture. Taskerr does not need to.

Throw raw tasks into the database all day from the CLI:

```bash
taskerr add "Call supplier"
taskerr add "Fix flaky payment test"
taskerr add "Book dentist"
```

Then open the TUI once or twice a day, assign tags, add details, and clean the list with keyboard-driven batch thinking. This works especially well if your brain resists over-structuring tasks in the moment.

### 2. Use tags like perspectives, not categories

A lot of people use tags as static folders. That is usually too rigid.

Taskerr’s OR/AND filtering makes tags more interesting when they represent perspectives:

- `deep-work`
- `5min`
- `blocked`
- `waiting`
- `writing`
- `ops`

In OR mode, you can ask, “show me anything related to ops or writing.”

In AND mode, you can ask better questions, like:

- everything that is both `ops` and `blocked`
- everything that is both `writing` and `deep-work`
- everything that is both `5min` and `admin`

That turns the left panel into a decision engine instead of a taxonomy browser.

### 3. Let the TUI react to external changes

One understated feature in Taskerr is automatic refresh. The TUI polls database state and reloads when something changes.

That means you can keep the interface open in one terminal and modify tasks from another:

```bash
taskerr add "Review staging logs" --tags=ops
taskerr task tag 7 urgent
```

The TUI notices and updates itself.

That opens up a nice split-screen workflow:

- terminal 1: Taskerr TUI as your live dashboard
- terminal 2: shell scripts, ad hoc commands, or fast task capture

It feels a little like having a personal Kanban board wired directly into your shell.

### 4. Use Taskerr as a lightweight daily review tool

The stats modal is a small feature, but it changes how the app feels. Instead of only storing tasks, Taskerr can reflect your behavior back to you.

You can check:

- total vs completed tasks
- completion rate
- how much you finished today, this week, and this month
- average time to completion
- top tags
- activity over the last 7 days

That makes Taskerr useful not just for execution, but for review. It is enough telemetry to answer “what am I actually spending time on?” without turning your task list into business intelligence software.

### 5. Run one database, many interfaces

Because the app supports SQLite, MySQL, and PostgreSQL, there is room for a more unusual setup:

- use SQLite for personal local task capture
- switch to PostgreSQL for a shared or synced environment
- keep the same commands and the same UI habits

Taskerr is not pretending to be a team collaboration suite, but the storage flexibility means it does not hit a wall the moment your environment gets more serious.

## How configuration works

Configuration is intentionally minimal.

Taskerr reads settings from:

- `~/.taskerr` as a YAML config file
- environment variables prefixed with `TASKERR_`

The default configuration points to SQLite and writes data to:

```bash
~/.taskerr.db
```

A typical environment-based setup looks like this:

```bash
export TASKERR_DB_PROVIDER=sqlite
export TASKERR_DB_CONNECTION=$HOME/.taskerr.db
```

Or, for a different backend:

```bash
export TASKERR_DB_PROVIDER=postgres
export TASKERR_DB_CONNECTION="host=localhost user=taskerr dbname=taskerr password=secret sslmode=disable"
```

This is exactly the kind of configuration surface a terminal tool should have: small, obvious, and easy to automate.

## How Taskerr is built

Taskerr is a compact Go application with a clean internal split between interface, configuration, and persistence.

### CLI: Cobra

The command-line interface is built with Cobra. That gives the project a standard command tree, argument validation, and a path to add more commands later without creating a mess.

The current shape is intentionally lean:

- `ls`
- `add`
- `tag create`
- `tag list`
- `tag delete`
- `task tag`
- `task untag`

That is enough to cover fast task entry and basic automation while leaving the richer editing experience to the TUI.

### TUI: Bubble Tea and Lipgloss

The interactive interface is built around Bubble Tea’s Model-View-Update pattern, with Lipgloss handling layout and styling.

The main window coordinates:

- a tags panel
- a tasks panel
- an add/edit modal
- a delete confirmation modal
- a statistics modal

This is a good technical fit. Bubble Tea keeps state transitions explicit, and Lipgloss makes it possible to build a TUI that feels structured rather than cramped.

Taskerr also uses Bubble Tea’s message flow for background refresh. A tick command checks database state every second, compares counts and update timestamps, and reloads tasks and tags if something changed. That is a pragmatic feature with a clean implementation.

### Data layer: GORM repositories

Persistence is handled through GORM, with repositories for tasks and tags plus a service layer that coordinates cross-entity operations.

The service layer does the important higher-level work:

- create tasks
- update tasks
- sync tag assignments
- attach and detach tags
- delete tasks and tags
- capture enough database state for TUI refresh detection

The task-tag relationship is modeled as a many-to-many association, which is exactly what this application needs. It keeps the data model simple while making filters powerful.

### Configuration: Koanf

Koanf handles config loading from both file and environment, which is a solid choice for a terminal app. It keeps local defaults easy while making automation and deployment straightforward.

### Database options

Out of the box, Taskerr supports:

- SQLite
- MySQL
- PostgreSQL

SQLite as the default is the right call. It keeps the first-run experience nearly frictionless. But because the storage backend is abstracted behind GORM, the project is not painted into a corner.

## Why this project works

Taskerr works because it makes a series of disciplined choices:

- the scope is narrow
- the interfaces are practical
- the defaults are sensible
- the implementation stack is boring in the best possible way

Go, Cobra, Bubble Tea, Lipgloss, GORM, and Koanf are not flashy choices. They are operational choices. Together they produce a tool that starts fast, stays understandable, and is realistic to extend.

That matters. Personal productivity software often fails because it tries to become a lifestyle. Taskerr is better positioned because it behaves like a tool.

## Who should try it

Taskerr makes the most sense for:

- developers who already spend most of their day in a terminal
- people who want local task storage without a cloud dependency
- users who like scriptable tools but still want a visual overview
- anyone tired of bloated task apps that confuse features with value

If you want recurring tasks, collaboration workflows, notifications, calendar sync, and a polished mobile experience, this is probably not your app.

If you want a sharp, local, keyboard-friendly task manager that feels native to terminal-based work, it absolutely might be.

## Final pitch

Taskerr is not trying to be the biggest productivity platform in the room. That is precisely why it is appealing.

It gives you:

- fast task capture
- a genuinely useful TUI
- tag-based filtering that supports real decision-making
- local-first storage
- room to grow into more serious database setups

And maybe most importantly, it stays out of the way.

For a terminal-native user, that is not a minor feature. That is the product.
