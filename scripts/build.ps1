param(
    [string[]]$Archs = @("x86_64", "arm_64", "arm-x41")
)

$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot
Set-Location $Root

Write-Host "==> Generating icons..."
go run ./scripts/gen-icons
if ($LASTEXITCODE -ne 0) { throw "gen-icons failed" }

Write-Host "==> Running Go unit tests..."
go test ./...
if ($LASTEXITCODE -ne 0) { throw "go test failed" }

$archMap = @{
    "x86_64"  = @{ GOARCH = "amd64"; GOARM = "" }
    "arm_64"  = @{ GOARCH = "arm64"; GOARM = "" }
    "arm-x41" = @{ GOARCH = "arm";   GOARM = "7" }
}

foreach ($k in $archMap.Keys) {
    $dir = Join-Path $Root "qpkg\$k"
    if (Test-Path $dir) { Remove-Item -Recurse -Force $dir }
}

$distDir = Join-Path $Root "dist"
New-Item -ItemType Directory -Force -Path $distDir | Out-Null

foreach ($arch in $Archs) {
    if (-not $archMap.ContainsKey($arch)) {
        throw "Unsupported QPKG arch: $arch"
    }
    $meta = $archMap[$arch]
    $outBin = Join-Path $Root "qpkg\$arch\www\api.cgi"
    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $outBin) | Out-Null

    Write-Host "==> Building api.cgi for $arch (GOARCH=$($meta.GOARCH))..."
    $env:CGO_ENABLED = "0"
    $env:GOOS = "linux"
    $env:GOARCH = $meta.GOARCH
    if ($meta.GOARM) { $env:GOARM = $meta.GOARM } else { Remove-Item Env:GOARM -ErrorAction SilentlyContinue }
    go build -trimpath -ldflags="-s -w" -o $outBin ./cmd/beszel-qnap
    if ($LASTEXITCODE -ne 0) { throw "go build failed for $arch" }

    Remove-Item Env:CGO_ENABLED, Env:GOOS, Env:GOARCH, Env:GOARM -ErrorAction SilentlyContinue

    $bundledSmartctl = Join-Path $Root "assets\smartctl\$arch\smartctl"
    if (Test-Path $bundledSmartctl) {
        $smartctlDest = Join-Path $Root "qpkg\$arch\bin\smartctl"
        New-Item -ItemType Directory -Force -Path (Split-Path -Parent $smartctlDest) | Out-Null
        Copy-Item $bundledSmartctl $smartctlDest -Force
        Write-Host "==> Bundled static smartctl for $arch"
    }

    Write-Host "==> Packing QPKG for $arch..."
    go run ./scripts/qpkg-pack -qpkg-dir qpkg -out-dir dist -arch $arch
    if ($LASTEXITCODE -ne 0) { throw "qpkg-pack failed for $arch" }
}

Write-Host "==> Built packages:"
Get-ChildItem $distDir -Filter *.qpkg | Format-Table Name, Length, LastWriteTime
