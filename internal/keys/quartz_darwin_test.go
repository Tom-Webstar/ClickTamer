//go:build darwin && !ios && cgo

package keys

import (
	"errors"
	"strconv"
	"testing"

	"clicktamer/internal/Quartz"
)

func TestQuartzPostsEveryResolvedBinding(t *testing.T) {
	check := func(token string) {
		t.Helper()
		binding, err := Resolve(token)
		if err != nil {
			t.Fatal(err)
		}
		want, err := quartz.Resolve(token)
		if err != nil {
			t.Fatal(err)
		}
		calls := 0
		if !postQuartzKey(binding.VK, func(code uint16) error {
			calls++
			if code != want {
				t.Errorf("post %q = %#x, want %#x", token, code, want)
			}
			return nil
		}) || calls != 1 {
			t.Errorf("posting %q made %d calls, want one successful call", token, calls)
		}
	}
	for token := range namedKeys {
		check(token)
	}
	for token := 'a'; token <= 'z'; token++ {
		check(string(token))
	}
	for token := '0'; token <= '9'; token++ {
		check(string(token))
	}
	for n := 1; n <= 12; n++ {
		check("f" + strconv.Itoa(n))
	}
}

func TestQuartzDropsUnsupportedKeys(t *testing.T) {
	if postQuartzKey(0xFFFF, func(uint16) error {
		t.Fatal("called the poster for an unsupported key")
		return nil
	}) {
		t.Fatal("unsupported key reported success")
	}
}

func TestQuartzReportsPostingError(t *testing.T) {
	calls := 0
	sent := postQuartzKey('A', func(uint16) error {
		calls++
		return errors.New("posting denied")
	})
	if sent || calls != 1 {
		t.Fatalf("post with error: success=%v, calls=%d, want false and one call", sent, calls)
	}
}
