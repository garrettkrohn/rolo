package tmux

import (
	"fmt"
	"os/exec"
	"strings"
)

// GetActiveSessions returns a list of active tmux session names
func GetActiveSessions() ([]string, error) {
	// Run tmux list-sessions command
	cmd := exec.Command("tmux", "list-sessions", "-F", "#{session_name}")
	output, err := cmd.Output()
	if err != nil {
		// Check if it's because tmux isn't running
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("tmux command failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to run tmux: %w", err)
	}

	// Parse the output
	content := strings.TrimSpace(string(output))
	if content == "" {
		return []string{}, nil
	}

	sessions := strings.Split(content, "\n")
	
	// Filter out empty lines
	filtered := make([]string, 0, len(sessions))
	for _, session := range sessions {
		session = strings.TrimSpace(session)
		if session != "" {
			filtered = append(filtered, session)
		}
	}

	return filtered, nil
}

// GetCurrentSession returns the name of the current tmux session
// Returns an error if not inside a tmux session
func GetCurrentSession() (string, error) {
	cmd := exec.Command("tmux", "display-message", "-p", "#S")
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("not in a tmux session: %s", string(exitErr.Stderr))
		}
		return "", fmt.Errorf("failed to get current session: %w", err)
	}

	session := strings.TrimSpace(string(output))
	if session == "" {
		return "", fmt.Errorf("no current session found")
	}

	return session, nil
}

// SwitchToSession switches to the specified tmux session
func SwitchToSession(sessionName string) error {
	cmd := exec.Command("tmux", "switch-client", "-t", sessionName)
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("failed to switch to session '%s': %s", sessionName, string(exitErr.Stderr))
		}
		return fmt.Errorf("failed to switch to session '%s': %w", sessionName, err)
	}

	return nil
}

// KillSession kills the specified tmux session
func KillSession(sessionName string) error {
	cmd := exec.Command("tmux", "kill-session", "-t", sessionName)
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("failed to kill session '%s': %s", sessionName, string(exitErr.Stderr))
		}
		return fmt.Errorf("failed to kill session '%s': %w", sessionName, err)
	}

	return nil
}

// AttachToSession attaches to the specified tmux session
// If already in a tmux session, it switches to the target session
// If not in a tmux session, it attaches to the target session
func AttachToSession(sessionName string) error {
	// Check if we're already in a tmux session
	_, err := GetCurrentSession()
	if err == nil {
		// We're in a tmux session, so use switch-client
		return SwitchToSession(sessionName)
	}

	// Not in a tmux session, so use attach-session
	cmd := exec.Command("tmux", "attach-session", "-t", sessionName)
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("failed to attach to session '%s': %s", sessionName, string(exitErr.Stderr))
		}
		return fmt.Errorf("failed to attach to session '%s': %w", sessionName, err)
	}

	return nil
}

// GetSessionWorkingDirectory returns the working directory of the specified tmux session
// It gets the current path of the first pane in the session
func GetSessionWorkingDirectory(sessionName string) (string, error) {
	cmd := exec.Command("tmux", "display-message", "-p", "-t", sessionName, "-F", "#{pane_current_path}")
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("failed to get working directory for session '%s': %s", sessionName, string(exitErr.Stderr))
		}
		return "", fmt.Errorf("failed to get working directory for session '%s': %w", sessionName, err)
	}

	path := strings.TrimSpace(string(output))
	if path == "" {
		return "", fmt.Errorf("no working directory found for session '%s'", sessionName)
	}

	return path, nil
}
