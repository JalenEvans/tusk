//go:build tools

package tools

// Blank imports to keep required dependencies in go.mod.
import (
	_ "github.com/charmbracelet/bubbles"
	_ "github.com/charmbracelet/bubbletea"
	_ "github.com/charmbracelet/lipgloss"
	_ "golang.org/x/sys/unix"
)
