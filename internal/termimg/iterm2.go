package termimg

import (
	"encoding/base64"
	"fmt"
)

// ITerm2Escape builds the iTerm2 inline-image escape sequence (also
// supported by WezTerm) to display png sized to cols x rows terminal
// cells. Size is given in cells, so the terminal handles scaling — no
// pixel-per-cell math is needed on our end.
func ITerm2Escape(png []byte, cols, rows int) string {
	encoded := base64.StdEncoding.EncodeToString(png)
	return fmt.Sprintf(
		"\x1b]1337;File=inline=1;width=%d;height=%d;preserveAspectRatio=0;size=%d:%s\a",
		cols, rows, len(png), encoded,
	)
}
