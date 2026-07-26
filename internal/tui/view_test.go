package tui

import (
	"strings"
	"testing"

	"github.com/benwtr/terphite/internal/dashboard"
	"github.com/benwtr/terphite/internal/termimg"
)

// frameRows counts the rows a rendered frame occupies. Overlay escapes are
// appended after the last line and draw out-of-band, so they add no rows.
func frameRows(view string) int {
	return strings.Count(view, "\n") + 1
}

func TestComposerFrameFitsTerminalHeight(t *testing.T) {
	sizes := [][2]int{{80, 24}, {170, 40}, {200, 50}, {120, 30}}
	for _, size := range sizes {
		w, h := size[0], size[1]
		for _, collapsed := range []bool{false, true} {
			m := newTestModel(t)
			m.width, m.height = w, h
			m.helpCollapsed = collapsed

			if got := frameRows(m.viewComposerScreen()); got > h {
				t.Errorf("%dx%d collapsed=%v: frame is %d rows, exceeds height %d",
					w, h, collapsed, got, h)
			}
		}
	}
}

// The image-mode frame must be the same height as the ASCII one: the image
// is painted as an overlay rather than embedded, so it must not add rows.
// Getting this wrong pushed the layout off-screen and smeared old frames.
func TestComposerFrameHeightUnchangedByImageMode(t *testing.T) {
	m := newTestModel(t)
	m.width, m.height = 170, 40

	ascii := frameRows(m.viewComposerScreen())

	m.imageProtocol = termimg.ProtocolITerm2
	m.imageBytes = []byte("fake-png-bytes")
	m.imageVersion = 1
	withImage := frameRows(m.viewComposerScreen())

	if withImage != ascii {
		t.Errorf("image mode changed frame height: %d rows vs %d in ascii mode", withImage, ascii)
	}
	if withImage > m.height {
		t.Errorf("image-mode frame is %d rows, exceeds height %d", withImage, m.height)
	}
}

func TestComposerImageOverlayEmittedOnlyInImageMode(t *testing.T) {
	m := newTestModel(t)
	m.width, m.height = 170, 40

	if strings.Contains(m.viewComposerScreen(), "\x1b7") {
		t.Error("ascii mode should not emit an image overlay")
	}

	m.imageProtocol = termimg.ProtocolITerm2
	m.imageBytes = []byte("fake-png-bytes")
	m.imageVersion = 1
	if !strings.Contains(m.viewComposerScreen(), "\x1b7") {
		t.Error("image mode should emit an image overlay")
	}
}

func TestDashboardFrameFitsTerminalHeight(t *testing.T) {
	panels := []dashboard.Panel{
		{Title: "a", Targets: []string{"stats.a"}, TimeFrom: "-1h"},
		{Title: "b", Targets: []string{"stats.b"}, TimeFrom: "-1h"},
		{Title: "c", Targets: []string{"stats.c"}, TimeFrom: "-1h"},
	}
	for _, size := range [][2]int{{80, 24}, {170, 40}, {200, 50}} {
		w, h := size[0], size[1]
		for _, proto := range []termimg.Protocol{termimg.ProtocolNone, termimg.ProtocolITerm2} {
			m := newTestModel(t)
			m.width, m.height = w, h
			m.currentDashboard = &dashboard.Dashboard{Name: "d", Panels: panels}
			m.panelImages = [][]byte{[]byte("png"), []byte("png"), []byte("png")}
			m.panelImageVersions = []int{1, 1, 1}
			m.panelImageCaches = make([]imageEscapeCache, len(panels))
			m.imageProtocol = proto

			if got := frameRows(m.viewDashboardScreen()); got > h {
				t.Errorf("%dx%d proto=%v: dashboard frame is %d rows, exceeds height %d",
					w, h, proto, got, h)
			}
		}
	}
}
