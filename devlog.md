# ClickTamer handoff

## 2026-11-19 — Dedicated arrow keys

Windows input now uses the dedicated E0 scan for each arrow on presses, repeats and releases. Left is VK 25 / E0 4B, Up 26 / E0 48, Right 27 / E0 4D and Down 28 / E0 50 (all hex). Windows derives the VK from the scan in this injection mode. The shipped Notepad double profile now maps the four directions to the matching arrows. Explicit letter bindings still produce letters.

The regression tests reproduced the missing extended flag for every arrow and the four literal letters in the old Notepad profile. All nine packages then passed, as did vet and the Windows amd64 build. A native check resolved the encoded events to all four expected Windows key names and VKs. A live injection probe stopped at its initial marker because Windows denied SendInput from the test process. In-game delivery still needs a controller check.

The new binary is `dist/clicktamer.exe`. The running copy observed during the check was the older `clicktamer.exe` in the repository root. Exit it before starting the new build. The installed config now parses and selects BikeTerra in double mode with the correct four arrow tokens. Its separate Notepad profile still contains the original four letters.

## 2026-11-19 — BikeTerra camera binding test

The user confirmed the shipped single-controller profile is authoritative. Updated `TestAppProfilesMapBothPods` in `profiles_test.go` to expect `c` for BikeTerra's `Y` button (cycle camera). The focused test originally failed with `Y is bound to "c", want "u"`. After the edit, `go test -count=1 ./...` passed across all nine packages, `go vet ./...` passed, and the Windows amd64 build succeeded.

The BikeTerra fix and the later Notepad changes share files. Stage their hunks separately when committing with write access to Git metadata. The existing user edit in `justfile` is independent.

## 2026-11-19 — YAML error reported at line 10

Fixed the indentation of `biketerra:` in `config.yaml` at line 230. Its single leading space broke the `profiles` mapping. The installed `%LOCALAPPDATA%\clicktamer\config.yml` initially contained the same typo and has since been corrected.

`TestShippedConfigDefaultProfileResolves` reproduced the original error and passed after the edit. All ten shipped profiles load. Correcting the indentation exposed the stale camera-binding assertion addressed above.

## 2026-11-19 — Arrow keys appear as numpad digits

Diagnosed on Windows at `5468be6`. The profile tokens resolve to the correct arrow virtual keys. `MapVirtualKeyW` in mode 4 returned `004B`, `004D`, `0048` and `0050` for Left, Right, Up and Down. `newScanKeyboardInput` sets `KEYEVENTF_EXTENDEDKEY` only for a returned `E0` prefix. All four events therefore used flags `0x08` on press and `0x0A` on release.

A temporary Go test called the real Windows mapper and the repository's event constructor. `GetKeyNameTextW` identified the resulting scans as Num 4, Num 6, Num 8 and Num 2. Setting the extended bit identified the dedicated arrows in every case. `TestScanCodeEncoding` passes because its arrow examples already include `E0`.

This diagnosis led to the dedicated-arrow fix recorded above. The original probe used no BLE hardware or foreground input.

At the time of the arrow diagnosis, the keyboard package tests and `go vet ./...` passed, and the full suite encountered the YAML issue described above. The local `justfile` already had edits at session start.
