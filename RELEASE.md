# Releasing Delta

How `delta` is shipped as a ready-made binary, and how to cut a new release. Nobody needs Go to install it.

**Status.** `v0.1.0` is published. It was built by the `release` workflow and installed with the one-line installer on Windows.

---

## 1. How it works

Pushing a version tag starts the whole pipeline:

```text
git tag vX.Y.Z  ──▶  release workflow  ──▶  GoReleaser  ──▶  GitHub Release
   (you)             (GitHub Actions)      builds, packs,     six archives
                                           checksums          + checksums.txt
                                                                   │
        install.sh / install.ps1  ◀────────── download, verify ────┘
        (served from the main branch)         put delta on the PATH
```

Users never touch the pipeline. They run one of the installers, which pick a release, check its SHA-256 against `checksums.txt` and put `delta` on the `PATH`. Install and update steps for users are in the [manual](MANUAL.md#install-the-cli).

### What gets built

| OS | Architectures | Archive | Contains |
|---|---|---|---|
| Linux | `amd64`, `arm64` | `.tar.gz` | `delta`, `LICENSE`, docs |
| Windows | `amd64`, `arm64` | `.zip` | `delta.exe`, `LICENSE`, docs |
| macOS | `amd64`, `arm64` | `.tar.gz` | `delta`, `LICENSE`, docs |

Archive names are `delta_<version>_<os>_<arch>.<ext>`, for example `delta_0.1.0_linux_amd64.tar.gz`. The version is the tag without the leading `v`, and `checksums.txt` sits next to the archives.

---

## 2. The files involved

| File | What it does |
|---|---|
| `.goreleaser.yaml` | Builds `./cmd/delta` for the six targets, injects the version, packs the archives and writes the checksums |
| `.github/workflows/release.yml` | Runs GoReleaser when a `v*` tag is pushed. It uses the built-in `GITHUB_TOKEN`, so no secret is needed |
| `install.sh` | Installer for Linux and macOS, in POSIX `sh` |
| `install.ps1` | Installer for Windows PowerShell |
| `LICENSE` | MIT, shipped inside every archive |
| `cmd/delta/main.go` | Holds `var version = "dev"`, which the release overwrites |

Choices in `.goreleaser.yaml` worth knowing:

- `main: ./cmd/delta` and `binary: delta` agree today. The binary must be called `delta` (hooks, permissions, `Locate("delta")` and the PATH check depend on it), so both are stated to keep a folder rename from changing it.
- The version is injected with `-X main.version={{ .Version }}`, so `delta version` prints `delta (deltas-ai) 0.1.0`. Keep that prefix, because other code relies on it.
- `-s -w -trimpath` shrinks the binary and removes local paths.
- `prerelease: auto` marks tags such as `v0.2.0-rc1` as pre-releases by itself.
- The changelog is built from commit messages and leaves out `docs`, `chore` and `test`.
- `go vet` runs before the build as a last guard. The full test suite belongs to `ci.yml`.

Both installers read the same environment settings:

| Variable | Meaning |
|---|---|
| `DELTA_VERSION` | Version to install, without the `v`. Default: the latest stable release |
| `DELTA_INSTALL_DIR` | Where to put the binary. Default: `~/.local/bin`, or `%LOCALAPPDATA%\Programs\delta` on Windows |
| `DELTA_BASE_URL` | Where the archives and `checksums.txt` are. For testing against a local folder |
| `DELTA_SKIP_PATH` | `install.ps1` only. Set to `1` to leave the user `PATH` alone, for testing |

The installers use GitHub's `releases/latest`, which **ignores pre-releases**. Until a stable release exists they stop with a message that tells the user to set `DELTA_VERSION`.

---

## 3. Cutting a release

1. **Check that `main` is ready.** The latest `ci` run on `main` must be green, and the commit you tag must be the one that passed. The `release-check` job of `ci` already runs `goreleaser check` and a snapshot build on every push and pull request, so a broken release config shows up there and not on a tag.
2. **Optionally dry-run the build by hand.** `goreleaser check` validates the config, and `goreleaser release --snapshot --clean` builds everything into `dist/` without publishing. Run `dist/delta_windows_amd64_v1/delta.exe version` (or the build for your system) and confirm the version is injected.
3. **Try a release candidate first when the pipeline changed.**

   ```sh
   git tag -a v0.2.0-rc1 -m "v0.2.0-rc1"
   git push origin v0.2.0-rc1
   ```

   It is published as a pre-release. Install it with `DELTA_VERSION=0.2.0-rc1`.
4. **Tag the release.**

   ```sh
   git switch main
   git pull
   git tag -a v0.2.0 -m "v0.2.0"
   git push origin v0.2.0
   ```

5. **Follow the `release` workflow** in the Actions tab. When it ends, the Releases page has the six archives, `checksums.txt` and a changelog.
6. **Test the installer without `DELTA_VERSION`** on a machine that is not yours. It must pick the new version, and a second run must replace the binary without errors.
7. **Optionally edit the release notes** to add a human summary on top.

Version numbers follow semantic versioning with a leading `v`. Stay on `v0.x.y` while the contract format and the commands may still change.

**If a release goes wrong**, delete the release on GitHub and the tag (`git push --delete origin vX.Y.Z`, `git tag -d vX.Y.Z`), fix the problem and tag again. Do this before people have installed it, or move on to the next patch version instead.

### Testing the installers without a release

Build a snapshot, then point an installer at the `dist/` folder:

```powershell
$env:DELTA_VERSION = "<snapshot version, as in the archive names>"
$env:DELTA_BASE_URL = "<path to the dist folder>"
$env:DELTA_INSTALL_DIR = "<a throwaway folder>"
$env:DELTA_SKIP_PATH = "1"
./install.ps1
```

For `install.sh`, `DELTA_BASE_URL` takes a `file://` URL. To check the safety net, copy the archive and a `checksums.txt` with one digit changed to another folder and install from there: it must refuse.

---

## 4. Decisions taken

| Question | Decision |
|---|---|
| Is the repository public? | Yes. The installers download without a token, which only works for a public repository |
| License | MIT, in `LICENSE`. The rules and skills are Delta's own work, with no third-party code to credit |
| macOS | Built and published, but not promoted as a main target |
| Code signing | Not for now. An unsigned binary may show a Windows SmartScreen or macOS Gatekeeper warning when downloaded with a browser; the installers do not trigger it. Signing needs paid certificates, and the manual mentions it |
| Where the installers are served from | The `main` branch. A user always gets the newest script, which picks the release itself |
| Entrypoint folder | `cmd/delta/`, so `go install github.com/ronz204/deltas-ai/cmd/delta@latest` also yields `delta` |

---

## 5. What is verified, and what is not

**Verified**
- The snapshot build produces six archives and `checksums.txt`, and the injected version shows in `delta version`.
- `install.ps1` installs, rejects a wrong checksum without installing anything, replaces the binary on a second run and adds the folder to the user `PATH`.
- `v0.1.0` was published by the `release` workflow and installs on Windows.

**Not verified**
- `install.sh` on a real Linux or macOS system. It was exercised with a simulated `uname`, which covers download, checksum, extraction and install, but the final run of the Linux binary was not.
- The `darwin` and `linux/arm64` binaries have not been run. They build, but nobody has started them.

---

## 6. Pending and ideas

- **Test `install.sh` on Linux.** A container or a VM is enough, with the one-line command and no `DELTA_VERSION`.
- **Package managers.** A Scoop bucket for Windows and a Homebrew tap for Linux and macOS can be added to `.goreleaser.yaml` once people use the tool. Each needs a separate repository for its manifest.
- **Code signing**, if the warnings turn out to put people off.
