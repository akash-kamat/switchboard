# Releases and verification

Version tags matching `vMAJOR.MINOR.PATCH` run the complete test and cross-build
matrix before publishing. GoReleaser builds all supported binaries, Linux native
packages, release notes, and `checksums.txt`. The release workflow also publishes
a CycloneDX SBOM and keyless Sigstore bundles tied to the GitHub Actions identity.

## Verify a download

Download the artifact, `checksums.txt`, and `checksums.txt.sigstore.json` from the
same GitHub release. First verify the digest:

```sh
sha256sum --ignore-missing -c checksums.txt
```

Then install Cosign and verify the GitHub Actions identity:

```sh
cosign verify-blob \
  --bundle checksums.txt.sigstore.json \
  --certificate-identity-regexp '^https://github.com/akash-kamat/switchboard/.github/workflows/release.yml@refs/tags/v' \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com \
  checksums.txt
```

The one-line installers always verify the selected archive against the published
SHA-256 manifest before changing the machine. For the strongest supply-chain
check, verify the signed checksum manifest first and then run the inspected local
installer.

## One-line installation

Linux or macOS (runs as root because it installs a system service):

```sh
curl -fsSL https://raw.githubusercontent.com/akash-kamat/switchboard/main/install.sh | sudo sh
```

Windows, from an Administrator PowerShell prompt:

```powershell
irm https://raw.githubusercontent.com/akash-kamat/switchboard/main/install.ps1 | iex
```

For an exact version, download and inspect the script, then pass `--version
v1.2.3` on Unix or `-Version v1.2.3` on Windows. Uninstall with `--uninstall` or
`-Uninstall`; configuration and state are preserved.
