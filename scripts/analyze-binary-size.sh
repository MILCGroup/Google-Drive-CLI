#!/bin/bash
# Binary size analysis script

set -e

BINARY_NAME="gdrv"
BINARY_DIR="bin"

echo "=== Binary Size Analysis ==="
echo ""

# Build debug version
echo "Building debug version..."
go build \
    -ldflags "-X github.com/dl-alexandre/gdrv/pkg/version.Version=$(git describe --tags --always --dirty 2>/dev/null || echo dev)" \
    -o "${BINARY_DIR}/${BINARY_NAME}-debug" \
    ./cmd/gdrv

# Build optimized version
echo "Building optimized version..."
go build \
    -trimpath \
    -ldflags "-s -w -X github.com/dl-alexandre/gdrv/pkg/version.Version=$(git describe --tags --always --dirty 2>/dev/null || echo dev)" \
    -o "${BINARY_DIR}/${BINARY_NAME}-optimized" \
    ./cmd/gdrv

DEBUG_SIZE=$(stat -f%z "${BINARY_DIR}/${BINARY_NAME}-debug" 2>/dev/null || stat -c%s "${BINARY_DIR}/${BINARY_NAME}-debug")
OPTIMIZED_SIZE=$(stat -f%z "${BINARY_DIR}/${BINARY_NAME}-optimized" 2>/dev/null || stat -c%s "${BINARY_DIR}/${BINARY_NAME}-optimized")

echo ""
echo "=== Size Comparison ==="
printf "%-20s %10s\n" "Version" "Size"
printf "%-20s %10s\n" "--------" "----"
printf "%-20s %10s\n" "Debug:" "$(ls -lh ${BINARY_DIR}/${BINARY_NAME}-debug | awk '{print $5}')"
printf "%-20s %10s\n" "Optimized (-s -w):" "$(ls -lh ${BINARY_DIR}/${BINARY_NAME}-optimized | awk '{print $5}')"

# Calculate reduction
REDUCTION=$((DEBUG_SIZE - OPTIMIZED_SIZE))
PERCENTAGE=$(awk "BEGIN {printf \"%.1f\", ($REDUCTION / $DEBUG_SIZE) * 100}")

echo ""
echo "=== Savings ==="
printf "%-20s %10s (%s%%)\n" "Size reduction:" "$(echo $REDUCTION | awk '{printf "%.1f MB", $1/1024/1024}')" "$PERCENTAGE"

# Cleanup
rm -f "${BINARY_DIR}/${BINARY_NAME}-debug"

echo ""
echo "=== Optimization Flags Used ==="
echo "  -trimpath          Remove file system paths from binary"
echo "  -ldflags -s        Disable symbol table"
echo "  -ldflags -w        Disable DWARF debug info"
echo ""
