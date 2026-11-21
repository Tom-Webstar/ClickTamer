# clicktamer tasks. Run `just` to list.

build_version := `git describe --tags --match '[0-9]*.[0-9]*.[0-9]*' --dirty 2>/dev/null || printf '%s\n' 0.0.0-dev`
windows_prefix := "clicktamer-win-amd64-"
linux_prefix := "clicktamer-linux-amd64-"

default:
    @just --list

# --- checks ---------------------------------------------------------------

test:
    go test -race ./...

# The upstream integration suite requires a writable /dev/uinput.
test-uinput:
    go -C internal/uinput vet ./...
    go -C internal/uinput test -race -run '^(TestCloseDeviceClosesFileAfterDestroyFailure|TestKeyboardAdvertisesKeyRepeat)$' ./...

vet:
    go vet ./...
    GOOS=windows go vet ./...

fmt:
    gofmt -w .

fmt-check:
    #!/bin/sh
    out=$(gofmt -l .)
    if [ -n "$out" ]; then
        echo "gofmt needed on:"; echo "$out"; exit 1
    fi

# --- build ----------------------------------------------------------------

# Resolve the default once when CI needs to name both a binary and its archive.
build-version:
    @printf '%s\n' {{quote(build_version)}}

_check-version version:
    @case {{quote(version)}} in ''|[!0-9]*|*[!0-9A-Za-z.+-]*) printf '%s\n' 'Invalid build version' >&2; exit 1 ;; esac

# Build the Windows exe icon resource (generated, never committed).
gen-icon:
    go run github.com/tc-hib/go-winres@v0.3.3 simply \
        --manifest none --arch amd64 --out rsrc \
        --icon res/icon/clicktamer-256.png

build version=build_version: (_check-version version) gen-icon
    GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath \
        -ldflags={{quote("-s -w -X main.version=" + version)}} \
        -o {{quote("dist/" + windows_prefix + version + ".exe")}} .

# Linux AMD64 is built separately from the Windows release.
build-linux version=build_version: (_check-version version)
    mkdir -p dist
    GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath \
        -ldflags={{quote("-s -w -X main.version=" + version)}} \
        -o {{quote("dist/" + linux_prefix + version)}} .

# macOS Apple Silicon build using Xcode Command Line Tools.
build-darwin-arm64 version=build_version: (_build-darwin "arm64" version)

# macOS Intel build using Xcode Command Line Tools.
build-darwin-amd64 version=build_version: (_build-darwin "amd64" version)

_build-darwin arch version: (_check-version version)
    @if [ {{quote(os())}} != 'macos' ]; then printf '%s\n' 'Darwin builds require macOS with Xcode Command Line Tools.' >&2; exit 1; fi
    mkdir -p dist
    GOOS=darwin GOARCH={{quote(arch)}} CGO_ENABLED=1 CC=clang go build -trimpath \
        -ldflags={{quote("-s -w -X main.version=" + version)}} \
        -o {{quote("dist/clicktamer-darwin-" + arch + "-" + version)}} .

# Check and build the Linux branch without touching Windows release assets.
ci-linux version=build_version: fmt-check vet test test-uinput (build-linux version)

# Local: formats, then checks + builds.
sanity: fmt test vet build

# CI: check-only, no rewrites.
ci: fmt-check vet test build

# --- release (needs zip + sha256sum: Ubuntu CI + local) -------------------

release version: (_check-version version)
    #!/usr/bin/env bash
    set -euo pipefail
    name={{quote(windows_prefix + version)}}
    rm -rf dist && mkdir dist
    just build {{quote(version)}}
    (cd dist && zip -q "$name.zip" "$name.exe")
    (cd dist && sha256sum "$name.exe" "$name.zip" > checksums.txt)

# --- scoop (local + CI) ---------------------------------------------------

# Read the unique Windows executable checksum for this release.
scoop-checksum $version $checksum_file:
    #!/usr/bin/env bash
    set -euo pipefail
    exe="{{windows_prefix}}${version}.exe"
    if ! awk -v exe="$exe" '
        $2 == exe { hash = $1; matches++ }
        END {
            if (matches != 1 || length(hash) != 64 || hash !~ /^[A-Fa-f0-9]+$/) exit 1
            print hash
        }
    ' "$checksum_file"; then
        printf 'expected one SHA256 for %s in %s\n' "$exe" "$checksum_file" >&2
        exit 1
    fi

# Render the ScoopInstaller/Extras manifest for VERSION + exe SHA256 to stdout.
scoop-manifest $version $hash:
    #!/usr/bin/env bash
    # Scoop saves the versioned download as clicktamer.exe via the URL fragment.
    set -euo pipefail
    exe="{{windows_prefix}}${version}.exe"
    cat <<EOF
    {
        "version": "$version",
        "description": "Turn a Zwift Click V2 controller into a Bluetooth keyboard for indoor cycling apps or any apps.",
        "homepage": "https://github.com/Tom-Webstar/clicktamer",
        "license": "MIT",
        "architecture": {
            "64bit": {
                "url": "https://github.com/Tom-Webstar/clicktamer/releases/download/$version/$exe#/clicktamer.exe",
                "hash": "$hash"
            }
        },
        "bin": "clicktamer.exe",
        "shortcuts": [["clicktamer.exe", "clicktamer"]],
        "checkver": "github",
        "autoupdate": {
            "architecture": {
                "64bit": {
                    "url": "https://github.com/Tom-Webstar/clicktamer/releases/download/\$version/{{windows_prefix}}\$version.exe#/clicktamer.exe"
                }
            },
            "hash": {
                "url": "\$baseurl/checksums.txt"
            }
        }
    }
    EOF

# Local end-to-end: build the release assets and render dist/clicktamer.json.
scoop-release $version: (release version)
    #!/usr/bin/env bash
    # Exercises the whole scoop pipeline before a live main push. POSIX-only
    # (zip/sha256sum), like `release`.
    set -euo pipefail
    sha=$(just scoop-checksum "$version" dist/checksums.txt)
    just scoop-manifest "$version" "$sha" > dist/clicktamer.json
    echo "wrote dist/clicktamer.json"

# --- deps -----------------------------------------------------------------

# Blanket dependency upgrade. Upgrades tinygo BLE libs too: hardware re-verify
# before a release that includes BLE dep bumps.
update-deps:
    go get -u ./...
    go mod tidy

# --- dev ------------------------------------------------------------------

run:
    go run .

demo:
    go run . -demo
