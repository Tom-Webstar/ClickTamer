//go:build darwin && !ios && cgo

package keys

import (
	"log/slog"
	"strconv"

	"clicktamer/internal/Quartz"
)

// The BLE hold tracker uses VKs for both ordinary edges and forced releases.
// Recover the token so each path can reuse the same Quartz mapping.
func quartzToken(vk uint16) string {
	switch {
	case vk >= 'A' && vk <= 'Z', vk >= '0' && vk <= '9':
		return string(rune(vk))
	case vk >= firstFunctionKeyVK && vk < firstFunctionKeyVK+functionKeyCount:
		return "f" + strconv.Itoa(int(vk-firstFunctionKeyVK)+1)
	}
	for token, code := range namedKeys {
		if code == vk {
			return token
		}
	}
	return ""
}

func postQuartzKey(vk uint16, post func(uint16) error) bool {
	code, err := quartz.Resolve(quartzToken(vk))
	if err != nil {
		slog.Warn("unsupported macOS key", "vk", vk)
		return false
	}
	if err := post(code); err != nil {
		slog.Warn("Quartz keyboard event failed", "vk", vk, "error", err)
		return false
	}
	return true
}
