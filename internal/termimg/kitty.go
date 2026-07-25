package termimg

import (
	"encoding/base64"
	"strconv"
	"strings"
)

// kittyChunkSize is the maximum base64 payload bytes the Kitty graphics
// protocol allows per escape command; larger images must be split across
// multiple chunks.
const kittyChunkSize = 4096

// KittyEscape builds the Kitty graphics protocol escape sequence (also
// supported by WezTerm and Ghostty) to transmit and display png sized to
// cols x rows terminal cells, chunking the base64 payload as the protocol
// requires.
func KittyEscape(png []byte, cols, rows int) string {
	encoded := base64.StdEncoding.EncodeToString(png)

	var b strings.Builder
	for len(encoded) > 0 {
		chunk := encoded
		if len(chunk) > kittyChunkSize {
			chunk = encoded[:kittyChunkSize]
		}
		encoded = encoded[len(chunk):]

		more := 0
		if len(encoded) > 0 {
			more = 1
		}

		if b.Len() == 0 {
			b.WriteString("\x1b_Ga=T,f=100,c=")
			b.WriteString(strconv.Itoa(cols))
			b.WriteString(",r=")
			b.WriteString(strconv.Itoa(rows))
			b.WriteString(",m=")
			b.WriteString(strconv.Itoa(more))
			b.WriteByte(';')
		} else {
			b.WriteString("\x1b_Gm=")
			b.WriteString(strconv.Itoa(more))
			b.WriteByte(';')
		}
		b.WriteString(chunk)
		b.WriteString("\x1b\\")
	}
	return b.String()
}
