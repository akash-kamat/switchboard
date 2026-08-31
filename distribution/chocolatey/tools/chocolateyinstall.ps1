$ErrorActionPreference = 'Stop'
$arch = if ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture -eq 'Arm64') { 'arm64' } else { 'amd64' }
$urls = @{ amd64 = 'https://github.com/akash-kamat/switchboard/releases/download/v0.1.1/switchboard_windows_amd64.zip'; arm64 = 'https://github.com/akash-kamat/switchboard/releases/download/v0.1.1/switchboard_windows_arm64.zip' }
$hashes = @{ amd64 = '8b86e6f935fbd53e754d7bc1976c4b02501f68d3c7479650be136c4efb111e40'; arm64 = 'ee480bcc2282bbbd75898b32d3a24fc4ecd14dc851943949f094da3ed1b7cf20' }
Install-ChocolateyZipPackage -PackageName 'switchboard' -Url64bit $urls[$arch] -Checksum64 $hashes[$arch] -ChecksumType64 'sha256' -UnzipLocation $toolsDir
