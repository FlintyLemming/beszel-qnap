#!/usr/bin/env bash
set -euo pipefail

SMARTMONTOOLS_VERSION="7.5"
SMARTMONTOOLS_URL="https://github.com/smartmontools/smartmontools/releases/download/RELEASE_7_5/smartmontools-${SMARTMONTOOLS_VERSION}.tar.gz"
SMARTMONTOOLS_SHA256="690b83ca331378da9ea0d9d61008c4b22dde391387b9bbad7f29387f2595f76e"

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

FORCE=0
ARCHS=()
for arg in "$@"; do
  case "$arg" in
    --force)
      FORCE=1
      ;;
    *)
      ARCHS+=("$arg")
      ;;
  esac
done

if [ "${#ARCHS[@]}" -eq 0 ]; then
  ARCHS=("x86_64" "arm_64" "arm-x41")
fi

TARGETS=()
for arch in "${ARCHS[@]}"; do
  case "$arch" in
    x86_64|arm_64|arm-x41) ;;
    *)
      echo "Unsupported QPKG arch: $arch" >&2
      exit 1
      ;;
  esac
  out_bin="assets/smartctl/${arch}/smartctl"
  if [ "$FORCE" -eq 0 ] && [ -f "$out_bin" ]; then
    echo "==> Skipping ${arch}: ${out_bin} already exists (use --force to rebuild)"
  else
    TARGETS+=("$arch")
  fi
done

if [ "${#TARGETS[@]}" -eq 0 ]; then
  echo "==> All requested smartctl binaries already exist."
  exit 0
fi

WORK_DIR="$(mktemp -d)"
trap 'rm -rf "$WORK_DIR"' EXIT

TARBALL="${WORK_DIR}/smartmontools-${SMARTMONTOOLS_VERSION}.tar.gz"
echo "==> Downloading smartmontools ${SMARTMONTOOLS_VERSION}..."
curl -fsSL --retry 3 --retry-delay 2 -o "$TARBALL" "$SMARTMONTOOLS_URL"
echo "${SMARTMONTOOLS_SHA256}  ${TARBALL}" | sha256sum -c -

JOBS="$(nproc 2>/dev/null || echo 2)"

for arch in "${TARGETS[@]}"; do
  case "$arch" in
    x86_64)
      host_triplet=""
      strip_bin="strip"
      readelf_bin="readelf"
      expected_machine="Advanced Micro Devices X86-64"
      ;;
    arm_64)
      host_triplet="aarch64-linux-gnu"
      strip_bin="aarch64-linux-gnu-strip"
      readelf_bin="aarch64-linux-gnu-readelf"
      expected_machine="AArch64"
      ;;
    arm-x41)
      host_triplet="arm-linux-gnueabihf"
      strip_bin="arm-linux-gnueabihf-strip"
      readelf_bin="arm-linux-gnueabihf-readelf"
      expected_machine="ARM"
      ;;
  esac

  echo "==> Building static smartctl ${SMARTMONTOOLS_VERSION} for ${arch} (${host_triplet:-native})..."
  src_dir="${WORK_DIR}/src-${arch}"
  mkdir -p "$src_dir"
  tar -xzf "$TARBALL" -C "$src_dir" --strip-components=1

  pushd "$src_dir" >/dev/null
  configure_args=(
    "--without-libcap-ng"
    "--without-libsystemd"
    "--without-selinux"
    "--with-nvme-devicescan=yes"
    "CFLAGS=-O2"
    "CXXFLAGS=-O2"
    "LDFLAGS=-static"
  )
  if [ -n "$host_triplet" ]; then
    configure_args=("--host=${host_triplet}" "${configure_args[@]}")
  fi

  ./configure "${configure_args[@]}"
  make -j"$JOBS" smartctl
  "$strip_bin" smartctl

  if "$readelf_bin" -l smartctl | grep -q 'INTERP'; then
    echo "ERROR: ${arch} smartctl is dynamically linked (PT_INTERP found)" >&2
    exit 1
  fi
  if ! "$readelf_bin" -h smartctl | grep -q "Machine:.*${expected_machine}"; then
    echo "ERROR: unexpected ELF machine for ${arch} (expected ${expected_machine})" >&2
    "$readelf_bin" -h smartctl >&2
    exit 1
  fi

  out_bin="${ROOT}/assets/smartctl/${arch}/smartctl"
  mkdir -p "$(dirname "$out_bin")"
  cp smartctl "${out_bin}.tmp"
  chmod 0755 "${out_bin}.tmp"
  mv "${out_bin}.tmp" "$out_bin"
  popd >/dev/null

  ls -lh "$out_bin"
done
