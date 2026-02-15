# Rolo

A terminal UI application for reordering tmux sessions using vim-like keybindings.

## Features

- Navigate sessions with `j`/`k` keys
- Delete/hide sessions with `d` key
- Repopulate or update session list with `p`/`u` keys
- Enter move mode with `m` to reorder sessions
- Save and exit with `Enter`
- Quit without saving with `q` or `Ctrl+C`
- Configurable wrap-around navigation in `~/.config/rolo/config.json`
- Persistent storage in `~/.config/rolo/rolo.json`

## Installation

```bash
go build -o rolo
```

## Configuration

Session names are stored in `~/.config/rolo/rolo.json` (one per line in JSON format):

```json
[
  {"name": "session-1", "deleted": false},
  {"name": "session-2", "deleted": false},
  {"name": "session-3", "deleted": false}
]
```

The config directory and files will be created automatically on first save.

### Settings

You can configure rolo behavior in `~/.config/rolo/config.json`:

```json
{
  "wrap_around": false
}
```

- `wrap_around`: When true, navigation wraps around (next from last session goes to first, prev from first goes to last). Default is `false`.

## Usage

### Populate from tmux

Fetch all active tmux sessions and save them to the config file:

```bash
./rolo populate
```

This will read all active tmux sessions and write them to `~/.config/rolo/rolo.json`.

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
- Use the order defined in `~/.config/rolo/rolo.json`
- Wrap around (next from last session goes to first, prev from first goes to last) if enabled in config
- Automatically skip sessions that no longer exist (marked as deleted)
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
   Use `j`/`k` to navigate, `d` to hide deleted sessions, `m` to move items, `Enter` to save

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
- `j` - Move cursor down (wraps around if enabled)
- `k` - Move cursor up (wraps around if enabled)
- `d` - Toggle delete/hide state for current session
- `p` - Repopulate list from active tmux sessions
- `u` - Update list (add new sessions, remove closed ones)
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
