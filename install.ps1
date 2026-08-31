[CmdletBinding()]
param(
    [string]$Version = "latest",
    [switch]$Uninstall
)

$ErrorActionPreference = "Stop"
$repository = "akash-kamat/switchboard"
$installDirectory = Join-Path $env:ProgramFiles "Switchboard"
$dataDirectory = Join-Path $env:ProgramData "Switchboard"
$executable = Join-Path $installDirectory "switchboard.exe"

function Stop-SwitchboardService {
    $service = Get-Service Switchboard -ErrorAction SilentlyContinue
    if ($null -eq $service -or $service.Status -eq [System.ServiceProcess.ServiceControllerStatus]::Stopped) {
        return
    }
    Stop-Service Switchboard
    $service.WaitForStatus(
        [System.ServiceProcess.ServiceControllerStatus]::Stopped,
        [TimeSpan]::FromSeconds(30)
    )
}

$identity = [Security.Principal.WindowsIdentity]::GetCurrent()
$principal = [Security.Principal.WindowsPrincipal]::new($identity)
if (-not $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) {
    throw "Run PowerShell as Administrator."
}

if ($Uninstall) {
    Stop-SwitchboardService
    & sc.exe delete Switchboard | Out-Null
    if (Test-Path -LiteralPath $installDirectory) { Remove-Item -LiteralPath $installDirectory -Recurse -Force }
    Write-Host "Removed Switchboard; $dataDirectory was preserved."
    exit 0
}

$architecture = switch ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString()) {
    "X64" { "amd64" }
    "Arm64" { "arm64" }
    default { throw "Unsupported Windows architecture." }
}
$base = if ($Version -eq "latest") {
    "https://github.com/$repository/releases/latest/download"
} else {
    "https://github.com/$repository/releases/download/$Version"
}
$archive = "switchboard_windows_$architecture.zip"
$temporary = Join-Path ([IO.Path]::GetTempPath()) ("switchboard-" + [Guid]::NewGuid())
New-Item -ItemType Directory -Path $temporary | Out-Null
try {
    Invoke-WebRequest "$base/$archive" -OutFile (Join-Path $temporary $archive)
    Invoke-WebRequest "$base/checksums.txt" -OutFile (Join-Path $temporary "checksums.txt")
    $line = Get-Content (Join-Path $temporary "checksums.txt") | Where-Object { $_ -match "\s+$([regex]::Escape($archive))$" } | Select-Object -First 1
    if (-not $line) { throw "Checksum for $archive is missing." }
    $expected = ($line -split "\s+")[0].ToLowerInvariant()
    $actual = (Get-FileHash (Join-Path $temporary $archive) -Algorithm SHA256).Hash.ToLowerInvariant()
    if ($actual -ne $expected) { throw "Checksum verification failed." }
    Expand-Archive (Join-Path $temporary $archive) -DestinationPath $temporary -Force
    New-Item -ItemType Directory -Path $installDirectory -Force | Out-Null
    New-Item -ItemType Directory -Path $dataDirectory -Force | Out-Null
    Stop-SwitchboardService
    Copy-Item (Join-Path $temporary "switchboard.exe") $executable -Force
    $configPath = Join-Path $dataDirectory "config.yaml"
    if (-not (Test-Path -LiteralPath $configPath)) { Copy-Item (Join-Path $temporary "config.example.yaml") $configPath }
    & icacls.exe $dataDirectory /inheritance:r /grant:r "Administrators:(OI)(CI)F" "SYSTEM:(OI)(CI)F" | Out-Null
    & sc.exe query Switchboard *> $null
    if ($LASTEXITCODE -eq 0) {
        & sc.exe config Switchboard binPath= ('"' + $executable + '" service --config "' + $configPath + '"') start= auto obj= LocalSystem | Out-Null
    } else {
        & sc.exe create Switchboard binPath= ('"' + $executable + '" service --config "' + $configPath + '"') start= auto obj= LocalSystem DisplayName= "Switchboard" | Out-Null
        & sc.exe description Switchboard "Lightweight service and container dashboard" | Out-Null
    }
    Start-Service Switchboard
    Start-Sleep -Seconds 1
    Write-Host "Switchboard installed. Configuration was preserved if it already existed."
    $listenLine = Get-Content -LiteralPath $configPath | Where-Object { $_ -match '^\s*listen\s*:' } | Select-Object -First 1
    $listenAddress = if ($listenLine) { (($listenLine -split ':', 2)[1]).Trim().Trim('"', "'") } else { ':8080' }
    Write-Host "Listen address: $listenAddress (see $configPath if the default port was occupied)."
} finally {
    if (Test-Path -LiteralPath $temporary) { Remove-Item -LiteralPath $temporary -Recurse -Force }
}
