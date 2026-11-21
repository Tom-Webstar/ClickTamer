//go:build linux

package keys

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/bendahl/uinput"
)

func TestLinuxKeyCodes(t *testing.T) {
	cases := map[string]int{
		"a": uinput.KeyA, "z": uinput.KeyZ, "q": uinput.KeyQ,
		"0": uinput.Key0, "9": uinput.Key9,
		"f1": uinput.KeyF1, "f10": uinput.KeyF10,
		"f11": uinput.KeyF11, "f12": uinput.KeyF12,
		"left": uinput.KeyLeft, "right": uinput.KeyRight,
		"up": uinput.KeyUp, "down": uinput.KeyDown,
		"insert": uinput.KeyInsert, "delete": uinput.KeyDelete,
		"home": uinput.KeyHome, "end": uinput.KeyEnd,
		"pageup": uinput.KeyPageup, "pagedown": uinput.KeyPagedown,
		"enter": uinput.KeyEnter, "space": uinput.KeySpace,
		"tab": uinput.KeyTab, "esc": uinput.KeyEsc,
		"backspace": uinput.KeyBackspace,
		"shift":     uinput.KeyLeftshift, "ctrl": uinput.KeyLeftctrl,
		"alt": uinput.KeyLeftalt, "capslock": uinput.KeyCapslock,
		"-": uinput.KeyMinus, "=": uinput.KeyEqual,
		",": uinput.KeyComma, ".": uinput.KeyDot,
		"/": uinput.KeySlash, ";": uinput.KeySemicolon,
		"'": uinput.KeyApostrophe, "[": uinput.KeyLeftbrace,
		"]": uinput.KeyRightbrace, "`": uinput.KeyGrave,
		"\\": uinput.KeyBackslash,
	}
	for token, want := range cases {
		b, err := Resolve(token)
		if err != nil {
			t.Fatalf("Resolve(%q): %v", token, err)
		}
		if got, ok := linuxKeyCode(b.VK); !ok || got != want {
			t.Errorf("%q: linuxKeyCode(%#x) = %d, %t, want %d", token, b.VK, got, ok, want)
		}
	}
	if _, ok := linuxKeyCode(0xFF); ok {
		t.Fatal("unsupported VK must not have a Linux mapping")
	}
}

// An injected device exercises the Linux path in CI without /dev/uinput access.
type fakeLinuxKeyboard struct {
	events  []string
	downErr error
	upErr   error
	closed  bool
}

func (f *fakeLinuxKeyboard) KeyPress(code int) error {
	f.events = append(f.events, fmt.Sprint("press:", code))
	return nil
}
func (f *fakeLinuxKeyboard) KeyDown(code int) error {
	f.events = append(f.events, fmt.Sprint("down:", code))
	return f.downErr
}
func (f *fakeLinuxKeyboard) KeyUp(code int) error {
	f.events = append(f.events, fmt.Sprint("up:", code))
	return f.upErr
}
func (f *fakeLinuxKeyboard) FetchSyspath() (string, error) {
	return "", nil
}
func (f *fakeLinuxKeyboard) Close() error {
	f.closed = true
	return nil
}

func useLinuxKeyboard(t *testing.T, device uinput.Keyboard, factory func() (uinput.Keyboard, error)) {
	t.Helper()
	linuxKeyboardMu.Lock()
	oldDevice, oldFactory := linuxKeyboard, linuxKeyboardFactory
	linuxKeyboard, linuxKeyboardFactory = device, factory
	linuxKeyboardMu.Unlock()
	t.Cleanup(func() {
		linuxKeyboardMu.Lock()
		linuxKeyboard, linuxKeyboardFactory = oldDevice, oldFactory
		linuxKeyboardMu.Unlock()
	})
}

func TestLinuxKeyboardPressAndRelease(t *testing.T) {
	fake := &fakeLinuxKeyboard{}
	useLinuxKeyboard(t, nil, func() (uinput.Keyboard, error) { return fake, nil })

	b, err := Resolve("left")
	if err != nil {
		t.Fatal(err)
	}
	if !Down(b) || !Up(b) {
		t.Fatal("left key down/up failed")
	}
	Tap(b)
	ReleaseVK(b.VK)
	want := []string{
		fmt.Sprint("down:", uinput.KeyLeft),
		fmt.Sprint("up:", uinput.KeyLeft),
		fmt.Sprint("down:", uinput.KeyLeft),
		fmt.Sprint("up:", uinput.KeyLeft),
		fmt.Sprint("up:", uinput.KeyLeft),
	}
	if !reflect.DeepEqual(fake.events, want) {
		t.Fatalf("events = %v, want %v", fake.events, want)
	}
}

func TestLinuxKeyboardInitFailureCanRetry(t *testing.T) {
	calls := 0
	fake := &fakeLinuxKeyboard{}
	useLinuxKeyboard(t, nil, func() (uinput.Keyboard, error) {
		calls++
		if calls == 1 {
			return nil, fmt.Errorf("permission denied")
		}
		return fake, nil
	})

	b, err := Resolve("a")
	if err != nil {
		t.Fatal(err)
	}
	if Down(b) || !Down(b) || !Up(b) {
		t.Fatal("expected first keydown to fail and second to succeed")
	}
	if calls != 2 {
		t.Fatalf("keyboard initializations: %d, want 2", calls)
	}
}

func TestLinuxKeyboardWriteFailureCanRecover(t *testing.T) {
	for _, edge := range []string{"down", "up"} {
		t.Run(edge, func(t *testing.T) {
			broken := &fakeLinuxKeyboard{}
			replacement := &fakeLinuxKeyboard{}
			useLinuxKeyboard(t, broken, func() (uinput.Keyboard, error) { return replacement, nil })

			b, err := Resolve("a")
			if err != nil {
				t.Fatal(err)
			}
			writeErr := fmt.Errorf("device write failed")
			if edge == "down" {
				broken.downErr = writeErr
				if Down(b) {
					t.Fatal("key down reported success after a failed write")
				}
			} else {
				if !Down(b) {
					t.Fatal("could not hold a key before the failed release")
				}
				broken.upErr = writeErr
				if Up(b) {
					t.Fatal("key up reported success after a failed write")
				}
			}
			if !broken.closed {
				t.Fatal("failed keyboard was left open with potentially held keys")
			}
			if !Down(b) || len(replacement.events) != 1 {
				t.Fatal("next key down did not use a replacement keyboard")
			}
		})
	}
}
