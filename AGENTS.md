# AGENTS.md - Taskerr

This document provides guidelines for AI coding agents working in this repository.

## Project Overview

Taskerr is a terminal-based task management application written in **Go 1.25+** using:
- **CLI**: [Cobra](https://github.com/spf13/cobra)
- **TUI**: [Bubble Tea](https://github.com/charmbracelet/bubbletea) with [Lipgloss](https://github.com/charmbracelet/lipgloss) for styling
- **Database**: [GORM](https://gorm.io/) with SQLite (default), MySQL, or PostgreSQL
- **Configuration**: [Koanf](https://github.com/knadh/koanf)

## Build/Lint/Test Commands

### Build Commands (via Makefile)

```bash
make build      # Build to build/taskerr
make run        # Build and run
make clean      # Remove build artifacts
make rebuild    # Clean and rebuild
make install    # Install to ~/.local/bin/taskerr
```

### Go Commands

```bash
go mod tidy                    # Sync dependencies
go build -o build/taskerr .    # Build binary
go run main.go                 # Run directly
```

### Testing

```bash
go test ./...                           # Run all tests
go test ./internal/data/tasks/...       # Run tests in specific package
go test -v ./internal/cli/...           # Verbose output for a package
go test -run TestFunctionName ./...     # Run a single test by name
go test -run TestFunctionName/SubTest   # Run a specific subtest
go test -cover ./...                    # Run with coverage
go test -race ./...                     # Run with race detector
```

### Writing Tests

Tests use [testify](https://github.com/stretchr/testify) for assertions and mocks:

```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestExample(t *testing.T) {
    // Use require for fatal assertions (stops test on failure)
    require.NoError(t, err)
    require.NotNil(t, result)

    // Use assert for non-fatal assertions (continues test on failure)
    assert.Equal(t, expected, actual)
    assert.Contains(t, slice, element)
}
```

**Test file conventions:**
- Test files: `*_test.go` in same package as code being tested
- Test functions: `TestXxx(t *testing.T)`
- Use table-driven tests for multiple scenarios:

```go
func TestFeature(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected int
    }{
        {"empty input", "", 0},
        {"single item", "a", 1},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Feature(tt.input)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

**Test database setup:**
Use in-memory SQLite for repository tests:

```go
func setupTestDB(t *testing.T) *gorm.DB {
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Silent),
    })
    require.NoError(t, err)
    db.AutoMigrate(&Task{}, &Tag{})
    return db
}
```

### Testing Full Application

To test the full application without affecting your personal data, use environment variables to override the configuration:

```bash
# Use a test database in the repository
TASKERR_DB_CONNECTION=./test.db ./build/taskerr

# Or export for multiple commands
export TASKERR_DB_CONNECTION=./test.db
./build/taskerr add "Test task"
./build/taskerr ls
./build/taskerr done 1
unset TASKERR_DB_CONNECTION

# Run TUI with test database
TASKERR_DB_CONNECTION=./test.db ./build/taskerr
```

**Environment variables:**

| Variable | Default | Description |
|----------|---------|-------------|
| `TASKERR_DB_PROVIDER` | `sqlite` | Database driver (sqlite, mysql, postgres) |
| `TASKERR_DB_CONNECTION` | `~/.taskerr.db` | Database path or connection string |

**Note:** The `test.db` file is ignored by git. Delete it to reset test state.

### Linting (not currently configured, but recommended)

```bash
go fmt ./...                    # Format code
go vet ./...                    # Static analysis
golangci-lint run               # If installed
```

## Directory Structure

```
taskerr/
├── main.go                 # Entry point
├── go.mod / go.sum         # Go modules
├── Makefile                # Build automation
└── internal/               # Internal packages (not externally importable)
    ├── cli/                # Cobra CLI commands
    ├── config/             # Configuration loading (Koanf)
    ├── data/               # Data layer
    │   ├── service.go      # Database initialization
    │   └── tasks/          # Task and Tag models/repos
    └── tui/                # Bubble Tea TUI components
```

## Code Style Guidelines

### Import Organization

Group imports with blank line separators in this order:
1. Standard library
2. Third-party packages
3. Internal project packages

```go
import (
    "fmt"
    "log"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    "github.com/mustafmst/taskerr/internal/data"
    "github.com/mustafmst/taskerr/internal/data/tasks"
)
```

Use package aliases when appropriate (e.g., `tea` for bubbletea).

### Naming Conventions

| Element | Convention | Example |
|---------|------------|---------|
| Files | snake_case | `add_task_modal.go`, `tag_repo.go` |
| Packages | lowercase | `config`, `tasks`, `tui` |
| Exported | PascalCase | `NewTasksRepository`, `MainWindowModel` |
| Unexported | camelCase | `activeTask`, `loadTasks` |
| Constants | PascalCase | `HeaderHeight`, `TaskHeight` |
| Enums | iota pattern | `TagsPanelFocus`, `FilterOR` |

Type suffixes:
- `Model` - TUI components: `MainWindowModel`, `TasksPanelModel`
- `Repository` - Data access: `TasksRepository`, `TagsRepository`
- `Msg` - Bubble Tea messages: `TasksLoadedMsg`, `TagToggledMsg`
- `Service` - Service layer: `Service`
- `Provider` - Configuration: `ConfigProvider`

### Constructor Pattern

Use `New` prefix for constructors:

```go
func NewTasksRepository(db *gorm.DB) *TasksRepository { ... }
func NewMainWindowModel(service *data.Service) MainWindowModel { ... }
```

### Receiver Conventions

- **Pointer receivers** for methods that modify state
- **Value receivers** for read-only methods

```go
func (m *TasksPanelModel) SetFocused(focused bool) { ... }  // Modifies state
func (m TasksPanelModel) IsFocused() bool { ... }           // Read-only
```

### Error Handling

1. **Return and check immediately:**
```go
db, err := initDB(cfg)
if err != nil {
    return nil, err
}
```

2. **Wrap errors with context using `%w`:**
```go
return nil, fmt.Errorf("failed to connect to database: %w", err)
```

3. **Use `log.Fatalf` for unrecoverable startup errors:**
```go
if err != nil {
    log.Fatalf("Error initializing config: %v", err)
}
```

### Comment Style

```go
// Package config provides configuration loading via koanf.
package config

// MainWindowModel is the root model that orchestrates the TUI panels
type MainWindowModel struct { ... }

// NewMainWindowModel creates a new MainWindowModel
func NewMainWindowModel(service *data.Service) MainWindowModel { ... }
```

Use section comments to group related code:
```go
// Setters

func (m *TasksPanelModel) SetFocused(focused bool) { ... }

// Getters

func (m TasksPanelModel) IsFocused() bool { ... }
```

### GORM Model Tags

```go
type Task struct {
    ID          uint       `gorm:"primaryKey" json:"id"`
    Description string     `gorm:"not null" json:"description"`
    State       bool       `gorm:"not null" json:"state"`
    CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
    Tags        []Tag      `gorm:"many2many:task_tags;" json:"tags,omitempty"`
}
```

### Bubble Tea Patterns

Follow Model-View-Update architecture:

```go
// Model - struct with state
type MainWindowModel struct {
    width, height int
    service       *data.Service
}

// Update - handle messages with type switch
func (m MainWindowModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "q", "ctrl+c":
            return m, tea.Quit
        }
    }
    return m, nil
}

// View - pure rendering
func (m MainWindowModel) View() string { ... }
```

Use `tea.Batch` for multiple commands:
```go
return m, tea.Batch(m.loadTasks, m.loadTags, TickCmd())
```

## Configuration

- Config file: `~/.taskerr` (YAML)
- Default database: `~/.taskerr.db` (SQLite)
- Environment variables: prefix `TASKERR_`

## Application Behavior

- No args: TUI mode
- With args: CLI mode (Cobra commands)
