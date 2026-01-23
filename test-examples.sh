#!/bin/bash
# Quick test examples for difftron

echo "=== Running Unit Tests ==="
go run scripts/task.go test

echo ""
echo "=== Running Integration Test ==="
go run scripts/task.go test-integration

echo ""
echo "=== Building Binary ==="
go run scripts/task.go build

echo ""
echo "=== Testing CLI with Fixtures ==="
./bin/difftron analyze \
  --coverage testdata/fixtures/tronswan-coverage.info \
  --diff testdata/fixtures/sample.diff \
  --threshold 0 \
  --output text

echo ""
echo "=== Testing JSON Output ==="
./bin/difftron analyze \
  --coverage testdata/fixtures/tronswan-coverage.info \
  --diff testdata/fixtures/sample.diff \
  --threshold 0 \
  --output json | head -20

echo ""
echo "=== All tests complete! ==="
