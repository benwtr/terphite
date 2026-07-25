// Package termimg detects terminal support for inline image protocols and
// builds the escape sequences that display an image using them.
package termimg

import "os"

// Protocol identifies which inline-image escape-sequence protocol (if any)
// the terminal supports.
type Protocol int

const (
	ProtocolNone Protocol = iota
	ProtocolITerm2
	ProtocolKitty
)

func (p Protocol) String() string {
	switch p {
	case ProtocolITerm2:
		return "iterm2"
	case ProtocolKitty:
		return "kitty"
	default:
		return "off"
	}
}

// Next cycles Off -> iTerm2 -> Kitty -> Off, for a manual override key.
func (p Protocol) Next() Protocol {
	return (p + 1) % 3
}

// Detect makes a best-effort guess at which inline-image protocol the
// current terminal supports, based on environment variables. It can guess
// wrong — tmux and many SSH setups don't relay these escape sequences to
// the outer terminal — so callers should offer a manual override rather
// than trusting this unconditionally.
func Detect() Protocol {
	if isKitty() {
		return ProtocolKitty
	}
	if isITerm2() {
		return ProtocolITerm2
	}
	return ProtocolNone
}

func isKitty() bool {
	if os.Getenv("KITTY_WINDOW_ID") != "" {
		return true
	}
	return os.Getenv("TERM") == "xterm-kitty"
}

func isITerm2() bool {
	switch os.Getenv("TERM_PROGRAM") {
	case "iTerm.app", "WezTerm":
		return true
	}
	switch os.Getenv("LC_TERMINAL") {
	case "iTerm2", "WezTerm":
		return true
	}
	return false
}
