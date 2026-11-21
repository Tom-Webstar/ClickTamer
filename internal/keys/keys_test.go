package keys

import "testing"

func TestResolve(t *testing.T) {
	cases := map[string]uint16{
		"a": 0x41, "z": 0x5A, "0": 0x30, "9": 0x39,
		"up": 0x26, "PAGEUP": 0x21, "pagedown": 0x22, "f1": 0x70, "f12": 0x7B,
		"-": 0xBD, "esc": 0x1B,
	}
	for in, want := range cases {
		got, err := Resolve(in)
		if err != nil {
			t.Fatalf("Resolve(%q): %v", in, err)
		}
		if got.VK != want {
			t.Errorf("Resolve(%q) = %#x, want %#x", in, got.VK, want)
		}
	}
	for _, bad := range []string{"", "nope", "f0", "f13", "ab", "f1x"} {
		if _, err := Resolve(bad); err == nil {
			t.Errorf("Resolve(%q): want error", bad)
		}
	}
}

func TestWindowTargetSetGet(t *testing.T) {
	SetWindowTarget("  MyWhoosh ")
	if got := CurrentWindowTarget(); got != "MyWhoosh" {
		t.Fatalf("got %q", got)
	}
	SetWindowTarget("")
	if got := CurrentWindowTarget(); got != "" {
		t.Fatalf("got %q", got)
	}
}

func TestExtendedNamedKeyValues(t *testing.T) {
	cases := map[string]uint16{
		"left": 0x25, "up": 0x26, "right": 0x27, "down": 0x28,
		"insert": 0x2D, "delete": 0x2E, "home": 0x24, "end": 0x23,
		"pageup": 0x21, "pagedown": 0x22,
	}
	if len(extended_namedkeys) != len(cases) {
		t.Fatalf("extended_namedkeys has %d keys, want %d", len(extended_namedkeys), len(cases))
	}
	for name, vk := range cases {
		if got, ok := namedKeys[name]; !ok || got != vk {
			t.Errorf("namedKeys[%q] = %#x (found %t), want %#x", name, got, ok, vk)
		}
		if !extended_namedkeys[vk] {
			t.Errorf("extended_namedkeys missing %#x for %s", vk, name)
		}
	}
}
