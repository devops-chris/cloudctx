package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	checkPassStyle = lipgloss.NewStyle().Bold(true).Foreground(colorGreen)
	checkFailStyle = lipgloss.NewStyle().Bold(true).Foreground(colorRed)
	checkWarnStyle = lipgloss.NewStyle().Bold(true).Foreground(colorYellow)
	checkHintStyle = lipgloss.NewStyle().Foreground(colorMuted).PaddingLeft(6)
)

func CheckPass(msg string) string {
	return fmt.Sprintf("  %s  %s", checkPassStyle.Render("✓"), msg)
}

func CheckFail(msg, hint string) string {
	s := fmt.Sprintf("  %s  %s", checkFailStyle.Render("✗"), msg)
	if hint != "" {
		s += "\n" + checkHintStyle.Render(hint)
	}
	return s
}

func CheckWarn(msg, hint string) string {
	s := fmt.Sprintf("  %s  %s", checkWarnStyle.Render("⚠"), msg)
	if hint != "" {
		s += "\n" + checkHintStyle.Render(hint)
	}
	return s
}

// CheckSection renders an indented section header for use inside check lists.
func CheckSection(name string) string {
	return fmt.Sprintf("\n  %s\n\n", sectionStyle.Render(name))
}
