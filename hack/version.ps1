# get version info from git
# PowerShell port of hack/version.sh

param(
    [string]$GitTag = "",
    [string]$GitRevision = ""
)

$Script:RemoteInfo = ""

function Fetch-RemoteGitInfo {
    if (-not $Script:RemoteInfo) {
        $Script:RemoteInfo = git ls-remote --tags https://github.com/go-musicfox/go-musicfox
    }
}

if (-not $GitTag) {
    $tagResult = git describe --tags --always --dirty=-dev 2>$null
    if ($LASTEXITCODE -eq 0) {
        $GitTag = $tagResult
    }
    else {
        Fetch-RemoteGitInfo
        $lines = @($Script:RemoteInfo -split "`n" | Where-Object { $_ -ne "" })
        if ($lines.Count -gt 0) {
            $lastLine = $lines[-1]
            if ($lastLine -match 'refs/tags/(.+)$') {
                $GitTag = $matches[1]
            }
        }
    }
}

if (-not $GitRevision) {
    $revResult = git rev-parse HEAD 2>$null
    if ($LASTEXITCODE -eq 0) {
        $GitRevision = $revResult
        git diff-index --quiet HEAD -- 2>$null
        if ($LASTEXITCODE -ne 0) {
            $GitRevision = "${GitRevision}-dev"
        }
    }
    else {
        Fetch-RemoteGitInfo
        foreach ($line in ($Script:RemoteInfo -split "`n")) {
            if ($line -match "^(\S+)\s+refs/tags/$([regex]::Escape($GitTag))$") {
                $GitRevision = $matches[1]
                break
            }
        }
    }
}

$version = "0.0.0-${GitRevision}"
if ($GitTag) {
    $version = $GitTag
}

Write-Output "AppVersion $version"
Write-Output "GitRevision $GitRevision"
Write-Output "User $(whoami)"
try {
    $hostEntry = [System.Net.Dns]::GetHostEntry('')
    Write-Output "Host $($hostEntry.HostName)"
}
catch {
    Write-Output "Host $env:COMPUTERNAME"
}
Write-Output "Time $(Get-Date -Format 'yyyy-MM-ddTHH:mm:ss')"
