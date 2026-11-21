//go:build darwin && !ios && cgo

package keys

import "clicktamer/internal/Quartz"

// Tap posts a complete press and release to the focused application.
func Tap(b Binding) {
	if Down(b) {
		Up(b)
	}
}

// Down posts a Quartz key-down event.
func Down(b Binding) bool { return postQuartzKey(b.VK, quartz.KeyDown) }

// Up posts a Quartz key-up event even if focus has changed.
func Up(b Binding) bool { return postQuartzKey(b.VK, quartz.KeyUp) }

// ReleaseVK frees a held key during disconnect or profile changes.
func ReleaseVK(vk uint16) { Up(Binding{VK: vk}) }

// TargetWindowPresent is always true because Quartz uses the focused application.
func TargetWindowPresent() bool { return true }
