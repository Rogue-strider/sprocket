# sprocket installer — native Windows PowerShell.
#
# Stages the prebuilt windows-amd64 binary as bin\sprocket-{learn,validate,hook}.exe
# and registers the plugin. No Go toolchain required.

$ErrorActionPreference = "Stop"
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$BinDir = Join-Path $ScriptDir "bin"

$arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else {
    Write-Error "Unsupported architecture: only 64-bit Windows is supported."
    exit 1
}
$platform = "windows-$arch"
Write-Host "==> Detected platform: $platform"

foreach ($cmd in @("sprocket-learn", "sprocket-validate", "sprocket-hook")) {
    $src = Join-Path $BinDir "$cmd-$platform.exe"
    $dest = Join-Path $BinDir "$cmd.exe"
    if (-not (Test-Path $src)) {
        Write-Error "No prebuilt binary for $platform ($src missing). If you have Go installed: go build -o $dest ./cmd/$cmd"
        exit 1
    }
    Copy-Item -Path $src -Destination $dest -Force
    Write-Host "==> Staged $dest"
}

if (-not (Get-Command claude -ErrorAction SilentlyContinue)) {
    Write-Error "Claude Code CLI not found on PATH. Install it first: https://docs.claude.com/en/docs/claude-code"
    exit 1
}

Write-Host "==> Registering local marketplace at $ScriptDir"
claude plugin marketplace add "$ScriptDir"

Write-Host "==> Installing sprocket plugin"
claude plugin install sprocket@sprocket-marketplace

Write-Host ""
Write-Host "Installed. Restart Claude Code (or start a new session) in any project to pick it up."
Write-Host "Try: /plan, /review, /fix-build, /learn"
