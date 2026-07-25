package tui

import (
	"os/exec"
	"runtime"

	"github.com/atotto/clipboard"
)

// openInBrowser opens url in the user's default browser, cross-platform.
func openInBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}

// copyToClipboard copies text to the system clipboard, cross-platform
// (replacing the original's iTerm2-only escape-sequence approach).
func copyToClipboard(text string) error {
	return clipboard.WriteAll(text)
}
