param (
    [switch]$Clean,
    [switch]$Test,
    [switch]$Rebuild
)

$ErrorActionPreference = "Stop"

$ProjRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
$ResDir   = Join-Path $ProjRoot "resources"
$OutExe   = Join-Path $ProjRoot "gclock.exe"
$CmdDir   = Join-Path $ProjRoot "cmd\alarmclock"

$MSysRoot = "C:\msys64\ucrt64"
$WrPath   = Join-Path $MSysRoot "bin\windres.exe"

$env:PATH = (Join-Path $MSysRoot "bin") + ";" + $env:PATH

if ($Clean) {
    Write-Host "Cleaning build artifacts..."
    if (Test-Path $OutExe) { Remove-Item -Force $OutExe }
    $Syso = Join-Path $CmdDir "app.syso"
    if (Test-Path $Syso) { Remove-Item -Force $Syso }
    Write-Host "Clean complete."
    exit 0
}

if (-not (Test-Path $WrPath)) {
    Write-Error "windres.exe not found at $WrPath. Please install MSYS2 UCRT64 toolchain."
    exit 1
}

$IconFile = Join-Path $ResDir "alarmclock.ico"
if (-not (Test-Path $IconFile)) {
    Write-Host "Generating app icon..." -ForegroundColor Cyan
    $GenScript = Join-Path $ProjRoot "generate_icon.ps1"
    powershell -ExecutionPolicy Bypass -File $GenScript
}

$RcFile = Join-Path $ResDir "app.rc"
$SysoFile = Join-Path $CmdDir "app.syso"

if (-not (Test-Path $CmdDir)) {
    New-Item -ItemType Directory -Path $CmdDir -Force | Out-Null
}

Write-Host "Compiling resources..." -ForegroundColor Cyan
& $WrPath @("-i", $RcFile, "-o", $SysoFile, "-I$ResDir")
if ($LASTEXITCODE -ne 0) { Write-Error "windres failed"; exit 1 }

Write-Host "Building Go application..." -ForegroundColor Cyan
Push-Location $ProjRoot
go build -ldflags="-H windowsgui -s -w" -o $OutExe ./cmd/alarmclock
if ($LASTEXITCODE -ne 0) { Write-Error "go build failed"; exit 1 }
Pop-Location

Write-Host "Build successful: $OutExe" -ForegroundColor Green

if ($Test) {
    Write-Host "Running tests..." -ForegroundColor Cyan
    Push-Location $ProjRoot
    go test ./...
    if ($LASTEXITCODE -ne 0) { Write-Error "tests failed"; exit 1 }
    Pop-Location
}
