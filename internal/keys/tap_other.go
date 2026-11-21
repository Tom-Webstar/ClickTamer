//go:build !windows && !linux && (!darwin || ios || !cgo)

package keys

// Tap/Down/Up/ReleaseVK are no-ops on builds without an input backend.
// Down and Up report success so the hold tracker observes both edges.
func Tap(b Binding) {}

func Down(b Binding) bool { return true }

func Up(b Binding) bool { return true }

func ReleaseVK(vk uint16) {}

// TargetWindowPresent is always true off Windows (no focus feature to warn about).
func TargetWindowPresent() bool { return true }
