// Package timerange parses and formats Graphite-style relative time strings,
// e.g. "-1d12h", used as the "from" parameter of a render request.
package timerange

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Unit sizes in seconds. Months and years are fixed-size approximations
// (30d and 365d respectively), matching Graphite's own relative-time semantics.
const (
	UnitSecond = 1
	UnitMinute = 60 * UnitSecond
	UnitHour   = 60 * UnitMinute
	UnitDay    = 24 * UnitHour
	UnitWeek   = 7 * UnitDay
	UnitMonth  = 30 * UnitDay
	UnitYear   = 365 * UnitDay
)

// Default is the time range used when no other value has been set.
const Default = "-1min"

// Components is a parsed relative time string broken into its constituent units.
type Components struct {
	Years, Months, Weeks, Days, Hours, Minutes, Seconds int
}

// TotalSeconds returns the total number of seconds represented by c.
func (c Components) TotalSeconds() int {
	return c.Years*UnitYear + c.Months*UnitMonth + c.Weeks*UnitWeek +
		c.Days*UnitDay + c.Hours*UnitHour + c.Minutes*UnitMinute + c.Seconds*UnitSecond
}

var pattern = regexp.MustCompile(`^-?(?:(\d+)y)?(?:(\d+)mon)?(?:(\d+)w)?(?:(\d+)d)?(?:(\d+)h)?(?:(\d+)min)?(?:(\d+)s)?$`)

// Parse parses a relative time string such as "-1y2mon3w4d5h6min7s" into its
// components. Any unit may be omitted; an empty/all-zero match is valid and
// parses to a zero Components.
func Parse(s string) (Components, error) {
	m := pattern.FindStringSubmatch(s)
	if m == nil {
		return Components{}, fmt.Errorf("timerange: invalid time range %q", s)
	}
	vals := make([]int, 7)
	for i := 1; i <= 7; i++ {
		if m[i] == "" {
			continue
		}
		n, err := strconv.Atoi(m[i])
		if err != nil {
			return Components{}, fmt.Errorf("timerange: invalid time range %q: %w", s, err)
		}
		vals[i-1] = n
	}
	return Components{
		Years:   vals[0],
		Months:  vals[1],
		Weeks:   vals[2],
		Days:    vals[3],
		Hours:   vals[4],
		Minutes: vals[5],
		Seconds: vals[6],
	}, nil
}

// ParseSeconds is a convenience wrapper around Parse that returns the total
// number of seconds a relative time string represents.
func ParseSeconds(s string) (int, error) {
	c, err := Parse(s)
	if err != nil {
		return 0, err
	}
	return c.TotalSeconds(), nil
}

// Format renders a total number of seconds back into a canonical relative
// time string, greedily decomposing into the largest units first (e.g. 5400
// seconds formats as "-1h30min"). A non-positive input formats as Default.
func Format(totalSeconds int) string {
	if totalSeconds <= 0 {
		return Default
	}

	units := []struct {
		suffix string
		size   int
	}{
		{"y", UnitYear},
		{"mon", UnitMonth},
		{"w", UnitWeek},
		{"d", UnitDay},
		{"h", UnitHour},
		{"min", UnitMinute},
		{"s", UnitSecond},
	}

	var sb strings.Builder
	sb.WriteByte('-')
	remaining := totalSeconds
	wrote := false
	for _, u := range units {
		if remaining < u.size {
			continue
		}
		n := remaining / u.size
		remaining -= n * u.size
		fmt.Fprintf(&sb, "%d%s", n, u.suffix)
		wrote = true
	}
	if !wrote {
		return Default
	}
	return sb.String()
}

// Normalize parses s and reformats it canonically via Format, collapsing
// e.g. "-90min" to "-1h30min". Returns an error if s is not a valid relative
// time string.
func Normalize(s string) (string, error) {
	seconds, err := ParseSeconds(s)
	if err != nil {
		return "", err
	}
	return Format(seconds), nil
}
