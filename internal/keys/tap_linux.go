//go:build linux

package keys

import (
	"log/slog"
	"sync"

	"github.com/bendahl/uinput"
)

// Linux synthesizes keypresses through /dev/uinput rather than Windows SendInput.
// Creating the device on the first keypress keeps -demo and -version usable
// without uinput permissions and lets subsequent presses retry after a failure.
const linuxUinputPath = "/dev/uinput"

var (
	linuxKeyboardMu      sync.Mutex
	linuxKeyboard        uinput.Keyboard
	linuxKeyboardFactory = func() (uinput.Keyboard, error) {
		return uinput.CreateKeyboard(linuxUinputPath, []byte("ClickTamer Virtual Keyboard"))
	}
)

// linuxKeyCode maps the public Windows VK bindings to Linux evdev key codes.
// Letter and number codes are not contiguous in evdev, so use explicit tables.
func linuxKeyCode(vk uint16) (int, bool) {
	if vk >= 'A' && vk <= 'Z' {
		letters := [...]int{
			uinput.KeyA, uinput.KeyB, uinput.KeyC, uinput.KeyD, uinput.KeyE,
			uinput.KeyF, uinput.KeyG, uinput.KeyH, uinput.KeyI, uinput.KeyJ,
			uinput.KeyK, uinput.KeyL, uinput.KeyM, uinput.KeyN, uinput.KeyO,
			uinput.KeyP, uinput.KeyQ, uinput.KeyR, uinput.KeyS, uinput.KeyT,
			uinput.KeyU, uinput.KeyV, uinput.KeyW, uinput.KeyX, uinput.KeyY,
			uinput.KeyZ,
		}
		return letters[vk-'A'], true
	}
	if vk >= '0' && vk <= '9' {
		digits := [...]int{
			uinput.Key0, uinput.Key1, uinput.Key2, uinput.Key3, uinput.Key4,
			uinput.Key5, uinput.Key6, uinput.Key7, uinput.Key8, uinput.Key9,
		}
		return digits[vk-'0'], true
	}
	if vk >= 0x70 && vk <= 0x79 { // F1-F10
		return uinput.KeyF1 + int(vk-0x70), true
	}
	switch vk {
	case 0x7A:
		return uinput.KeyF11, true
	case 0x7B:
		return uinput.KeyF12, true
	case 0x25:
		return uinput.KeyLeft, true
	case 0x26:
		return uinput.KeyUp, true
	case 0x27:
		return uinput.KeyRight, true
	case 0x28:
		return uinput.KeyDown, true
	case 0x2D:
		return uinput.KeyInsert, true
	case 0x2E:
		return uinput.KeyDelete, true
	case 0x24:
		return uinput.KeyHome, true
	case 0x23:
		return uinput.KeyEnd, true
	case 0x21:
		return uinput.KeyPageup, true
	case 0x22:
		return uinput.KeyPagedown, true
	case 0x0D:
		return uinput.KeyEnter, true
	case 0x20:
		return uinput.KeySpace, true
	case 0x09:
		return uinput.KeyTab, true
	case 0x1B:
		return uinput.KeyEsc, true
	case 0x08:
		return uinput.KeyBackspace, true
	case 0x10:
		return uinput.KeyLeftshift, true
	case 0x11:
		return uinput.KeyLeftctrl, true
	case 0x12:
		return uinput.KeyLeftalt, true
	case 0x14:
		return uinput.KeyCapslock, true
	case 0xBD:
		return uinput.KeyMinus, true
	case 0xBB:
		return uinput.KeyEqual, true
	case 0xBC:
		return uinput.KeyComma, true
	case 0xBE:
		return uinput.KeyDot, true
	case 0xBF:
		return uinput.KeySlash, true
	case 0xBA:
		return uinput.KeySemicolon, true
	case 0xDE:
		return uinput.KeyApostrophe, true
	case 0xDB:
		return uinput.KeyLeftbrace, true
	case 0xDD:
		return uinput.KeyRightbrace, true
	case 0xC0:
		return uinput.KeyGrave, true
	case 0xDC:
		return uinput.KeyBackslash, true
	}
	return 0, false
}

// Tap injects one complete press/release pair.
func Tap(b Binding) {
	if Down(b) {
		Up(b)
	}
}

// Down presses and holds the mapped Linux key. Linux input handles repeat
// while held, so no Windows-style repeat goroutine is required here.
func Down(b Binding) bool {
	code, ok := linuxKeyCode(b.VK)
	if !ok {
		slog.Warn("unsupported Linux key", "vk", b.VK)
		return false
	}
	linuxKeyboardMu.Lock()
	defer linuxKeyboardMu.Unlock()
	if linuxKeyboard == nil {
		dev, err := linuxKeyboardFactory()
		if err != nil {
			slog.Warn("cannot create virtual keyboard (check /dev/uinput permissions)", "error", err)
			return false
		}
		linuxKeyboard = dev
	}
	if err := linuxKeyboard.KeyDown(code); err != nil {
		slog.Warn("Linux keyboard key down failed", "vk", b.VK, "error", err)
		discardLinuxKeyboard()
		return false
	}
	return true
}

// Up releases a key even if focus has changed.
func Up(b Binding) bool {
	code, ok := linuxKeyCode(b.VK)
	if !ok {
		return false
	}
	linuxKeyboardMu.Lock()
	defer linuxKeyboardMu.Unlock()
	if linuxKeyboard == nil {
		return false
	}
	if err := linuxKeyboard.KeyUp(code); err != nil {
		slog.Warn("Linux keyboard key up failed", "vk", b.VK, "error", err)
		discardLinuxKeyboard()
		return false
	}
	return true
}

// A failed write can leave key state uncertain. Closing the device releases
// its held keys, and the next press opens a fresh keyboard. Caller holds the mutex.
func discardLinuxKeyboard() {
	err := linuxKeyboard.Close()
	linuxKeyboard = nil
	if err != nil {
		slog.Warn("Linux keyboard close failed", "error", err)
	}
}

// ReleaseVK frees a held key when a BLE session ends.
func ReleaseVK(vk uint16) { Up(Binding{VK: vk}) }

// Linux does not implement the Windows-specific window-title focus lookup.
func TargetWindowPresent() bool { return true }
