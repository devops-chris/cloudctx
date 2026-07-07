package cmd

import (
	"fmt"
	"strings"

	"atomicgo.dev/cursor"
	"atomicgo.dev/keyboard"
	"atomicgo.dev/keyboard/keys"
	"github.com/charmbracelet/lipgloss"
	"github.com/devops-chris/clihq/ui"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

// pickItem is one row in the interactive picker. display is the (possibly
// styled) string shown to the user; search is the plain text the filter runs
// against; value is what the picker returns when the row is chosen.
type pickItem struct {
	display string
	search  string
	value   string
}

var (
	pickerSearchStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#9B9B9B", Dark: "#5C5C5C"})
	pickerCursorStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}).Bold(true)
	pickerCountStyle  = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#9B9B9B", Dark: "#5C5C5C"})
)

// matchesQuery reports whether target matches query using fzf-style semantics:
// the query is split on whitespace and every token must be a (case-insensitive)
// fuzzy subsequence of target, in any order.
func matchesQuery(query, target string) bool {
	for _, token := range strings.Fields(query) {
		if !fuzzy.MatchFold(token, target) {
			return false
		}
	}
	return true
}

// filterItems returns the indices of items matching query, in original order.
func filterItems(items []pickItem, query string) []int {
	matched := make([]int, 0, len(items))
	for i, it := range items {
		if matchesQuery(query, it.search) {
			matched = append(matched, i)
		}
	}
	return matched
}

// windowBounds returns the [start, end) slice of a list of length total that
// keeps cursor visible within at most height rows.
func windowBounds(total, height, cursor int) (start, end int) {
	if height <= 0 || height > total {
		height = total
	}
	start = cursor - height + 1
	if start < 0 {
		start = 0
	}
	end = start + height
	if end > total {
		end = total
		start = end - height
		if start < 0 {
			start = 0
		}
	}
	return start, end
}

// liveArea manages an in-place updating block of terminal output.
type liveArea struct {
	lines int
}

func (a *liveArea) update(content string) {
	if a.lines > 0 {
		fmt.Printf("\033[%dA", a.lines)
	}
	fmt.Print("\033[J")
	fmt.Print(content)
	a.lines = strings.Count(content, "\n")
}

// runPicker shows an interactive, filterable list. It returns the chosen item's
// value and ok=true on Enter, or ok=false if the user cancels with Esc/Ctrl+C.
// maxHeight caps the number of visible rows.
func runPicker(items []pickItem, maxHeight int) (value string, ok bool) {
	query := ""
	matched := filterItems(items, query)
	cur := 0 // index into matched

	render := func() string {
		var b strings.Builder
		b.WriteString(pickerSearchStyle.Render("Search: ") + ui.Highlight(query) + "\n")
		if len(matched) == 0 {
			b.WriteString(pickerSearchStyle.Render("  (no matches)") + "\n")
			return b.String()
		}
		start, end := windowBounds(len(matched), maxHeight, cur)
		for i := start; i < end; i++ {
			it := items[matched[i]]
			if i == cur {
				b.WriteString(pickerCursorStyle.Render("❯ ") + it.display + "\n")
			} else {
				b.WriteString("  " + it.display + "\n")
			}
		}
		if len(matched) > (end - start) {
			b.WriteString(pickerCountStyle.Render(fmt.Sprintf("  %d/%d", cur+1, len(matched))) + "\n")
		}
		return b.String()
	}

	clamp := func() {
		if cur < 0 {
			cur = 0
		}
		if cur > len(matched)-1 {
			cur = len(matched) - 1
		}
		if cur < 0 {
			cur = 0
		}
	}

	initial := render()
	fmt.Print(initial)
	area := &liveArea{lines: strings.Count(initial, "\n")}

	cursor.Hide()
	defer cursor.Show()

	canceled := false
	_ = keyboard.Listen(func(key keys.Key) (stop bool, err error) {
		switch key.Code {
		case keys.RuneKey:
			query += string(key.Runes)
			matched = filterItems(items, query)
			cur = 0
		case keys.Space:
			query += " "
			matched = filterItems(items, query)
			cur = 0
		case keys.Backspace, keys.CtrlH:
			if r := []rune(query); len(r) > 0 {
				query = string(r[:len(r)-1])
				matched = filterItems(items, query)
				cur = 0
			}
		case keys.Up, keys.CtrlP, keys.CtrlK:
			cur--
		case keys.Down, keys.CtrlN, keys.CtrlJ:
			cur++
		case keys.Enter:
			if len(matched) > 0 {
				value = items[matched[cur]].value
				ok = true
			}
			return true, nil
		case keys.Escape, keys.CtrlC:
			canceled = true
			return true, nil
		}
		clamp()
		area.update(render())
		return false, nil
	})

	if canceled {
		return "", false
	}
	return value, ok
}
