$ErrorActionPreference = 'Stop'
$arch = if ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture -eq 'Arm64') { 'arm64' } else { 'amd64' }
$urls = @{ amd64 = 'https://github.com/akash-kamat/switchboard/releases/download/v0.1.0/switchboard_windows_amd64.zip'; arm64 = 'https://github.com/akash-kamat/switchboard/releases/download/v0.1.0/switchboard_windows_arm64.zip' }
$hashes = @{ amd64 = 'db7c3d139e24957a34f255535cb06159ba152b70c4604887891e532d432724ce'; arm64 = '77bb367ce60aadb9d21d40c6a147e4f62f84898cf4e73ef4221e6a7780c9e7b2' }
Install-ChocolateyZipPackage -PackageName 'switchboard' -Url64bit $urls[$arch] -Checksum64 $hashes[$arch] -ChecksumType64 'sha256' -UnzipLocation $toolsDir
