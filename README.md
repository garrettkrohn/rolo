# Rolo

A terminal UI application for reordering tmux sessions using vim-like keybindings.

## Features

- Navigate sessions with `j`/`k` keys
- Enter move mode with `m` to reorder sessions
- Save and exit with `Enter`
- Quit without saving with `q` or `Ctrl+C`
- Persistent storage in `~/.config/rolo/rolo.txt`

## Installation

```bash
go build -o rolo
```

## Configuration

Session names are stored in `~/.config/rolo/rolo.txt` (one per line):

```
session-1
session-2
session-3
```

The config directory and file will be created automatically on first save.

## Usage

### Populate from tmux

Fetch all active tmux sessions and save them to the config file:

```bash
./rolo populate
```

This will read all active tmux sessions and write them to `~/.config/rolo/rolo.txt`.

### Interactive Mode

Launch the UI to reorder sessions:

```bash
./rolo
```

### Navigate Sessions

Switch to the next session in your ordered list:

```bash
./rolo next
```

Switch to the previous session in your ordered list:

```bash
./rolo prev
```

These commands:
- Use the order defined in `~/.config/rolo/rolo.txt`
- Wrap around (next from last session goes to first, prev from first goes to last)
- Must be run from inside a tmux session

### Help

```bash
./rolo help
```

## Workflow

1. **Initial Setup** - Populate your session list:
   ```bash
   ./rolo populate
   ```

2. **Organize** - Reorder sessions to your preference:
   ```bash
   ./rolo
   ```
   Use `j`/`k` to navigate, `m` to move items, `Enter` to save

3. **Navigate** - Switch between sessions using your custom order:
   ```bash
   ./rolo next  # Go to next session
   ./rolo prev  # Go to previous session
   ```

**Tip:** Bind these to tmux keys for quick navigation!

```tmux
# Add to ~/.tmux.conf
bind-key f run-shell "rolo next"
bind-key d run-shell "rolo prev"
```

Then use `prefix + f` to go forward (next) and `prefix + d` to go back (previous)!

## Keybindings

### Normal Mode
- `j` - Move cursor down
- `k` - Move cursor up
- `m` - Enter move mode
- `Enter` - Save order and quit
- `q` or `Ctrl+C` - Quit without saving

### Move Mode
- `j` - Move current item down
- `k` - Move current item up
- `m` - Return to normal mode
- `Enter` - Save order and quit

## Dependencies

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - Terminal UI framework

## Development

This project uses Go modules. To build:

```bash
go build
```

To run directly:

```bash
go run main.go
```
# rolo
