package tui

import (
	"testing"

	"github.com/benwtr/terphite/internal/termimg"
)

func TestImageEscapeCacheHitsWhenNothingChanged(t *testing.T) {
	var c imageEscapeCache
	png := []byte("fake-png")

	first := c.escape(termimg.ProtocolITerm2, png, 1, 40, 20)
	if first == "" {
		t.Fatal("expected non-empty escape sequence")
	}
	second := c.escape(termimg.ProtocolITerm2, png, 1, 40, 20)
	if second != first {
		t.Errorf("expected cache hit to return identical string, got different output")
	}
}

func TestImageEscapeCacheRebuildsOnVersionChange(t *testing.T) {
	var c imageEscapeCache
	first := c.escape(termimg.ProtocolITerm2, []byte("png-v1"), 1, 40, 20)
	second := c.escape(termimg.ProtocolITerm2, []byte("png-v2"), 2, 40, 20)
	if first == second {
		t.Error("expected different output after version bump with new bytes")
	}
}

func TestImageEscapeCacheRebuildsOnSizeChange(t *testing.T) {
	var c imageEscapeCache
	png := []byte("fake-png")
	first := c.escape(termimg.ProtocolITerm2, png, 1, 40, 20)
	second := c.escape(termimg.ProtocolITerm2, png, 1, 80, 40)
	if first == second {
		t.Error("expected different output after display size changed")
	}
}

func TestBuildImageEscapeNoneProtocol(t *testing.T) {
	if got := buildImageEscape(termimg.ProtocolNone, []byte("png"), 40, 20); got != "" {
		t.Errorf("expected empty string for ProtocolNone, got %q", got)
	}
}

func TestBuildImageEscapeEmptyPNG(t *testing.T) {
	if got := buildImageEscape(termimg.ProtocolITerm2, nil, 40, 20); got != "" {
		t.Errorf("expected empty string for empty png, got %q", got)
	}
}

func TestBuildImageEscapeKitty(t *testing.T) {
	got := buildImageEscape(termimg.ProtocolKitty, []byte("png"), 40, 20)
	if got == "" {
		t.Error("expected non-empty escape sequence for Kitty protocol")
	}
}
