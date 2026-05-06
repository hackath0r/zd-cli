<#
.SYNOPSIS
    Install zd-cli on Windows (PowerShell).

.DESCRIPTION
    Downloads the latest zd-cli release for windows/amd64 from GitHub,
    verifies the sha256 checksum, extracts the zip into the user's
    Programs directory (or a custom path via -Prefix), and creates a
    'ximr.exe' alongside 'zd.exe'.

.PARAMETER Version
    Pin a specific release tag (e.g. v0.1.0). Defaults to "latest".

.PARAMETER Prefix
    Install prefix. Defaults to "$env:LOCALAPPDATA\Programs\zd-cli".

.EXAMPLE
    iwr -useb https://raw.githubusercontent.com/hackath0r/zd-cli/main/scripts/install.ps1 | iex
#>
param(
    [string]$Version = $env:ZD_VERSION,
    [string]$Prefix = "$env:LOCALAPPDATA\Programs\zd-cli"
)

$ErrorActionPreference = 'Stop'
$Repo = 'hackath0r/zd-cli'
$Bin  = 'zd.exe'
$Alias = 'ximr.exe'

function Resolve-Version {
    if ($Version) { return $Version }
    $latest = Invoke-RestMethod "https://api.github.com/repos/$Repo/releases/latest"
    return $latest.tag_name
}

function Get-Arch {
    switch -regex ((Get-CimInstance Win32_Processor).Architecture) {
        9 { return 'amd64' }       # x64
        12 { return 'arm64' }      # ARM64
        Default { throw "unsupported architecture" }
    }
}

$resolvedVersion = Resolve-Version
if (-not $resolvedVersion) { throw "could not resolve version" }

$arch = Get-Arch
$cleanVersion = $resolvedVersion.TrimStart('v')
$archive = "zd_${cleanVersion}_windows_${arch}.zip"
$archiveUrl = "https://github.com/$Repo/releases/download/$resolvedVersion/$archive"
$checksumsUrl = "https://github.com/$Repo/releases/download/$resolvedVersion/checksums.txt"

Write-Host "Installing zd-cli $resolvedVersion for windows/$arch into $Prefix"

$tempDir = New-Item -ItemType Directory -Path "$env:TEMP\zd-cli-install-$([guid]::NewGuid())"
try {
    $archivePath = Join-Path $tempDir $archive
    $checksumsPath = Join-Path $tempDir 'checksums.txt'

    Write-Host "Downloading $archive"
    Invoke-WebRequest -Uri $archiveUrl -OutFile $archivePath -UseBasicParsing
    Invoke-WebRequest -Uri $checksumsUrl -OutFile $checksumsPath -UseBasicParsing

    $expected = (Get-Content $checksumsPath | Where-Object { $_ -match " $archive$" }) -split '\s+' | Select-Object -First 1
    if (-not $expected) { throw "checksum line not found for $archive" }
    $actual = (Get-FileHash -Algorithm SHA256 $archivePath).Hash.ToLower()
    if ($actual -ne $expected.ToLower()) {
        throw "checksum mismatch: expected $expected, got $actual"
    }

    Expand-Archive -Path $archivePath -DestinationPath $tempDir -Force

    $null = New-Item -ItemType Directory -Path "$Prefix\bin" -Force
    Copy-Item -Path (Join-Path $tempDir $Bin) -Destination "$Prefix\bin\$Bin" -Force
    Copy-Item -Path (Join-Path $tempDir $Bin) -Destination "$Prefix\bin\$Alias" -Force

    Write-Host "Installed: $Prefix\bin\$Bin and $Prefix\bin\$Alias"

    $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
    if ($userPath -notlike "*$Prefix\bin*") {
        Write-Host "Adding $Prefix\bin to user PATH"
        [Environment]::SetEnvironmentVariable('Path', "$userPath;$Prefix\bin", 'User')
        Write-Host "Open a new shell to pick up the PATH change."
    }

    & "$Prefix\bin\$Bin" version
    Write-Host "Next: run '$Bin config init' to set up a profile."
}
finally {
    Remove-Item -Recurse -Force $tempDir -ErrorAction SilentlyContinue
}
