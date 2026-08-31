# Distribution manifests

This directory is populated from signed GitHub release artifacts by the
`Generate distribution manifests` workflow. It contains versioned Homebrew,
Scoop, Winget, Chocolatey, and AUR definitions whose SHA-256 values come directly
from the release checksum manifest.

The definitions can be submitted or synchronized to their public upstream
repositories after a release. Repository signing material for apt and yum/dnf is
never stored here; see [package-manager-distribution.md](../docs/package-manager-distribution.md).
