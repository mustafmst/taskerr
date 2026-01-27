# Taskerr - Feature Roadmap

This file tracks planned features and improvements for Taskerr.

## Legend

- [ ] Not started
- [x] Completed

---

## Critical - Core Functionality Gaps

These features address fundamental gaps in the application.

- [ ] **Due Dates & Deadlines** - Add `DueDate` field to tasks with visual indicators (overdue, due today, upcoming). Essential for task prioritization.

- [ ] **Task Priorities** - Priority levels (high/medium/low) with visual indicators and sorting options.

- [ ] **CLI Task Completion** - `taskerr done <id>` and `taskerr undone <id>` commands to toggle task state from CLI.

- [ ] **CLI Task Edit** - `taskerr edit <id> --desc="new desc" --tags=tag1,tag2` to modify tasks from CLI.

- [ ] **CLI Task Delete** - `taskerr delete <id>` or `taskerr rm <id>` to remove tasks from CLI.

---

## High - Significant Improvements

Features that significantly enhance usability and functionality.

- [ ] **Task Search/Filter** - Search tasks by description text. TUI: `/` key opens search. CLI: `taskerr search <query>`.

- [ ] **Subtasks/Checklists** - Nested subtasks under main tasks with independent completion tracking.

- [ ] **Recurring Tasks** - Daily/weekly/monthly recurring tasks that auto-create on completion.

- [ ] **Undo/Redo** - Undo last action (delete, complete, edit) with `u` key or `Ctrl+Z`.

- [ ] **Bulk Operations** - Multi-select tasks with `Shift+Space` for bulk complete/delete/tag operations.

- [ ] **Task Notes/Details** - Extended description/notes field viewable in a detail pane or modal.

- [ ] **Keyboard Shortcut Help Modal** - `?` key to display all available shortcuts in a modal.

---

## Medium - Nice-to-Have Enhancements

Features that improve organization, customization, and power-user workflows.

- [ ] **Projects/Lists** - Group tasks into projects with separate views and filtering.

- [ ] **Sort Options** - Sort by created date, due date, priority, alphabetical. `o` key to cycle sort modes.

- [ ] **Tag Edit/Rename** - Edit tag name, color, and description from TUI and CLI.

- [ ] **Export/Import** - Export to JSON/CSV, import from other task managers.

- [ ] **Archive Completed Tasks** - Move old completed tasks to archive instead of delete, with archive view.

- [ ] **Dark/Light Theme Toggle** - Switch between color themes, persist preference in config.

- [ ] **Task Reordering** - Manual reorder of tasks with `Shift+j/k` or drag-like interface.

- [ ] **Vim-style Movements** - `gg` (top), `G` (bottom), `Ctrl+d/u` (page down/up).

- [ ] **CLI Completion Scripts** - Shell completion for bash/zsh/fish.

- [ ] **Config Command** - `taskerr config show`, `taskerr config set key value` for easier configuration.

---

## Low - Future Considerations

Long-term features for future versions.

- [ ] **Sync/Cloud Backend** - Optional sync to remote server or file sync services (Dropbox, etc.).

- [ ] **Notifications/Reminders** - Desktop notifications for due tasks (separate daemon process).

- [ ] **Time Tracking** - Track time spent on tasks with start/stop timer.

- [ ] **Pomodoro Integration** - Built-in pomodoro timer for focused work sessions.

- [ ] **Calendar View** - View tasks by due date in calendar format.

- [ ] **Natural Language Input** - Parse "Buy milk tomorrow" as task with due date.

- [ ] **Task Templates** - Save and reuse task templates for repetitive task creation.

- [ ] **Markdown in Descriptions** - Rich text/markdown rendering in task descriptions.

- [ ] **Web UI** - Optional web interface using the same database.

- [ ] **Mobile Companion App** - Mobile app that syncs with the same database.

---

## Implementation Notes

### Recommended Order

1. Start with **Critical** items - Due dates and priorities transform this into a proper task manager
2. Complete CLI commands for feature parity
3. Add Search and Undo from **High** priority
4. Pick **Medium** items based on user feedback

### Database Migrations

Features requiring schema changes:
- Due Dates: Add `due_date` column to `tasks` table
- Priorities: Add `priority` column to `tasks` table
- Subtasks: Add `parent_id` column or new `subtasks` table
- Recurring: Add `recurrence_rule` column to `tasks` table
- Notes: Add `notes` text column to `tasks` table
- Archive: Add `archived` boolean or `archived_at` timestamp
- Projects: New `projects` table with foreign key in `tasks`

---

## Completed Features

_Move completed items here with completion date._

<!-- Example:
- [x] **Feature Name** - Description. (Completed: 2025-01-27)
-->
