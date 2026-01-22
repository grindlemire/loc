package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/urfave/cli/v3"
)

// TestCLIHelp tests that the help output is displayed correctly
func TestCLIHelp(t *testing.T) {
	var buf bytes.Buffer

	cmd := buildCLICommand()
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"loc", "--help"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	output := buf.String()

	// Check that expected flags are present in help
	expectedFlags := []string{
		"--by-language",
		"--by-dir",
		"--by-package",
		"--code-only",
		"--output",
		"--include",
		"--exclude",
		"--no-gitignore",
		"--workers",
	}

	for _, flag := range expectedFlags {
		if !strings.Contains(output, flag) {
			t.Errorf("expected help to contain %q, but it didn't", flag)
		}
	}
}

// TestCLIDefaultBehavior tests running loc with default settings
func TestCLIDefaultBehavior(t *testing.T) {
	// Create a temporary directory with some test files
	testDir := t.TempDir()

	// Create a simple Go file
	goFile := filepath.Join(testDir, "main.go")
	err := os.WriteFile(goFile, []byte(`package main

func main() {
	// Hello world
	println("Hello")
}
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer

	cmd := buildCLICommand()
	cmd.Writer = &buf

	err = cmd.Run(context.Background(), []string{"loc", testDir})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	output := buf.String()

	// Should show pretty table output with Files, Code, Comments columns
	if !strings.Contains(output, "Files") {
		t.Errorf("expected 'Files' header in pretty output, got: %s", output)
	}
	if !strings.Contains(output, "Code") {
		t.Errorf("expected 'Code' header in pretty output, got: %s", output)
	}
	if !strings.Contains(output, "Comments") {
		t.Errorf("expected 'Comments' header in pretty output, got: %s", output)
	}

	// Should have table border characters (rounded borders)
	if !strings.Contains(output, "\u2502") && !strings.Contains(output, "\u2500") {
		t.Errorf("expected table border characters in output, got: %s", output)
	}
}

// TestCLIByLanguage tests the --by-language flag
func TestCLIByLanguage(t *testing.T) {
	testDir := t.TempDir()

	// Create a Go file
	err := os.WriteFile(filepath.Join(testDir, "main.go"), []byte(`package main

func main() {}
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer

	cmd := buildCLICommand()
	cmd.Writer = &buf

	err = cmd.Run(context.Background(), []string{"loc", "--by-language", testDir})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	output := buf.String()

	// Should show language breakdown
	if !strings.Contains(output, "By Language") {
		t.Errorf("expected 'By Language' section, got: %s", output)
	}

	// Should show Go language
	if !strings.Contains(output, "Go") {
		t.Errorf("expected 'Go' in output, got: %s", output)
	}
}

// TestCLIJSONOutput tests the --output json flag
func TestCLIJSONOutput(t *testing.T) {
	testDir := t.TempDir()

	// Create a Go file
	err := os.WriteFile(filepath.Join(testDir, "main.go"), []byte(`package main

func main() {}
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer

	cmd := buildCLICommand()
	cmd.Writer = &buf

	err = cmd.Run(context.Background(), []string{"loc", "--output", "json", testDir})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	output := buf.String()

	// Should be valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Errorf("expected valid JSON, got error: %v\nOutput: %s", err, output)
	}

	// Check for expected fields
	if _, ok := result["total"]; !ok {
		t.Error("expected 'total' field in JSON output")
	}
	if _, ok := result["byLanguage"]; !ok {
		t.Error("expected 'byLanguage' field in JSON output")
	}
}

// TestCLIRawOutput tests the --output raw flag
func TestCLIRawOutput(t *testing.T) {
	testDir := t.TempDir()

	// Create a Go file
	err := os.WriteFile(filepath.Join(testDir, "main.go"), []byte(`package main

func main() {}
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer

	cmd := buildCLICommand()
	cmd.Writer = &buf

	err = cmd.Run(context.Background(), []string{"loc", "--output", "raw", testDir})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	output := buf.String()

	// Should contain labeled statistics
	expectedLabels := []string{"Files:", "Lines:", "Code:", "Blanks:", "Comments:"}
	for _, label := range expectedLabels {
		if !strings.Contains(output, label) {
			t.Errorf("expected %q in raw output, got: %s", label, output)
		}
	}
}

// TestCLIExcludePattern tests the --exclude flag
func TestCLIExcludePattern(t *testing.T) {
	testDir := t.TempDir()

	// Create multiple files
	err := os.WriteFile(filepath.Join(testDir, "main.go"), []byte(`package main
func main() {}
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(filepath.Join(testDir, "main_test.go"), []byte(`package main
func TestMain() {}
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer

	cmd := buildCLICommand()
	cmd.Writer = &buf

	// Exclude test files
	err = cmd.Run(context.Background(), []string{"loc", "--output", "raw", "--exclude", "*_test.go", testDir})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	output := buf.String()

	// Should only count one file (main.go)
	if !strings.Contains(output, "Files: 1") {
		t.Errorf("expected 'Files: 1' after exclude, got: %s", output)
	}
}

// TestCLIIncludePattern tests the --include flag
func TestCLIIncludePattern(t *testing.T) {
	testDir := t.TempDir()

	// Create files with different extensions
	err := os.WriteFile(filepath.Join(testDir, "main.go"), []byte(`package main
func main() {}
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(filepath.Join(testDir, "data.txt"), []byte(`some text
more text
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer

	cmd := buildCLICommand()
	cmd.Writer = &buf

	// Include only txt files (not normally a source file)
	err = cmd.Run(context.Background(), []string{"loc", "--output", "raw", "--include", "*.txt", testDir})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	output := buf.String()

	// Should only count the txt file
	if !strings.Contains(output, "Files: 1") {
		t.Errorf("expected 'Files: 1' with include pattern, got: %s", output)
	}
}

// TestCLIInvalidOutputFormat tests that invalid output format returns an error
func TestCLIInvalidOutputFormat(t *testing.T) {
	testDir := t.TempDir()

	var buf bytes.Buffer

	cmd := buildCLICommand()
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"loc", "--output", "invalid", testDir})
	if err == nil {
		t.Error("expected error for invalid output format")
	}

	if !strings.Contains(err.Error(), "invalid output format") {
		t.Errorf("expected 'invalid output format' error, got: %v", err)
	}
}

// TestCLIMultipleFlags tests using multiple flags together
func TestCLIMultipleFlags(t *testing.T) {
	testDir := t.TempDir()

	// Create subdirectory structure
	subDir := filepath.Join(testDir, "pkg")
	err := os.MkdirAll(subDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(filepath.Join(testDir, "main.go"), []byte(`package main
func main() {}
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(filepath.Join(subDir, "lib.go"), []byte(`package pkg
func Lib() {}
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer

	cmd := buildCLICommand()
	cmd.Writer = &buf

	// Use multiple breakdown flags
	err = cmd.Run(context.Background(), []string{"loc", "--by-language", "--by-dir", "--by-package", testDir})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	output := buf.String()

	// Should show all breakdowns
	if !strings.Contains(output, "By Language") {
		t.Errorf("expected 'By Language' section, got: %s", output)
	}
	if !strings.Contains(output, "By Directory") {
		t.Errorf("expected 'By Directory' section, got: %s", output)
	}
	if !strings.Contains(output, "By Package") {
		t.Errorf("expected 'By Package' section, got: %s", output)
	}
}

// TestCLINoGitignore tests the --no-gitignore flag
func TestCLINoGitignore(t *testing.T) {
	testDir := t.TempDir()

	// Create a file
	err := os.WriteFile(filepath.Join(testDir, "main.go"), []byte(`package main
func main() {}
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create a .gitignore that ignores main.go
	err = os.WriteFile(filepath.Join(testDir, ".gitignore"), []byte(`main.go
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// First, test without --no-gitignore (should respect gitignore)
	var buf1 bytes.Buffer
	cmd1 := buildCLICommand()
	cmd1.Writer = &buf1

	err = cmd1.Run(context.Background(), []string{"loc", "--output", "raw", testDir})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	output1 := buf1.String()
	if !strings.Contains(output1, "Files: 0") {
		t.Errorf("expected 'Files: 0' when respecting gitignore, got: %s", output1)
	}

	// Now test with --no-gitignore (should ignore gitignore)
	var buf2 bytes.Buffer
	cmd2 := buildCLICommand()
	cmd2.Writer = &buf2

	err = cmd2.Run(context.Background(), []string{"loc", "--output", "raw", "--no-gitignore", testDir})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	output2 := buf2.String()
	if !strings.Contains(output2, "Files: 1") {
		t.Errorf("expected 'Files: 1' with --no-gitignore, got: %s", output2)
	}
}

// TestCLIShortFlags tests using short flag aliases
func TestCLIShortFlags(t *testing.T) {
	testDir := t.TempDir()

	err := os.WriteFile(filepath.Join(testDir, "main.go"), []byte(`package main
func main() {}
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer

	cmd := buildCLICommand()
	cmd.Writer = &buf

	// Use short flags: -l for --by-language, -o for --output
	err = cmd.Run(context.Background(), []string{"loc", "-l", "-o", "raw", testDir})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	output := buf.String()

	// Should show language breakdown in raw format
	if !strings.Contains(output, "Go:") {
		t.Errorf("expected language stats in raw output, got: %s", output)
	}
}

// TestCLIEmptyDirectory tests running on an empty directory
func TestCLIEmptyDirectory(t *testing.T) {
	testDir := t.TempDir()

	var buf bytes.Buffer

	cmd := buildCLICommand()
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"loc", "--output", "raw", testDir})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	output := buf.String()

	// Should show zero files
	if !strings.Contains(output, "Files: 0") {
		t.Errorf("expected 'Files: 0' for empty directory, got: %s", output)
	}
}

// TestCLINonexistentPath tests running on a path that doesn't exist
func TestCLINonexistentPath(t *testing.T) {
	var buf bytes.Buffer

	cmd := buildCLICommand()
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"loc", "/nonexistent/path/that/should/not/exist"})
	if err == nil {
		t.Error("expected error for nonexistent path")
	}
}

// buildCLICommand creates a CLI command for testing
// This mirrors the setup in main() but returns the command for testing
func buildCLICommand() *cli.Command {
	return &cli.Command{
		Name:    "loc",
		Usage:   "A fast, parallel lines-of-code counter for software projects",
		Version: "test",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "by-language",
				Aliases: []string{"l"},
				Usage:   "Show breakdown by language",
			},
			&cli.BoolFlag{
				Name:    "by-dir",
				Aliases: []string{"d"},
				Usage:   "Show breakdown by directory",
			},
			&cli.BoolFlag{
				Name:    "by-package",
				Aliases: []string{"p"},
				Usage:   "Show breakdown by package",
			},
			&cli.BoolFlag{
				Name:    "code-only",
				Aliases: []string{"c"},
				Usage:   "Only count code lines (exclude blanks/comments)",
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Value:   "pretty",
				Usage:   "Output format: pretty, json, raw",
			},
			&cli.StringSliceFlag{
				Name:    "include",
				Aliases: []string{"i"},
				Usage:   "Additional glob patterns to include",
			},
			&cli.StringSliceFlag{
				Name:    "exclude",
				Aliases: []string{"e"},
				Usage:   "Glob patterns to exclude",
			},
			&cli.BoolFlag{
				Name:  "no-gitignore",
				Usage: "Don't respect .gitignore files",
			},
			&cli.IntFlag{
				Name:    "workers",
				Aliases: []string{"w"},
				Value:   4,
				Usage:   "Number of parallel workers",
			},
		},
		Action: run,
	}
}
