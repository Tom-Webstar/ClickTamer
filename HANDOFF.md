# Handoff

## 2026-11-21 — Versioned build filenames

Windows builds produce `clicktamer-win-amd64-<version>.exe` and release ZIPs
use the same base name. Linux builds and CI tarballs use
`clicktamer-linux-amd64-<version>`. The build recipes stamp the filename's
version into the binary. Releases use semantic-release's version. Local
builds use Git's version unless a version is supplied explicitly.

Verified both AMD64 builds, the Windows executable's `-version`, Windows
ZIP members and checksums, the Linux tarball, and the Scoop manifest.
Scoop saves the versioned download as `clicktamer.exe`. The pinned older
manifest retains its published download. Go tests, vet, and workflow lint
passed. Verify the uploads on the next hosted release and Linux build.

## 2026-11-21 — dev-linux review

Reviewed `dev-linux` at `0a20ab2` against `dev` at `a6ec371`.

- Linux uinput advertises held-key repeat and recreates the virtual keyboard
  after an event write fails. Device teardown closes the file even when its
  destroy ioctl fails.
- BlueZ scan failures clear adapter state. Session teardown stops button
  callbacks, releases held keys, unsubscribes, and disconnects. Windows keeps
  its probe-before-disconnect safeguard.
- Linux ignores the Windows focus and auto-launch options at runtime. Shared
  profiles keep those settings on disk. The Linux CI download is a tarball
  that preserves the executable bit.

Verification passed for Windows `go test ./...`, Windows and Linux AMD64
`go vet ./...` and builds, and all nine application test packages in a Linux
i686 VM. The nested uinput close regression passed in the VM. The archive
extracted with mode `0755` and unchanged binary contents. The live repeat test
skipped because the guest has no `/dev/uinput`.

Next, run `just ci-linux` on a Linux AMD64 host with access to `/dev/uinput`.
Check a held key, both pods, and reconnect after toggling Bluetooth. TinyGo
Bluetooth v0.17.0 still has a race between its internal scan cancellation and
the timeout call to `StopScan`. The app serializes its cleanup and recovers
panics. Recheck the dependency when upgrading it.

## 2026-11-20

`TestViewFooterFitsCardWidth` now requires exactly two lines for the full shortcut footer at the maximum card width of 136 usable cells. The test checks the width of each line. The existing renderer already produces this layout.

`just ci` passed on Windows, including formatting, vet, race-enabled tests across all nine packages, and the Windows amd64 build.

The footer expectation is in commit `3cc79f1`.

The release recipe targets `windows/amd64` and the Scoop manifest advertises only `64bit`. During the architecture check, `go build` also produced `windows/arm64` and `windows/386` executables from this source. `go version -m` confirmed each target. Runtime behaviour with controllers on those two architectures is unverified.

Windows 11 on Arm supports x64 emulation for the published executable. ClickTamer's Bluetooth and keyboard behaviour under emulation still needs an ARM64 hardware check.
