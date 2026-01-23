#!/bin/bash
# Dogfood script: Run difftron on itself
# Usage: ./scripts/dogfood.sh [base-ref] [head-ref]

set -e

BASE_REF="${1:-HEAD~1}"
HEAD_REF="${2:-HEAD}"

echo "🐕 Dogfooding Difftron on itself"
echo "=================================="
echo "Base: $BASE_REF"
echo "Head: $HEAD_REF"
echo ""

# Build difftron
echo "Building difftron..."
go build -o bin/difftron ./cmd/difftron

# Generate coverage
echo "Generating coverage..."
go test -coverprofile=coverage.out ./...

# Generate diff
echo "Generating diff..."
git diff "$BASE_REF".."$HEAD_REF" > diff.patch || {
    echo "Warning: No changes detected or diff failed"
    exit 0
}

# Check if there are any changes
if [ ! -s diff.patch ]; then
    echo "No changes to analyze."
    exit 0
fi

# Run analysis
echo ""
echo "Running difftron analysis..."
echo ""

./bin/difftron analyze \
    --coverage coverage.out \
    --diff diff.patch \
    --threshold 80 \
    --output text

EXIT_CODE=$?

echo ""
if [ $EXIT_CODE -eq 0 ]; then
    echo "✅ Coverage check passed!"
else
    echo "❌ Coverage check failed!"
fi

# Cleanup
rm -f diff.patch

exit $EXIT_CODE
