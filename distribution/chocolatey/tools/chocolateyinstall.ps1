$ErrorActionPreference = 'Stop'
$arch = if ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture -eq 'Arm64') { 'arm64' } else { 'amd64' }
$urls = @{ amd64 = 'https://github.com/akash-kamat/switchboard/releases/download/v0.1.2/switchboard_windows_amd64.zip'; arm64 = 'https://github.com/akash-kamat/switchboard/releases/download/v0.1.2/switchboard_windows_arm64.zip' }
$hashes = @{ amd64 = 'ead7d2800038e8ada8c88da05e20f9aa3b987549bc526f6e6f03b7c3433941e2'; arm64 = '986825cb20fe83f76563ec9a68c360b1bc3690354f0a56d1f3f170ca23b2a412' }
Install-ChocolateyZipPackage -PackageName 'switchboard' -Url64bit $urls[$arch] -Checksum64 $hashes[$arch] -ChecksumType64 'sha256' -UnzipLocation $toolsDir
