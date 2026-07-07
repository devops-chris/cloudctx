package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	colorPurple = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	colorGreen  = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}
	colorRed    = lipgloss.AdaptiveColor{Light: "#D0342C", Dark: "#FF4672"}
	colorYellow = lipgloss.AdaptiveColor{Light: "#A07A10", Dark: "#FFCC66"}
	colorMuted  = lipgloss.AdaptiveColor{Light: "#9B9B9B", Dark: "#5C5C5C"}

	successStyle  = lipgloss.NewStyle().Foreground(colorGreen).Bold(true)
	errorStyle    = lipgloss.NewStyle().Foreground(colorRed).Bold(true)
	warningStyle  = lipgloss.NewStyle().Foreground(colorYellow).Bold(true)
	infoStyle     = lipgloss.NewStyle().Foreground(colorPurple).Bold(true)
	subtleStyle   = lipgloss.NewStyle().Foreground(colorMuted)
	highlightStyle = lipgloss.NewStyle().Foreground(colorPurple).Bold(true)
	greenStyle    = lipgloss.NewStyle().Foreground(colorGreen)
	yellowStyle   = lipgloss.NewStyle().Foreground(colorYellow)
	redStyle      = lipgloss.NewStyle().Foreground(colorRed)
	cyanStyle     = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#0097A7", Dark: "#4DD0E1"})

	bannerStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPurple).
			Padding(1, 3).
			Width(52)

	bannerTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorPurple)

	bannerSubtitleStyle = lipgloss.NewStyle().
				Faint(true)

	sectionStyle = lipgloss.NewStyle().Bold(true).Foreground(colorPurple)

	// TableHeaderStyle is exported for use in table renderers.
	TableHeaderStyle = lipgloss.NewStyle().Bold(true).Foreground(colorPurple)
	// TableBorderStyle is exported for use in table renderers.
	TableBorderStyle = lipgloss.NewStyle().Foreground(colorMuted)
)

func Banner() string {
	content := lipgloss.JoinVertical(lipgloss.Left,
		bannerTitleStyle.Render("cloudctx"),
		bannerSubtitleStyle.Render("cloud context switcher"),
	)
	return bannerStyle.Render(content)
}

func SectionHeader(title string) string { return sectionStyle.Render(title) }

func Info(msg string) string      { return infoStyle.Render(msg) }
func Infof(f string, a ...any) string { return Info(fmt.Sprintf(f, a...)) }

func Success(msg string) string      { return successStyle.Render(msg) }
func Successf(f string, a ...any) string { return Success(fmt.Sprintf(f, a...)) }

func Error(msg string) string      { return errorStyle.Render(msg) }
func Errorf(f string, a ...any) string { return Error(fmt.Sprintf(f, a...)) }

func Warning(msg string) string      { return warningStyle.Render(msg) }
func Warningf(f string, a ...any) string { return Warning(fmt.Sprintf(f, a...)) }

func Subtle(msg string) string      { return subtleStyle.Render(msg) }
func Subtlef(f string, a ...any) string { return Subtle(fmt.Sprintf(f, a...)) }

func Highlight(msg string) string { return highlightStyle.Render(msg) }
func Green(msg string) string     { return greenStyle.Render(msg) }
func Yellow(msg string) string    { return yellowStyle.Render(msg) }
func Red(msg string) string       { return redStyle.Render(msg) }
func Cyan(msg string) string      { return cyanStyle.Render(msg) }
