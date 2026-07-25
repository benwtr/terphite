package tui

import "fmt"

// renderStatus renders the status line, truncating the error message (if
// any) so the whole line fits within width — an untruncated long error
// would wrap inside its box and grow it past its declared height, pushing
// everything below it off screen.
func renderStatus(timeFrom string, autorefreshSeconds, maxDataPoints int, errMsg string, width int) string {
	auto := "off"
	if autorefreshSeconds > 0 {
		auto = fmt.Sprintf("%ds", autorefreshSeconds)
	}
	maxdp := "unlimited"
	if maxDataPoints > 0 {
		maxdp = fmt.Sprintf("%d", maxDataPoints)
	}
	prefix := fmt.Sprintf("from: %s  autorefresh: %s  maxdatapoints: %s", timeFrom, auto, maxdp)
	if errMsg == "" {
		return truncateLine(prefix, width)
	}

	errPart := "   error: " + errMsg
	avail := width - len(prefix)
	if width > 0 && len(errPart) > avail {
		if avail > 1 {
			errPart = errPart[:avail-1] + "…"
		} else {
			errPart = ""
		}
	}
	return prefix + errorStyle.Render(errPart)
}

func truncateLine(s string, width int) string {
	if width > 0 && len(s) > width {
		return s[:width]
	}
	return s
}
