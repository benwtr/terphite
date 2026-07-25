package termimg

import (
	"encoding/base64"
	"strconv"
	"strings"
	"testing"
)

func TestITerm2EscapeStructure(t *testing.T) {
	png := []byte("fake-png-bytes")
	seq := ITerm2Escape(png, 40, 20)

	if !strings.HasPrefix(seq, "\x1b]1337;File=") {
		t.Fatalf("missing OSC 1337 prefix: %q", seq)
	}
	if !strings.HasSuffix(seq, "\a") {
		t.Fatalf("missing BEL terminator: %q", seq)
	}
	if !strings.Contains(seq, "width="+strconv.Itoa(40)) {
		t.Errorf("missing width=40: %q", seq)
	}
	if !strings.Contains(seq, "height="+strconv.Itoa(20)) {
		t.Errorf("missing height=20: %q", seq)
	}
	if !strings.Contains(seq, "inline=1") {
		t.Errorf("missing inline=1: %q", seq)
	}

	idx := strings.LastIndex(seq, ":")
	payload := seq[idx+1 : len(seq)-1]
	decoded, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		t.Fatalf("payload isn't valid base64: %v", err)
	}
	if string(decoded) != string(png) {
		t.Errorf("decoded payload = %q, want %q", decoded, png)
	}
}
