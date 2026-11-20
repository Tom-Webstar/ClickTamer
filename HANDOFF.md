# Handoff

## 2026-11-20

`TestViewFooterFitsCardWidth` now requires exactly two lines for the full shortcut footer at the maximum card width of 136 usable cells. The test checks the width of each line. The existing renderer already produces this layout.

`just ci` passed on Windows, including formatting, vet, race-enabled tests across all nine packages, and the Windows amd64 build.

The commit is pending because this checkout denied creation of `.git/index.lock`. With write access, commit the footer test and this handoff as `fix: expect two lines in the full footer test`.

The release recipe targets `windows/amd64` and the Scoop manifest advertises only `64bit`. During the architecture check, `go build` also produced `windows/arm64` and `windows/386` executables from this source. `go version -m` confirmed each target. Runtime behaviour with controllers on those two architectures is unverified.

Windows 11 on Arm supports x64 emulation for the published executable. ClickTamer's Bluetooth and keyboard behaviour under emulation still needs an ARM64 hardware check.
