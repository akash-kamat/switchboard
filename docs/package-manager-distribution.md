# Package-manager distribution

GitHub Releases are the source of truth. After a release is published, the
distribution workflow verifies its checksum manifest and updates ready-to-submit
definitions under `distribution/` for Homebrew, Scoop, Winget, Chocolatey, and
the AUR.

No external package-manager repository is currently maintained. Homebrew,
Scoop, Winget, Chocolatey, and AUR definitions remain generated under
`distribution/` so publication can be resumed later. Registry publication
requires a separate repository or upstream review; apt and yum/dnf repositories
also require an offline-controlled GPG signing key.

## Homebrew

The generated formula is stored at
`distribution/homebrew/Formula/switchboard.rb`. It is not currently published
as a tap. Use the macOS shell installer from the main repository instead.

## Scoop

The generated manifest is stored at `distribution/scoop/switchboard.json`. It is
not currently published as a bucket. Use `install.ps1` from the main repository
for Windows Service registration.

## Maintainer release procedure

1. Update `CHANGELOG.md`, run all CI workflows, and tag `vMAJOR.MINOR.PATCH`.
2. Confirm the Release workflow publishes checksums, SBOM, and Sigstore bundles.
3. Confirm the distribution workflow commits manifests containing no placeholders.
4. If external repositories are introduced again, sync the Homebrew formula and
   Scoop manifest to their dedicated tap/bucket and add installation checks.
5. Submit the generated Winget manifest to `microsoft/winget-pkgs`, the Chocolatey
   package to Chocolatey Community Repository, and the `PKGBUILD` to the AUR.
6. Publish signed apt/rpm metadata from the exact `.deb`/`.rpm` release assets.
7. Test install, upgrade, and uninstall through every public channel before
   announcing it.

External publication is intentionally never performed merely by pushing source
code. This prevents an unreviewed commit or pull request from publishing packages.
