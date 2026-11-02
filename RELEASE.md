# Distributing Delta

A plan to ship `delta` as a ready-made binary, so nobody needs Go to use it.

**Goal.** Someone on Linux, Windows or macOS installs `delta` with one command and has it on their `PATH`. No Go, no clone, no build.

**How.** Pushing a version tag makes GitHub Actions run GoReleaser, which builds the binaries, packs them and publishes them as a GitHub Release. Two small install scripts download the right one.

---

## 1. Where things stand

Checked on this repository:

| Fact | Consequence |
|---|---|
| The entrypoint is `./cmd/delta`, named after the binary, which must be called `delta` (hooks, permissions, `Locate("delta")` and the PATH check depend on it) | The name comes out right by default. Keep `binary: delta` in GoReleaser anyway, to make it explicit. `go install github.com/ronz204/deltas-ai/cmd/delta@latest` also works |
| `main.version` already exists (`cmd/delta/main.go:10`, default `"dev"`) and is set with `-ldflags "-X main.version=..."` | GoReleaser can inject the tag with no code change |
| `delta version` prints `delta (deltas-ai) <version>` | Keep that prefix. Only the version part changes |
| The code is pure Go (`CGO_ENABLED=0` builds) | No C toolchain, no per-OS runner needed |
| `go build` for `linux`, `windows` and `darwin`, each on `amd64` and `arm64`, all succeed | Cross-compiling from one Linux runner works |
| CI (`.github/workflows/ci.yml`) already tests on Linux, macOS and Windows | The release can rely on it |
| The remote is `github.com/ronz204/deltas-ai` | The release and install URLs derive from it |
| `dist/` is already in `.gitignore` | GoReleaser's output folder is `dist/` too, so nothing new to ignore |

No source code needs to change for the binary itself. The work is configuration, two scripts and documentation.

---

## 2. What gets built

| OS | Architectures | Archive | Contains |
|---|---|---|---|
| Linux | `amd64`, `arm64` | `.tar.gz` | `delta` |
| Windows | `amd64`, `arm64` | `.zip` | `delta.exe` |
| macOS | `amd64`, `arm64` | `.tar.gz` | `delta` |

Archive names follow `delta_<version>_<os>_<arch>.<ext>`, for example `delta_0.1.0_linux_amd64.tar.gz`. A `checksums.txt` with SHA-256 sums is published next to them.

macOS costs nothing extra, because it is the same Go cross-compile. It is built and published but not documented as a main target.

---

## 3. Files to add

### 3.1 `.goreleaser.yaml`

```yaml
version: 2

project_name: delta

before:
  hooks:
    - go mod tidy
    - go vet ./...

builds:
  - id: delta
    main: ./cmd/delta
    binary: delta
    env:
      - CGO_ENABLED=0
    goos: [linux, windows, darwin]
    goarch: [amd64, arm64]
    flags:
      - -trimpath
    ldflags:
      - -s -w -X main.version={{ .Version }}

archives:
  - id: delta
    formats: [tar.gz]
    name_template: "delta_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
    format_overrides:
      - goos: windows
        formats: [zip]
    files:
      - LICENSE*
      - README.md
      - MANUAL.md

checksum:
  name_template: checksums.txt

changelog:
  use: github
  filters:
    exclude:
      - "^docs"
      - "^chore"
      - "^test"

release:
  draft: false
  prerelease: auto
```

Notes on the choices:

- `main: ./cmd/delta` points at the entrypoint and `binary: delta` fixes the name. They agree today; stating both keeps a future folder rename from changing the binary.
- `{{ .Version }}` is the tag without the leading `v`, so `v0.1.0` gives `0.1.0`. `delta version` then prints `delta (deltas-ai) 0.1.0`.
- `-s -w -trimpath` shrinks the binary and removes local paths. Debug symbols are not needed.
- `prerelease: auto` marks tags like `v0.2.0-rc1` as pre-releases by itself.
- `files` ships the docs inside each archive. Remove `LICENSE*` if the repository has no license yet (see section 6).
- `go vet` runs before the build as a last guard. The full test suite is the job of `ci.yml`.

