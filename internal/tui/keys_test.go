package tui

import (
	"strings"
	"testing"

	"github.com/benwtr/terphite/internal/termimg"
)

func composerHelp(width, maxRows int, collapsed bool) string {
	return helpText(viewComposer, popupNone, drawLine, termimg.ProtocolNone, width, maxRows, collapsed)
}

func helpRows(s string) int { return strings.Count(s, "\n") + 1 }

func TestHelpTextCollapsedIsOneLine(t *testing.T) {
	got := composerHelp(200, 6, true)
	if strings.Contains(got, "\n") {
		t.Errorf("collapsed help should be a single line, got:\n%s", got)
	}
}

func TestHelpTextUsesMultipleColumns(t *testing.T) {
	got := composerHelp(200, 6, false)
	if rows := helpRows(got); rows >= len(composerKeys) {
		t.Errorf("expected multi-column layout to use fewer than %d rows, got %d:\n%s",
			len(composerKeys), rows, got)
	}
}

// The row budget is what keeps the help box from overflowing the frame, so
// it must hold even at widths too narrow to lay the bindings out nicely.
func TestHelpTextRespectsRowBudget(t *testing.T) {
	for _, width := range []int{20, 40, 80, 120, 200} {
		for _, maxRows := range []int{1, 3, 6, 12} {
			got := composerHelp(width, maxRows, false)
			if rows := helpRows(got); rows > maxRows {
				t.Errorf("width %d budget %d: got %d rows:\n%s", width, maxRows, rows, got)
			}
		}
	}
}

func TestHelpTextNeverExceedsWidth(t *testing.T) {
	for _, width := range []int{20, 40, 80, 120, 200} {
		for _, maxRows := range []int{1, 3, 6, 12} {
			got := composerHelp(width, maxRows, false)
			for _, line := range strings.Split(got, "\n") {
				if n := len([]rune(line)); n > width {
					t.Errorf("width %d budget %d: line of %d runes overflows: %q",
						width, maxRows, n, line)
				}
			}
		}
	}
}

func TestHelpTextIncludesEveryBindingKey(t *testing.T) {
	got := composerHelp(200, 6, false)
	for _, b := range composerKeys {
		if !strings.Contains(got, b.Key) {
			t.Errorf("help is missing binding %q:\n%s", b.Key, got)
		}
	}
}

func TestHelpTextAnnotatesToggleState(t *testing.T) {
	got := helpText(viewComposer, popupNone, drawStacked, termimg.ProtocolKitty, 200, 6, false)
	if !strings.Contains(got, "[stacked]") {
		t.Errorf("expected graph style annotation, got:\n%s", got)
	}
	if !strings.Contains(got, "[kitty]") {
		t.Errorf("expected image protocol annotation, got:\n%s", got)
	}
}
