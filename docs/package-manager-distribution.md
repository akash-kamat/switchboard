# Package-manager distribution

GitHub Releases are the source of truth. After a release is published, the
distribution workflow verifies its checksum manifest and updates ready-to-submit
definitions under `distribution/` for Homebrew, Scoop, Winget, Chocolatey, and
the AUR.

The maintained Homebrew tap and Scoop bucket are live. Winget, Chocolatey, and
AUR publication still requires maintainer accounts or upstream review; apt and
yum/dnf repositories additionally require an offline-controlled GPG signing key.
Tokens and private keys must be stored as protected secrets, not committed to
this repository.

## Homebrew

```sh
brew tap akash-kamat/switchboard
brew install switchboard
brew services start switchboard
```

Use `brew update && brew upgrade switchboard` to upgrade. Stop and remove it with
`brew services stop switchboard && brew uninstall switchboard`. The tap is
published at <https://github.com/akash-kamat/homebrew-switchboard> and its formula
is install-tested on macOS by a separate workflow.

## Scoop

```powershell
scoop bucket add switchboard https://github.com/akash-kamat/scoop-switchboard
scoop install switchboard
```

Use `scoop update switchboard` to upgrade and `scoop uninstall switchboard` to
remove it. Scoop installs the portable CLI; use `install.ps1` when Windows Service
registration is wanted. The bucket has a separate Windows installation test.

## Maintainer release procedure

1. Update `CHANGELOG.md`, run all CI workflows, and tag `vMAJOR.MINOR.PATCH`.
2. Confirm the Release workflow publishes checksums, SBOM, and Sigstore bundles.
3. Confirm the distribution workflow commits manifests containing no placeholders.
4. Sync the Homebrew formula and Scoop manifest to their dedicated tap/bucket;
   both repositories run independent installation checks.
5. Submit the generated Winget manifest to `microsoft/winget-pkgs`, the Chocolatey
   package to Chocolatey Community Repository, and the `PKGBUILD` to the AUR.
6. Publish signed apt/rpm metadata from the exact `.deb`/`.rpm` release assets.
7. Test install, upgrade, and uninstall through every public channel before
   announcing it.

External publication is intentionally never performed merely by pushing source
code. This prevents an unreviewed commit or pull request from publishing packages.