Validate it locally before the first tag with `goreleaser check`, and try a build with `goreleaser release --snapshot --clean`, which publishes nothing.

### 3.2 `.github/workflows/release.yml`

```yaml
name: release

on:
  push:
    tags: ["v*"]

permissions:
  contents: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
      - uses: goreleaser/goreleaser-action@v6
        with:
          version: "~> v2"
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

- `fetch-depth: 0` is required: GoReleaser reads the git history for the changelog.
- `permissions: contents: write` lets the built-in `GITHUB_TOKEN` create the release. No personal token or secret is needed.
- Tests are not repeated here. Only tag a commit whose `ci` run is green (see section 5).

### 3.3 `install.sh` (Linux and macOS)

Placed at the repository root so it is served from `raw.githubusercontent.com`.

What it must do:

1. Detect the OS (`uname -s`) and the architecture (`uname -m`: `x86_64` becomes `amd64`, `aarch64` or `arm64` becomes `arm64`). Stop with a clear message on anything else.
2. Pick the version: `$DELTA_VERSION` if set, otherwise the latest release (`https://api.github.com/repos/ronz204/deltas-ai/releases/latest`).
3. Download `delta_<version>_<os>_<arch>.tar.gz` and `checksums.txt` into a temporary directory.
4. **Verify the SHA-256** of the archive against `checksums.txt` (`sha256sum`, or `shasum -a 256` on macOS). Stop if it does not match.
5. Extract `delta` and install it into `$DELTA_INSTALL_DIR`, defaulting to `~/.local/bin`, with mode `755`.
6. Run `delta version` to confirm it works.
7. If the install directory is not on `PATH`, print the exact line to add to the shell profile. Do not edit the profile.
8. Clean up the temporary directory, even on failure (`trap`).

Usage:

```sh
curl -fsSL https://raw.githubusercontent.com/ronz204/deltas-ai/main/install.sh | sh
```

Write it in POSIX `sh` (not bash-only), with `set -eu`. Only use `curl` or `wget`, `tar`, and `sha256sum` or `shasum`.

### 3.4 `install.ps1` (Windows)

What it must do:

1. Detect the architecture (`$env:PROCESSOR_ARCHITECTURE`: `AMD64` becomes `amd64`, `ARM64` becomes `arm64`).
2. Pick the version: `$env:DELTA_VERSION` if set, otherwise the latest release from the GitHub API.
3. Download `delta_<version>_windows_<arch>.zip` and `checksums.txt` to a temporary folder.
4. **Verify the SHA-256** (`Get-FileHash`). Stop if it does not match.
5. Extract `delta.exe` (`Expand-Archive`) into `$env:DELTA_INSTALL_DIR`, defaulting to `$env:LOCALAPPDATA\Programs\delta`.
6. Add that folder to the **user** `PATH` (`[Environment]::SetEnvironmentVariable(..., 'User')`) if it is not there yet. This needs no administrator rights. Say that open terminals must be restarted.
7. Run `delta version` to confirm it works.
8. Clean up the temporary folder.

Usage:

```powershell
irm https://raw.githubusercontent.com/ronz204/deltas-ai/main/install.ps1 | iex
```

Set `$ErrorActionPreference = 'Stop'` and use TLS 1.2 explicitly, so an older Windows PowerShell does not fail on the download.

---

## 4. Files to change

| File | Change |
|---|---|
| `MANUAL.md` → "Install the CLI" (around line 36) | Make the one-line installers the main path. Move "build from source with Go" into a short "Build from source" subsection. Update the requirements list: Go is only needed to build from source |
| `MANUAL.md` → "Requirements" (around line 28) | Drop "Go, only to build the `delta` binary" from the required list. Mention it as optional |
| `MANUAL.md` → `delta version` (around line 44) | `delta (deltas-ai) dev` is what a local build prints. An installed release prints the real version, for example `delta (deltas-ai) 0.1.0` |
| `MANUAL.md` → Troubleshooting (around line 372) | The fix for "delta is not on the PATH" should mention the installer's folder (`~/.local/bin`, `%LOCALAPPDATA%\Programs\delta`), not only `dist/delta` |
| `MANUAL.md` → "Keeping Delta up to date" | Add: to update the binary itself, run the installer again (it replaces the old one). `delta update` still updates the installed harness, not the binary |
| `README.md` → top | Add an "Install" section with the two one-liners right after "What is Delta?", so a newcomer sees it early |
| `README.md` → "Development" | Keep the `go build` lines, since they are for contributors. Add a link to this file for how releases are made |
| `.github/workflows/ci.yml` | Optional: add a `goreleaser check` step, and a `goreleaser release --snapshot --clean` step, so a broken release config is caught on a normal pull request instead of on a tag |

