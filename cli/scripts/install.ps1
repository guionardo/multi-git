# Usage (PowerShell):
#   irm https://raw.githubusercontent.com/guionardo/multi-git/main/cli/scripts/install.ps1 | iex
#   $env:MULTI_GIT_ARGS="apply --dry-run"; irm ... | iex
param()

$ErrorActionPreference = "Stop"
$Repo = if ($env:MULTI_GIT_REPO) { $env:MULTI_GIT_REPO } else { "guionardo/multi-git" }
$Version = if ($env:MULTI_GIT_VERSION) { $env:MULTI_GIT_VERSION } else { "latest" }
$Base = "https://github.com/$Repo/releases/$Version/download"

$Arch = if ([Environment]::Is64BitOperatingSystem) {
    if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { "arm64" } else { "amd64" }
} else { "386" }
$Bin = "multi-git_windows_$Arch.exe"
$Url = "$Base/$Bin"

$Tmp = New-Item -ItemType Directory -Path (Join-Path $env:TEMP "multi-git") -Force
$Out = Join-Path $Tmp.FullName "multi-git.exe"
Write-Host "Downloading $Url ..."
Invoke-WebRequest -Uri $Url -OutFile $Out

$Args = if ($env:MULTI_GIT_ARGS) { $env:MULTI_GIT_ARGS.Split(" ") } else { @() }
& $Out @Args