package ui

import (
	"github.com/charmbracelet/lipgloss"
	lgtable "github.com/charmbracelet/lipgloss/table"
)

// Table renders a styled table. Headers are bold purple; data cells are unstyled
// so pre-colored strings (ui.Green, ui.Cyan, etc.) pass through unchanged.
func Table(headers []string, rows [][]string) string {
	t := lgtable.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(TableBorderStyle).
		Headers(headers...).
		Rows(rows...).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == lgtable.HeaderRow {
				return TableHeaderStyle
			}
			return lipgloss.NewStyle()
		})
	return t.Render()
}
