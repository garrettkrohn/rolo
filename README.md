# Rolo

A terminal UI for organizing and navigating tmux sessions with vim-like keybindings. Organize your sessions into groups for better project management.

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

2. Launch the UI to manage sessions and groups:
   ```bash
   rolo
   ```

3. Navigate between sessions:
   ```bash
   rolo next  # Switch to next session in current group
   rolo prev  # Switch to previous session in current group
   ```

4. Navigate between groups:
   ```bash
   rolo left   # Switch to previous group
   rolo right  # Switch to next group
   ```

## Groups

Rolo lets you organize your tmux sessions into groups, perfect for managing multiple projects or tickets.

### Example Workflow

```bash
# You have a "main" group by default
rolo populate  # Adds all sessions to current group

# Sessions are organized up/down (j/k in TUI, next/prev in CLI)
rolo next      # Move down through sessions
rolo prev      # Move up through sessions

# Groups are organized left/right (h/l in TUI, left/right in CLI)
rolo left      # Switch to previous group
rolo right     # Switch to next group
```

### Group Navigation in TUI

When you run `rolo`, the TUI displays your current group and adjacent groups:

```
← previous_group | CURRENT_GROUP | next_group →
```

Use **h/l** to switch between groups, and the session list will update automatically.

## Tmux Integration

Add to `~/.tmux.conf`:

```tmux
# Open rolo UI
bind-key r run-shell "tmux popup -E -w 40% -h 60% 'rolo'"

# Navigate sessions (up/down within current group)
bind-key j run-shell "rolo next"
bind-key k run-shell "rolo prev"

# Navigate groups (left/right between groups)
bind-key h run-shell "rolo left"
bind-key l run-shell "rolo right"
```

## Keybindings

### Normal Mode

#### Session Navigation (Up/Down)
- `j/k` - Navigate up/down through sessions
- `o` - Attach to session
- `d` - Toggle delete/hide session
- `D` - Kill tmux session (permanent deletion)
- `t` - Toggle visibility of deleted sessions
- `u` - Update list (sync with active sessions)
- `p` - Repopulate list (replace with active sessions)
- `m` - Enter move mode

#### Group Navigation (Left/Right)
- `h/l` - Switch to previous/next group
- `H/L` - Move current session to previous/next group
- Sessions list updates automatically when switching groups

#### Group Management
- `n` - Create new group
- `r` - Rename current group
- `X` - Delete current group (requires confirmation, cannot delete last group)

#### Other
- `?` - Toggle help display
- `s/Enter` - Save and quit
- `q/Ctrl+C` - Quit without saving
- `Esc` - Exit move mode (returns to normal mode)

### Move Mode
- `j/k` - Move item up/down within current group
- `m` - Return to normal mode
- `s/Enter` - Save and quit

## Configuration

Settings are stored in `~/.config/rolo/config.json`:

```json
{
  "wrap_around": false,
  "show_deleted": true
}
```

- `wrap_around`: Enable wrap-around navigation (default: `false`)
  - For sessions: wraps from bottom to top (or vice versa)
  - For groups: wraps from rightmost to leftmost (or vice versa)
- `show_deleted`: Show deleted/hidden sessions in the TUI (default: `true`)
  - Can be toggled with `t` key in TUI
  - Preference is automatically saved when toggled

### Data Storage

The application uses **only** `groups.json` for all operations:

- **Groups and sessions**: `~/.config/rolo/groups.json` ✅ (active)
- **Settings**: `~/.config/rolo/config.json`

Files are created automatically on first save.

### Migration

If you're upgrading from an older version of Rolo:

1. On first run, `rolo.json` (old format) is automatically detected
2. All sessions are migrated to `groups.json` in a "main" group
3. **`rolo.json` is preserved** as a backup but never used again
4. You can safely delete `rolo.json` after verifying the migration worked

**After migration**: The app only reads/writes `groups.json`. The old `rolo.json` is never touched again.

## Commands

### Interactive
- `rolo` - Launch interactive UI with full group and session management

### Session Navigation
- `rolo next` - Switch to next session in current group
- `rolo prev` - Switch to previous session in current group

### Group Navigation
- `rolo left` - Switch to previous group
- `rolo right` - Switch to next group

### Management
- `rolo populate` - Fetch and save all active tmux sessions to current group
- `rolo context` - List working directories of sessions in current group
- `rolo help` - Show help message

## Advanced Usage

### Multiple Groups Example

```bash
# Scenario: You're working on multiple tickets

# Create and populate first group (automatically named "main")
rolo populate  # Adds all current sessions

# Switch to work on a different ticket
# (In the TUI, you could create a new group, but for now groups are managed via the data file)

# Navigate between your groups
rolo left   # Go to previous group
rolo right  # Go to next group

# Each group maintains its own session order
rolo next   # Next session in current group
rolo prev   # Previous session in current group
```

### Behavior Notes

- **Single session groups**: If a group has only one (non-deleted) session, `rolo next` and `rolo prev` will do nothing (no switching)
- **Persistent current group**: The current group is **automatically saved** when you switch groups in the TUI (using `h`/`l`). This ensures CLI commands like `rolo next/prev` always operate on the correct group, even if you exit the TUI without explicitly saving
- **Auto-save on group switch**: When you press `h` or `l` to switch groups in the TUI, the current group is immediately saved to `groups.json`
- **Wrap-around**: Controlled by the `wrap_around` setting in `config.json` - applies to both session and group navigation

#### Example Workflow

```bash
# Start with 2 groups: "main" (with sessions rolo, rx) and "rolo_group" (with only rolo)
rolo  # Open TUI

# Press 'l' to switch to "rolo_group"
# -> Current group is AUTO-SAVED immediately

# Press 'q' to quit (without explicit save)

# Now run CLI command
rolo next
# -> Does nothing! Because current group is "rolo_group" with only 1 session
# -> Without auto-save, it would incorrectly use "main" and switch to "rx"
```

## Claude Code Integration

You can use `rolo context` to add all session working directories from your current group to Claude Code's context:

```bash
# List working directories for current group
rolo context

# Example: Add all directories to Claude Code context
# (This assumes you have a way to pass directories to Claude)
rolo context | xargs -I {} echo "Adding: {}"
```

This is useful when you want Claude to have context about all the projects/directories you're working on in a particular group.

### Data Format

The `groups.json` file structure:

```json
{
  "groups": [
    {
      "name": "main",
      "sessions": [
        {"name": "session1", "deleted": false},
        {"name": "session2", "deleted": false}
      ]
    },
    {
      "name": "APICAL-123",
      "sessions": [
        {"name": "backend", "deleted": false},
        {"name": "frontend", "deleted": false}
      ]
    }
  ],
  "current_group": 0
}
```

You can manually edit this file to create groups or rename them (while rolo is not running).
