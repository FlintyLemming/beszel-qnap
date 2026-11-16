#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

if [ "$#" -gt 0 ]; then
  ARCHS=("$@")
else
  ARCHS=("x86_64" "arm_64" "arm-x41")
fi

echo "==> Generating icons..."
go run ./scripts/gen-icons

echo "==> Running Go unit tests..."
go test ./...

rm -rf qpkg/x86_64 qpkg/arm_64 qpkg/arm-x41
mkdir -p dist

for arch in "${ARCHS[@]}"; do
  case "$arch" in
    x86_64)
      goarch="amd64"
      goarm=""
      ;;
    arm_64)
      goarch="arm64"
      goarm=""
      ;;
    arm-x41)
      goarch="arm"
      goarm="7"
      ;;
    *)
      echo "Unsupported QPKG arch: $arch" >&2
      exit 1
      ;;
  esac

  out_bin="qpkg/${arch}/www/api.cgi"
  mkdir -p "$(dirname "$out_bin")"

  echo "==> Building api.cgi for ${arch} (GOARCH=${goarch})..."
  if [ -n "$goarm" ]; then
    CGO_ENABLED=0 GOOS=linux GOARCH="$goarch" GOARM="$goarm" go build -trimpath -ldflags="-s -w" -o "$out_bin" ./cmd/beszel-qnap
  else
    CGO_ENABLED=0 GOOS=linux GOARCH="$goarch" go build -trimpath -ldflags="-s -w" -o "$out_bin" ./cmd/beszel-qnap
  fi

  bundled_smartctl="assets/smartctl/${arch}/smartctl"
  if [ -f "$bundled_smartctl" ]; then
    mkdir -p "qpkg/${arch}/bin"
    cp "$bundled_smartctl" "qpkg/${arch}/bin/smartctl"
    chmod 0755 "qpkg/${arch}/bin/smartctl"
    echo "==> Bundled static smartctl for ${arch}"
  fi

  echo "==> Packing QPKG for ${arch}..."
  go run ./scripts/qpkg-pack -qpkg-dir qpkg -out-dir dist -arch "$arch"
done

echo "==> Built packages:"
ls -lh dist/*.qpkg
