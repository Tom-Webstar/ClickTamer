//go:build darwin && !ios && cgo

package quartz

/*
#cgo LDFLAGS: -framework CoreGraphics -framework CoreFoundation
#include <stdbool.h>
#include <CoreFoundation/CoreFoundation.h>
#include <CoreGraphics/CoreGraphics.h>

static bool quartzPostKey(CGKeyCode code, bool down) {
    CGEventRef event = CGEventCreateKeyboardEvent(NULL, code, down);
    if (event == NULL) {
        return false;
    }
    CGEventPost(kCGHIDEventTap, event);
    CFRelease(event);
    return true;
}
*/
import "C"

import (
	"errors"
	"fmt"
)

// KeyDown posts a key-down event for a macOS virtual key code.
func KeyDown(code uint16) error {
	return postKey(code, true)
}

// KeyUp posts a key-up event for a macOS virtual key code.
func KeyUp(code uint16) error {
	return postKey(code, false)
}

func postKey(code uint16, down bool) error {
	if !bool(C.CGPreflightPostEventAccess()) {
		return errors.New("quartz: Accessibility permission is required to post keyboard events")
	}
	if !bool(C.quartzPostKey(C.CGKeyCode(code), C.bool(down))) {
		return fmt.Errorf("quartz: could not create keyboard event for key code %#x", code)
	}
	return nil
}
