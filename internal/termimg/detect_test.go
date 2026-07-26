package termimg

import "testing"

func clearDetectEnv(t *testing.T) {
	t.Helper()
	t.Setenv("TERM_PROGRAM", "")
	t.Setenv("LC_TERMINAL", "")
	t.Setenv("KITTY_WINDOW_ID", "")
	t.Setenv("TERM", "xterm-256color")
}

func TestDetectNone(t *testing.T) {
	clearDetectEnv(t)
	if got := Detect(); got != ProtocolNone {
		t.Errorf("Detect() = %v, want ProtocolNone", got)
	}
}

func TestDetectITerm2(t *testing.T) {
	clearDetectEnv(t)
	t.Setenv("TERM_PROGRAM", "iTerm.app")
	if got := Detect(); got != ProtocolITerm2 {
		t.Errorf("Detect() = %v, want ProtocolITerm2", got)
	}
}

func TestDetectWezTermUsesITerm2Protocol(t *testing.T) {
	clearDetectEnv(t)
	t.Setenv("TERM_PROGRAM", "WezTerm")
	if got := Detect(); got != ProtocolITerm2 {
		t.Errorf("Detect() = %v, want ProtocolITerm2 (WezTerm supports it)", got)
	}
}

func TestDetectITerm2ViaLCTerminal(t *testing.T) {
	clearDetectEnv(t)
	t.Setenv("LC_TERMINAL", "iTerm2")
	if got := Detect(); got != ProtocolITerm2 {
		t.Errorf("Detect() = %v, want ProtocolITerm2", got)
	}
}

func TestDetectKittyViaWindowID(t *testing.T) {
	clearDetectEnv(t)
	t.Setenv("KITTY_WINDOW_ID", "1")
	if got := Detect(); got != ProtocolKitty {
		t.Errorf("Detect() = %v, want ProtocolKitty", got)
	}
}

func TestDetectKittyViaTerm(t *testing.T) {
	clearDetectEnv(t)
	t.Setenv("TERM", "xterm-kitty")
	if got := Detect(); got != ProtocolKitty {
		t.Errorf("Detect() = %v, want ProtocolKitty", got)
	}
}

func TestDetectPrefersKittyWhenBothMatch(t *testing.T) {
	clearDetectEnv(t)
	t.Setenv("TERM_PROGRAM", "WezTerm")
	t.Setenv("KITTY_WINDOW_ID", "1")
	if got := Detect(); got != ProtocolKitty {
		t.Errorf("Detect() = %v, want ProtocolKitty (preferred when both match)", got)
	}
}

func TestProtocolNextCycles(t *testing.T) {
	seq := []Protocol{ProtocolNone, ProtocolITerm2, ProtocolKitty, ProtocolNone}
	for i := 0; i < len(seq)-1; i++ {
		if got := seq[i].Next(); got != seq[i+1] {
			t.Errorf("%v.Next() = %v, want %v", seq[i], got, seq[i+1])
		}
	}
}

func TestProtocolString(t *testing.T) {
	cases := map[Protocol]string{
		ProtocolNone:   "off",
		ProtocolITerm2: "iterm2",
		ProtocolKitty:  "kitty",
	}
	for p, want := range cases {
		if got := p.String(); got != want {
			t.Errorf("%v.String() = %q, want %q", p, got, want)
		}
	}
}
