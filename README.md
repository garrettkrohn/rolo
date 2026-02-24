# Rolo

A terminal UI for organizing and navigating tmux sessions with vim-like keybindings.

## Installation

### Homebrew

```bash
brew install garrettkrohn/rolo/rolo
```

### From Source

```bash
go install github.com/garrettkrohn/rolo@latest
```

Or build locally:

```bash
git clone https://github.com/garrettkrohn/rolo
cd rolo
go build
```

## Quick Start

1. Populate your session list:
   ```bash
   rolo populate
   ```

2. Launch the UI to reorder sessions:
   ```bash
   rolo
   ```

3. Navigate between sessions:
   ```bash
   rolo next  # Switch to next session
   rolo prev  # Switch to previous session
   ```

## Tmux Integration

Add to `~/.tmux.conf`:

```tmux
bind-key r run-shell "tmux popup -E -w 40% -h 60% 'rolo'"
bind-key j run-shell "rolo next"
bind-key k run-shell "rolo prev"
```

## Keybindings

### Normal Mode
- `j/k` - Navigate up/down
- `o` - Attach to session
- `d` - Toggle delete/hide session
- `D` - Kill tmux session
- `u` - Update list (sync with active sessions)
- `p` - Repopulate list (replace with active sessions)
- `m` - Enter move mode
- `s/Enter` - Save and quit
- `q/Esc/Ctrl+C` - Quit without saving

### Move Mode
- `j/k` - Move item up/down
- `m` - Return to normal mode
- `s/Enter` - Save and quit

## Configuration

Settings are stored in `~/.config/rolo/config.json`:

```json
{
  "wrap_around": false
}
```

- `wrap_around`: Enable wrap-around navigation (default: `false`), for example
if you're at the bottom session in a stack, when you hit next, does it go back
to the top or do nothing.

Session data is stored in `~/.config/rolo/rolo.json`. Files are created automatically on first save.

## Commands

- `rolo` - Launch interactive UI
- `rolo populate` - Fetch and save all active tmux sessions
- `rolo next` - Switch to next session in order
- `rolo prev` - Switch to previous session in order
- `rolo help` - Show help message
