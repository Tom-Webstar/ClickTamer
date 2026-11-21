// Package quartz resolves ClickTamer key tokens to macOS virtual key codes
// and posts keyboard events on macOS.
package quartz

import (
	"fmt"
	"strconv"
	"strings"
)

// These codes come from HIToolbox/Events.h. ANSI codes identify physical
// US keyboard positions. The active keyboard layout determines the text.
var ansiLetters = [...]uint16{
	0x00, 0x0B, 0x08, 0x02, 0x0E, 0x03, 0x05, // a–g
	0x04, 0x22, 0x26, 0x28, 0x25, 0x2E, 0x2D, // h–n
	0x1F, 0x23, 0x0C, 0x0F, 0x01, 0x11, 0x20, // o–u
	0x09, 0x0D, 0x07, 0x10, 0x06, // v–z
}

var ansiDigits = [...]uint16{0x1D, 0x12, 0x13, 0x14, 0x15, 0x17, 0x16, 0x1A, 0x1C, 0x19}
var functionKeys = [...]uint16{0x7A, 0x78, 0x63, 0x76, 0x60, 0x61, 0x62, 0x64, 0x65, 0x6D, 0x67, 0x6F}

var namedKeys = map[string]uint16{
	"up":        0x7E,
	"down":      0x7D,
	"left":      0x7B,
	"right":     0x7C,
	"enter":     0x24,
	"space":     0x31,
	"tab":       0x30,
	"esc":       0x35,
	"escape":    0x35,
	"backspace": 0x33, // kVK_Delete
	"delete":    0x75, // kVK_ForwardDelete
	"insert":    0x72, // kVK_Help, used for Insert
	"home":      0x73,
	"end":       0x77,
	"pageup":    0x74,
	"pagedown":  0x79,
	"shift":     0x38,
	"ctrl":      0x3B,
	"alt":       0x3A, // Option
	"capslock":  0x39,
	"-":         0x1B,
	"=":         0x18,
	",":         0x2B,
	".":         0x2F,
	"/":         0x2C,
	";":         0x29,
	"'":         0x27,
	"[":         0x21,
	"]":         0x1E,
	"`":         0x32,
	"\\":        0x2A,
}

// Resolve returns the macOS virtual key code for a config token. KeyDown and
// KeyUp accept this value directly.
func Resolve(token string) (uint16, error) {
	key := strings.ToLower(strings.TrimSpace(token))
	if code, ok := namedKeys[key]; ok {
		return code, nil
	}
	if len(key) >= 2 && len(key) <= 3 && key[0] == 'f' {
		if n, err := strconv.Atoi(key[1:]); err == nil && n >= 1 && n <= len(functionKeys) {
			return functionKeys[n-1], nil
		}
	}
	if len(key) == 1 {
		code := key[0]
		switch {
		case code >= 'a' && code <= 'z':
			return ansiLetters[code-'a'], nil
		case code >= '0' && code <= '9':
			return ansiDigits[code-'0'], nil
		}
	}
	return 0, fmt.Errorf("quartz: unknown key %q", token)
}
