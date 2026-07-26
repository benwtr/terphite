package termimg

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestKittyEscapeSingleChunk(t *testing.T) {
	png := []byte("small-png-bytes")
	seq := KittyEscape(png, 40, 20)

	if !strings.HasPrefix(seq, "\x1b_Ga=T,f=100,c=40,r=20,m=0;") {
		t.Fatalf("unexpected header: %q", seq)
	}
	if !strings.HasSuffix(seq, "\x1b\\") {
		t.Fatalf("missing ST terminator: %q", seq)
	}

	payload := strings.TrimPrefix(seq, "\x1b_Ga=T,f=100,c=40,r=20,m=0;")
	payload = strings.TrimSuffix(payload, "\x1b\\")
	decoded, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		t.Fatalf("payload isn't valid base64: %v", err)
	}
	if string(decoded) != string(png) {
		t.Errorf("decoded payload = %q, want %q", decoded, png)
	}
}

func TestKittyEscapeChunking(t *testing.T) {
	png := make([]byte, 10000)
	for i := range png {
		png[i] = byte(i % 256)
	}
	seq := KittyEscape(png, 10, 5)

	var chunks []string
	for _, c := range strings.Split(seq, "\x1b\\") {
		if c != "" {
			chunks = append(chunks, c)
		}
	}
	if len(chunks) < 2 {
		t.Fatalf("expected multiple chunks for a large payload, got %d", len(chunks))
	}

	if !strings.HasPrefix(chunks[0], "\x1b_Ga=T,f=100,c=10,r=5,m=1;") {
		t.Errorf("first chunk header wrong: %q", chunks[0])
	}
	if !strings.HasPrefix(chunks[len(chunks)-1], "\x1b_Gm=0;") {
		t.Errorf("last chunk should have m=0: %q", chunks[len(chunks)-1])
	}
	for _, c := range chunks[1 : len(chunks)-1] {
		if !strings.HasPrefix(c, "\x1b_Gm=1;") {
			t.Errorf("middle chunk should have m=1: %q", c)
		}
	}

	var b strings.Builder
	for _, c := range chunks {
		idx := strings.Index(c, ";")
		b.WriteString(c[idx+1:])
	}
	decoded, err := base64.StdEncoding.DecodeString(b.String())
	if err != nil {
		t.Fatalf("reassembled payload isn't valid base64: %v", err)
	}
	if string(decoded) != string(png) {
		t.Errorf("reassembled payload mismatch: got %d bytes, want %d", len(decoded), len(png))
	}
}
