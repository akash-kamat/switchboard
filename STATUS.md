# Switchboard status

Updated: 2026-08-31

## Done

- Built the Go dashboard, API, CLI, validated YAML editor, native service
  controls, Docker Engine integration, and native host metrics.
- Added Linux, Windows, and macOS service definitions and one-line installers.
- Published GitHub release `v0.1.2` with Linux, Windows, and macOS binaries.
- Published `.deb`, `.rpm`, and Arch Linux packages for supported Linux
  architectures.
- Added SHA-256 checksums, SBOMs, and Sigstore verification bundles.
- Made Linux installations run as root and Windows installations run as
  LocalSystem so service and Docker controls work without manual permissions.
- Added automatic fallback from port 8080 to the first free port in 8081-8099.
- Made fresh installations start with an empty service list.
- Tested Go code, configuration, cross-compilation, release packaging, Linux
  package lifecycle, and Linux/Windows/macOS installer lifecycle in GitHub
  Actions.
- Kept generated Homebrew, Scoop, Winget, Chocolatey, and AUR manifests inside
  this repository for possible future publication.

## Left to do

- Delete `akash-kamat/homebrew-switchboard` and
  `akash-kamat/scoop-switchboard`. The authenticated GitHub CLI token does not
  have the required `delete_repo` scope. Run
  `gh auth refresh -h github.com -s delete_repo`, then delete them with
  `gh repo delete akash-kamat/homebrew-switchboard --yes` and
  `gh repo delete akash-kamat/scoop-switchboard --yes`.
- Optionally publish the generated package definitions to official Homebrew,
  Scoop, Winget, Chocolatey, or AUR channels in the future.
- Optionally operate signed apt and yum/dnf repositories if normal repository
  upgrades are desired.

## Current installation channels

- GitHub Releases, including direct `.deb`, `.rpm`, Arch, ZIP, and tar downloads.
- Linux/macOS `install.sh` from this repository.
- Windows `install.ps1` from this repository.

The main repository is <https://github.com/akash-kamat/switchboard>.
