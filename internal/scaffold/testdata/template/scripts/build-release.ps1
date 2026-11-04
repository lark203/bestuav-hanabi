# Builds signed Windows installers and the update manifest.
# With -Upload, uploads them to the GitHub draft release of this version.
#
# The update private key never enters the repository. Pass it in
# MYGO_UPDATER_PRIVATE_KEY, or let the script read the key that
# `go tool mygo keygen` wrote outside the checkout.
param(
    [string]$KeyPath = "",
    [switch]$SkipInstaller,
    [switch]$Upload
)

$ErrorActionPreference = "Stop"

if (-not $env:MYGO_UPDATER_PRIVATE_KEY) {
    if (-not $KeyPath) {
        $KeyPath = Join-Path (Split-Path $PSScriptRoot -Parent) "..\bestuav-hanabi-update-keys\mygo-update.key"
    }
    if (-not (Test-Path -LiteralPath $KeyPath)) {
        throw "Update key not found at $KeyPath. Set MYGO_UPDATER_PRIVATE_KEY or pass -KeyPath."
    }
    $env:MYGO_UPDATER_PRIVATE_KEY = (Get-Content -Raw -LiteralPath $KeyPath).Trim()
}

$buildArgs = @("build")
if ($SkipInstaller) { $buildArgs += "-skip-installer" }
if ($Upload) {
    if (-not $env:GH_TOKEN) {
        throw "GH_TOKEN is required for -Upload (GitHub token with repo release permission)."
    }
    $buildArgs += "-upload"
}
& go tool mygo @buildArgs
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host ""
Write-Host "Artifacts:"
Get-ChildItem -LiteralPath (Join-Path $PSScriptRoot "..\build\windows-amd64") |
    Select-Object Name, Length |
    Format-Table -AutoSize
