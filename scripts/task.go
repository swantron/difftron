//go:build ignore

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	task := os.Args[1]
	args := os.Args[2:]

	switch task {
	case "build":
		run("go", "build", "-o", "bin/difftron", "./cmd/difftron")
	case "test":
		run("go", "test", "-v", "./...")
	case "test-coverage":
		run("go", "test", "-coverprofile=coverage.out", "./...")
		run("go", "tool", "cover", "-html=coverage.out", "-o", "coverage.html")
	case "install":
		run("go", "install", "./cmd/difftron")
	case "fmt":
		run("go", "fmt", "./...")
	case "lint":
		run("golangci-lint", "run")
	case "clean":
		clean()
	case "run":
		if len(args) > 0 {
			// Pass remaining args to the CLI
			cmd := exec.Command("go", append([]string{"run", "./cmd/difftron"}, args...)...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin
			if err := cmd.Run(); err != nil {
				os.Exit(1)
			}
		} else {
			run("go", "run", "./cmd/difftron")
		}
	case "test-integration":
		// Run integration test with fixtures
		coverageFile := "testdata/fixtures/tronswan-coverage.info"
		diffFile := "testdata/fixtures/sample.diff"
		if _, err := os.Stat(coverageFile); os.IsNotExist(err) {
			fmt.Printf("Error: Coverage file not found: %s\n", coverageFile)
			os.Exit(1)
		}
		if _, err := os.Stat(diffFile); os.IsNotExist(err) {
			fmt.Printf("Error: Diff file not found: %s\n", diffFile)
			os.Exit(1)
		}
		run("go", "run", "./cmd/difftron", "analyze",
			"--coverage", coverageFile,
			"--diff", diffFile,
			"--threshold", "0")
	default:
		fmt.Printf("Unknown task: %s\n\n", task)
		printUsage()
		os.Exit(1)
	}
}

func run(command string, args ...string) {
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error: Command failed: %s %v\n", command, args)
		os.Exit(1)
	}
}

func clean() {
	dirs := []string{"bin", "coverage.out", "coverage.html"}
	for _, dir := range dirs {
		if err := os.RemoveAll(dir); err != nil {
			fmt.Printf("Warning: Failed to remove %s: %v\n", dir, err)
		}
	}
	fmt.Println("Cleaned build artifacts")
}

func printUsage() {
	exe := filepath.Base(os.Args[0])
	fmt.Printf("Usage: go run %s <task> [args...]\n\n", exe)
	fmt.Println("Available tasks:")
	fmt.Println("  build           - Build the difftron binary")
	fmt.Println("  test            - Run all tests")
	fmt.Println("  test-coverage   - Run tests with coverage report")
	fmt.Println("  test-integration - Run integration test with fixtures")
	fmt.Println("  install         - Install the binary to $GOPATH/bin")
	fmt.Println("  fmt             - Format code")
	fmt.Println("  lint            - Run linter (requires golangci-lint)")
	fmt.Println("  clean           - Remove build artifacts")
	fmt.Println("  run [args...]   - Run the CLI locally (passes args to CLI)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Printf("  go run %s build\n", exe)
	fmt.Printf("  go run %s test\n", exe)
	fmt.Printf("  go run %s run analyze --coverage coverage.info\n", exe)
}
