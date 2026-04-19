$ErrorActionPreference = "Stop"

$OutputDir = "build"

# Clean up and ensure build directories exist
if (Test-Path $OutputDir) {
    Remove-Item -Recurse -Force $OutputDir
}
New-Item -ItemType Directory -Path "$OutputDir/windows" | Out-Null
New-Item -ItemType Directory -Path "$OutputDir/linux" | Out-Null

Write-Host "Compiling Windows binaries (amd64)..." -ForegroundColor Cyan
$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build -o "$OutputDir/windows/agent.exe" ./cmd/agent
if ($LASTEXITCODE -ne 0) { throw "Agent Windows build failed" }

go build -o "$OutputDir/windows/server.exe" ./cmd/server
if ($LASTEXITCODE -ne 0) { throw "Server Windows build failed" }
Write-Host "  > Windows build complete.`n" -ForegroundColor Green


Write-Host "Compiling Linux binaries (amd64)..." -ForegroundColor Cyan
$env:GOOS = "linux"
$env:GOARCH = "amd64"
go build -o "$OutputDir/linux/agent" ./cmd/agent
if ($LASTEXITCODE -ne 0) { throw "Agent Linux build failed" }

go build -o "$OutputDir/linux/server" ./cmd/server
if ($LASTEXITCODE -ne 0) { throw "Server Linux build failed" }
Write-Host "  > Linux build complete.`n" -ForegroundColor Green

Write-Host "All builds finished successfully! Binaries are located in '$OutputDir'." -ForegroundColor Magenta
