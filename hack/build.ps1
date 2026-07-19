# PowerShell port of hack/build.sh
# Build script for go-musicfox on Windows

param(
    [Parameter(Position = 0)]
    [string]$Action = "build",

    [string]$GoBinary = "go",

    [string]$Goos = "",
    [string]$Goarch = "",

    [string]$Ldflags = "-s -w",
    [string]$InjectPackage = "github.com/go-musicfox/go-musicfox/internal/types",

    [string]$LastfmKey = "",
    [string]$LastfmSecret = "",

    [string]$BuildTarget = "",
    [string]$BuildOutput = "",
    [string]$BuildTags = ""
)

$ErrorActionPreference = "Stop"

$root = Split-Path -Parent (Split-Path -Parent $PSCommandPath)

if (-not $Goos) { $Goos = & $GoBinary env GOOS 2>$null }
if (-not $Goarch) { $Goarch = & $GoBinary env GOARCH 2>$null }
if (-not $BuildTags) { $BuildTags = $env:BUILD_TAGS }

if ($Action -eq "build") {
    $BuildOutput = "${root}\bin\musicfox.exe"
}

# Get version info from version.ps1
$versionInfo = & "${root}\hack\version.ps1"

foreach ($line in ($versionInfo -split "`n")) {
    if ($line -match '^(\S+)\s+(.+)$') {
        $key = $matches[1]
        $value = $matches[2]
        $Ldflags = "${Ldflags} -X ${InjectPackage}.${key}=${value}"
    }
}

$Ldflags = "${Ldflags} -X ${InjectPackage}.LastfmKey=${LastfmKey}"
$Ldflags = "${Ldflags} -X ${InjectPackage}.LastfmSecret=${LastfmSecret}"
$Ldflags = "${Ldflags} -X ${InjectPackage}.BuildTags=${BuildTags}"

# Build
$env:CGO_ENABLED = "1"
$env:GOOS = $Goos
$env:GOARCH = $Goarch

Write-Host "Building with: CGO_ENABLED=1 GOOS=${Goos} GOARCH=${Goarch}"
Write-Host "  Action: ${Action}"
Write-Host "  Tags: ${BuildTags}"
Write-Host "  Ldflags: ${Ldflags}"
Write-Host "  Output: ${BuildOutput}"

# Use array splatting for native command to ensure correct argument passing
$goArgs = @(
    $Action
    "-tags=${BuildTags}"
    "-ldflags=${Ldflags}"
)
if ($BuildOutput) {
    $goArgs += @('-o', $BuildOutput)
}
if ($BuildTarget) {
    $goArgs += $BuildTarget
}
$goArgs += "$root\cmd"

& $GoBinary @goArgs
if ($LASTEXITCODE -ne 0) {
    throw "Build failed with exit code ${LASTEXITCODE}."
}

Write-Host "Build succeeded."