No change is needed in Go code, tests, hooks or `delta init`.

---

## 5. How to release

Once everything above is merged:

```sh
git switch main
git pull
# make sure the latest ci run on main is green
git tag v0.1.0
git push origin v0.1.0
```

Then:

1. The `release` workflow starts. Follow it in the **Actions** tab.
2. When it finishes, the **Releases** page has the six archives, `checksums.txt` and a changelog built from the commit messages.
3. Test the installers on a clean machine (or a container or VM) against that release.
4. Edit the release notes if you want a human summary at the top.

Version numbers follow semantic versioning with a leading `v`. Use `v0.x.y` while the contract format and the commands may still change.

If a release goes wrong, delete the release and the tag (`git push --delete origin v0.1.0`, `git tag -d v0.1.0`), fix the problem and tag again.

---

## 6. Open decisions

These need an answer before or during the work.

1. **Is the repository public?** Decided: **yes**. The install scripts download from `raw.githubusercontent.com` and the Releases API without a token, which works because the repository is public.
2. **Is there a license?** Decided: **MIT**, in `LICENSE` at the repository root, so the `LICENSE*` entry in `.goreleaser.yaml` stays. One thing to check: `agents/rules/code-change-scope.md` says its ladder is adapted from ponytail (MIT). MIT asks that the original copyright notice travels with substantial copied portions, and the README acknowledgment that used to carry it was removed. Decide whether to restore a credit (for example a `NOTICE` file or a line in the rule).
3. **Windows SmartScreen and macOS Gatekeeper.** An unsigned binary can show a warning on first run. This is cosmetic for a CLI installed from a script (the script does not trigger it the way a browser download does), but a binary downloaded by hand may. Signing needs paid certificates. Not worth it for now; mention it in the manual instead.
4. **Package managers (later).** Scoop for Windows and a Homebrew tap for Linux and macOS can be added to the same `.goreleaser.yaml` once people use the tool. Both need a separate repository for the manifest. Out of scope for this first pass.
5. **Install script location.** Serving from `main` means the script a user runs always matches the latest scripts, but not necessarily the latest tag. Acceptable, because the script picks the release itself.

---

## 7. Order of work

1. Decide the open points in section 6 (public repo, license).
2. Add `.goreleaser.yaml`. Run `goreleaser check` and `goreleaser release --snapshot --clean`. Run the resulting `dist/delta` from the Linux or Windows archive and confirm `delta version` prints the snapshot version.
3. Add `.github/workflows/release.yml`.
4. Add `install.sh` and `install.ps1`. Test each against the snapshot archives first by pointing them at a local file or a pre-release.
5. Update `MANUAL.md` and `README.md` (section 4).
6. Optionally add the GoReleaser steps to `ci.yml`.
7. Commit in small steps: one for the GoReleaser config and workflow, one for the installers, one for the docs.
8. Tag `v0.1.0` (or `v0.1.0-rc1` first as a dry run, which `prerelease: auto` marks as a pre-release) and test the installers on clean machines.

## 8. How to know it works

- `goreleaser release --snapshot --clean` produces six archives and `checksums.txt` in `dist/`.
- On a machine **without Go**, the one-liner installs `delta`, and `delta version` prints `delta (deltas-ai) 0.1.0`.
- `delta init` in a fresh git repository installs the harness, and `delta status` reports no warning about the PATH.
- The installers refuse to install when the checksum is wrong (edit `checksums.txt` on purpose to test it).
- Running the installer a second time replaces the binary without errors.
