#!/bin/bash
# Generate coverage report for difftron itself

set -e

echo "Generating coverage for difftron..."

# Run tests with coverage
go test -coverprofile=coverage.out ./...

# Generate LCOV format (for difftron to consume)
go tool cover -func=coverage.out | awk '
BEGIN {
    current_file = ""
}
/^github.com\/swantron\/difftron\// {
    # Extract file path
    match($0, /github.com\/swantron\/difftron\/(.+):/, arr)
    file = arr[1]
    
    # Extract line coverage info
    match($0, /:([0-9]+)\.[0-9]+%/, arr2)
    coverage = arr2[1]
    
    if (file != current_file) {
        if (current_file != "") {
            print "end_of_record"
        }
        print "SF:" file
        current_file = file
    }
    
    # Note: Go's cover tool doesn't give us line-by-line data
    # This is a simplified version. For full LCOV, we'd need to use
    # go test -coverprofile with go tool cover -html
    print "DA:" NR "," (coverage > 0 ? "1" : "0")
}
END {
    if (current_file != "") {
        print "end_of_record"
    }
}' > coverage.lcov || {
    echo "Note: Simplified LCOV generation. Using coverage.out directly."
    # Convert coverage.out to a basic LCOV format
    go tool cover -func=coverage.out > coverage.txt
}

echo "Coverage files generated:"
echo "  - coverage.out (Go format)"
echo "  - coverage.txt (human readable)"
echo ""
echo "To analyze your changes:"
echo "  git diff HEAD~1 HEAD > diff.patch"
echo "  ./bin/difftron analyze --coverage coverage.out --diff diff.patch"
