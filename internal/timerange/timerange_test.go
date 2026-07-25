package timerange

import "testing"

func TestParse(t *testing.T) {
	cases := []struct {
		in   string
		want Components
	}{
		{"-1min", Components{Minutes: 1}},
		{"-1d12h", Components{Days: 1, Hours: 12}},
		{"-1y2mon3w4d5h6min7s", Components{Years: 1, Months: 2, Weeks: 3, Days: 4, Hours: 5, Minutes: 6, Seconds: 7}},
		{"-90min", Components{Minutes: 90}},
		{"", Components{}},
		{"-", Components{}},
	}
	for _, c := range cases {
		got, err := Parse(c.in)
		if err != nil {
			t.Errorf("Parse(%q) unexpected error: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("Parse(%q) = %+v, want %+v", c.in, got, c.want)
		}
	}
}

func TestParseInvalid(t *testing.T) {
	for _, in := range []string{"bogus", "-1decade", "-1d1"} {
		if _, err := Parse(in); err == nil {
			t.Errorf("Parse(%q) expected error, got nil", in)
		}
	}
}

func TestTotalSeconds(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"-1min", UnitMinute},
		{"-1h", UnitHour},
		{"-1d12h", UnitDay + 12*UnitHour},
		{"-1w", UnitWeek},
		{"-1mon", UnitMonth},
		{"-1y", UnitYear},
	}
	for _, c := range cases {
		got, err := ParseSeconds(c.in)
		if err != nil {
			t.Fatalf("ParseSeconds(%q) unexpected error: %v", c.in, err)
		}
		if got != c.want {
			t.Errorf("ParseSeconds(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestFormat(t *testing.T) {
	cases := []struct {
		in   int
		want string
	}{
		{0, "-1min"},
		{-5, "-1min"},
		{UnitMinute, "-1min"},
		{90 * UnitMinute, "-1h30min"},
		{UnitDay + 12*UnitHour, "-1d12h"},
		{UnitYear + 2*UnitMonth + 3*UnitWeek + 4*UnitDay + 5*UnitHour + 6*UnitMinute + 7*UnitSecond, "-1y2mon3w4d5h6min7s"},
	}
	for _, c := range cases {
		got := Format(c.in)
		if got != c.want {
			t.Errorf("Format(%d) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestNormalizeRoundTrip(t *testing.T) {
	cases := []struct{ in, want string }{
		{"-90min", "-1h30min"},
		{"-1d12h", "-1d12h"},
		{"-1min", "-1min"},
	}
	for _, c := range cases {
		got, err := Normalize(c.in)
		if err != nil {
			t.Fatalf("Normalize(%q) unexpected error: %v", c.in, err)
		}
		if got != c.want {
			t.Errorf("Normalize(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
