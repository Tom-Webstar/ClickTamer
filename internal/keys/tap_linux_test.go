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
		"shift": uinput.KeyLeftshift, "ctrl": uinput.KeyLeftctrl,
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
	events []string
}
func (f *fakeLinuxKeyboard) KeyPress(code int) error { f.events = append(f.events, fmt.Sprint("press:", code)); return nil }
func (f *fakeLinuxKeyboard) KeyDown(code int) error { f.events = append(f.events, fmt.Sprint("down:", code)); return nil }
func (f *fakeLinuxKeyboard) KeyUp(code int) error { f.events = append(f.events, fmt.Sprint("up:", code)); return nil }
func (f *fakeLinuxKeyboard) FetchSyspath() (string, error) { return "", nil }
func (f *fakeLinuxKeyboard) Close() error { return nil }

func TestLinuxKeyboardPressAndRelease(t *testing.T) {
	fake := &fakeLinuxKeyboard{}
	linuxKeyboardMu.Lock()
	oldDevice, oldFactory := linuxKeyboard, linuxKeyboardFactory
	linuxKeyboard = nil
	linuxKeyboardFactory = func() (uinput.Keyboard, error) { return fake, nil }
	linuxKeyboardMu.Unlock()
	t.Cleanup(func() {
		linuxKeyboardMu.Lock()
		linuxKeyboard, linuxKeyboardFactory = oldDevice, oldFactory
		linuxKeyboardMu.Unlock()
	})

	b, err := Resolve("left")
	if err != nil { t.Fatal(err) }
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
	linuxKeyboardMu.Lock()
	oldDevice, oldFactory := linuxKeyboard, linuxKeyboardFactory
	linuxKeyboard = nil
	calls := 0
	fake := &fakeLinuxKeyboard{}
	linuxKeyboardFactory = func() (uinput.Keyboard, error) {
		calls++
		if calls == 1 { return nil, fmt.Errorf("permission denied") }
		return fake, nil
	}
	linuxKeyboardMu.Unlock()
	t.Cleanup(func() {
		linuxKeyboardMu.Lock()
		linuxKeyboard, linuxKeyboardFactory = oldDevice, oldFactory
		linuxKeyboardMu.Unlock()
	})

	b, err := Resolve("a")
	if err != nil { t.Fatal(err) }
	if Down(b) || !Down(b) || !Up(b) {
		t.Fatal("expected first keydown to fail and second to succeed")
	}
	if calls != 2 { t.Fatalf("keyboard initializations: %d, want 2", calls) }
}
