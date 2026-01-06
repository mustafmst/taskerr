# Taskerr

Taskerr is a task management application written in Go. It provides both a Command Line Interface (CLI) and a Text User Interface (TUI) for managing tasks. Users can create, read, update, and delete tasks, with support for configuration through a config file.

## Features

- **CLI**: Manage tasks directly from the command line.
- **TUI**: A simple text-based interface for managing tasks interactively.
- **Task Management**: Create, read, update, and delete tasks.
- **Configurable**: Supports configuration through a YAML file and environment variables.
- **Database Support**: Works with SQLite, MySQL, and PostgreSQL.

## Project Structure

- `internal/config`: Handles application configuration.
- `internal/data`: Manages database connections and task data.
- `internal/cli`: Implements the CLI interface.
- `internal/tui`: Implements the TUI interface.
- `main.go`: Entry point for the application.

## Dependencies

- Go 1.25.1 or higher
- [Cobra](https://github.com/spf13/cobra) for CLI.
- [Bubbletea](https://github.com/charmbracelet/bubbletea) and [Lipgloss](https://github.com/charmbracelet/lipgloss) for TUI.
- [Koanf](https://github.com/knadh/koanf) for configuration management.
- [GORM](https://gorm.io/) for database handling.

## Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/mustafmst/taskerr.git
   cd taskerr
   ```

2. Install dependencies:

   ```bash
   go mod tidy
   ```

3. Build the application:
   ```bash
   make build
   ```

## Usage

### Run the Application

- **CLI Mode**:

  ```bash
  ./build/taskerr <command> [arguments]
  ```

  Example:

  ```bash
  ./build/taskerr add "Buy groceries"
  ./build/taskerr ls
  ```

- **TUI Mode**:
  Simply run the application without arguments:
  ```bash
  ./build/taskerr
  ```

### Clean Build Artifacts

To clean up build artifacts:

```bash
make clean
```

### Rebuild the Application

To rebuild the application:

```bash
make rebuild
```

## Configuration

The application reads configuration from:

1. A YAML file located at `~/.taskerr`.
2. Environment variables prefixed with `TASKERR_`.

Example environment variable:

```bash
export TASKERR_DB_PROVIDER=sqlite
export TASKERR_DB_CONNECTION=/path/to/database.db
```

## License
