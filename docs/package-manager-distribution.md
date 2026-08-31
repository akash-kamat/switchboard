# Package-manager distribution

GitHub Releases are the source of truth. After a release is published, the
distribution workflow verifies its checksum manifest and updates ready-to-submit
definitions under `distribution/` for Homebrew, Scoop, Winget, Chocolatey, and
the AUR.

Until each external registry accepts the package, use the verified one-line
installer or a native Linux package from GitHub Releases. Registry publication is
an external operation: Winget, Chocolatey, and AUR require maintainer accounts or
review; apt and yum/dnf repositories additionally require an offline-controlled
GPG signing key. Tokens and private keys must be stored as protected secrets, not
committed to this repository.

## Maintainer release procedure

1. Update `CHANGELOG.md`, run all CI workflows, and tag `vMAJOR.MINOR.PATCH`.
2. Confirm the Release workflow publishes checksums, SBOM, and Sigstore bundles.
3. Confirm the distribution workflow commits manifests containing no placeholders.
4. Sync the Homebrew formula and Scoop manifest to their dedicated tap/bucket.
5. Submit the generated Winget manifest to `microsoft/winget-pkgs`, the Chocolatey
   package to Chocolatey Community Repository, and the `PKGBUILD` to the AUR.
6. Publish signed apt/rpm metadata from the exact `.deb`/`.rpm` release assets.
7. Test install, upgrade, and uninstall through every public channel before
   announcing it.

External publication is intentionally never performed merely by pushing source
code. This prevents an unreviewed commit or pull request from publishing packages.
