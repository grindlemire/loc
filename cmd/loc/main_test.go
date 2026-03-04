package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCLIHelp tests that the help output is displayed correctly
func TestCLIHelp(t *testing.T) {
	var buf bytes.Buffer

	cmd := buildCommand()
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

	cmd := buildCommand()
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

	cmd := buildCommand()
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

	cmd := buildCommand()
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

	cmd := buildCommand()
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

	cmd := buildCommand()
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

	cmd := buildCommand()
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

	cmd := buildCommand()
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

	cmd := buildCommand()
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
	cmd1 := buildCommand()
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
	cmd2 := buildCommand()
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

	cmd := buildCommand()
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

	cmd := buildCommand()
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

	cmd := buildCommand()
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"loc", "/nonexistent/path/that/should/not/exist"})
	if err == nil {
		t.Error("expected error for nonexistent path")
	}
}

// TestCLINoColorFlag tests the --no-color flag
func TestCLINoColorFlag(t *testing.T) {
	testDir := t.TempDir()

	// Create a Go file
	err := os.WriteFile(filepath.Join(testDir, "main.go"), []byte(`package main

func main() {}
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer

	cmd := buildCommand()
	cmd.Writer = &buf

	err = cmd.Run(context.Background(), []string{"loc", "--no-color", testDir})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	output := buf.String()

	// Should still have table structure with headers
	if !strings.Contains(output, "Files") {
		t.Errorf("expected 'Files' header in no-color output, got: %s", output)
	}
	if !strings.Contains(output, "Code") {
		t.Errorf("expected 'Code' header in no-color output, got: %s", output)
	}
	if !strings.Contains(output, "Comments") {
		t.Errorf("expected 'Comments' header in no-color output, got: %s", output)
	}

	// Should have table border characters (rounded borders still work in no-color mode)
	if !strings.Contains(output, "\u2502") && !strings.Contains(output, "\u2500") {
		t.Errorf("expected table border characters in no-color output, got: %s", output)
	}
}

// TestCLINoColorFlagWithBreakdowns tests --no-color with breakdown flags
func TestCLINoColorFlagWithBreakdowns(t *testing.T) {
	testDir := t.TempDir()

	// Create a Go file
	err := os.WriteFile(filepath.Join(testDir, "main.go"), []byte(`package main

func main() {}
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer

	cmd := buildCommand()
	cmd.Writer = &buf

	err = cmd.Run(context.Background(), []string{"loc", "--no-color", "--by-language", testDir})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	output := buf.String()

	// Should show language breakdown
	if !strings.Contains(output, "By Language") {
		t.Errorf("expected 'By Language' section in no-color output, got: %s", output)
	}

	// Should have percentage column
	if !strings.Contains(output, "%") {
		t.Errorf("expected '%%' column in no-color output, got: %s", output)
	}

	// Should have Total row
	if !strings.Contains(output, "Total") {
		t.Errorf("expected 'Total' row in no-color output, got: %s", output)
	}
}

// TestCLICombinedFlag tests the --combined flag
func TestCLICombinedFlag(t *testing.T) {
	testDir := t.TempDir()

	// Create source and test files
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

	t.Run("default shows src/test breakdown", func(t *testing.T) {
		var buf bytes.Buffer
		cmd := buildCommand()
		cmd.Writer = &buf

		err = cmd.Run(context.Background(), []string{"loc", testDir})
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		output := buf.String()

		// Default output should show src/test sub-rows
		if !strings.Contains(output, "src") {
			t.Errorf("expected 'src' sub-row in default output, got: %s", output)
		}
		if !strings.Contains(output, "test") {
			t.Errorf("expected 'test' sub-row in default output, got: %s", output)
		}
	})

	t.Run("combined hides src/test breakdown", func(t *testing.T) {
		var buf bytes.Buffer
		cmd := buildCommand()
		cmd.Writer = &buf

		err = cmd.Run(context.Background(), []string{"loc", "--combined", testDir})
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		output := buf.String()

		// Combined output should NOT show src/test sub-rows
		// The output should just have the simple table without breakdown labels
		if strings.Contains(output, "src") && strings.Contains(output, "test") {
			// Check it's not a breakdown - just basic stats
			// In combined mode, we don't see the breakdown prefixes
			if strings.Contains(output, "+- src") || strings.Contains(output, "\u251C\u2500 src") {
				t.Errorf("expected no src/test breakdown with --combined flag, got: %s", output)
			}
		}
	})
}

// TestCLIAllFlag tests the --all flag
func TestCLIAllFlag(t *testing.T) {
	testDir := t.TempDir()

	// Create a Go file
	err := os.WriteFile(filepath.Join(testDir, "main.go"), []byte(`package main

func main() {}
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create a markdown file (non-source file)
	err = os.WriteFile(filepath.Join(testDir, "README.md"), []byte(`# README

This is documentation.
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create a JSON config file
	err = os.WriteFile(filepath.Join(testDir, "config.json"), []byte(`{
  "key": "value"
}
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("without all flag", func(t *testing.T) {
		var buf bytes.Buffer
		cmd := buildCommand()
		cmd.Writer = &buf

		err = cmd.Run(context.Background(), []string{"loc", "--output", "raw", testDir})
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		output := buf.String()

		// Without --all, should only count source file (1 file)
		if !strings.Contains(output, "Files: 1") {
			t.Errorf("expected 'Files: 1' without --all flag, got: %s", output)
		}
	})

	t.Run("with all flag", func(t *testing.T) {
		var buf bytes.Buffer
		cmd := buildCommand()
		cmd.Writer = &buf

		err = cmd.Run(context.Background(), []string{"loc", "--output", "raw", "--all", testDir})
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		output := buf.String()

		// With --all, should count all files (3 files: main.go, README.md, config.json)
		if !strings.Contains(output, "Files: 3") {
			t.Errorf("expected 'Files: 3' with --all flag, got: %s", output)
		}

		// Should show "other" files in the breakdown
		if !strings.Contains(output, "Other Files:") {
			t.Errorf("expected 'Other Files:' in raw output with --all flag, got: %s", output)
		}
	})

	t.Run("with all flag short form", func(t *testing.T) {
		var buf bytes.Buffer
		cmd := buildCommand()
		cmd.Writer = &buf

		err = cmd.Run(context.Background(), []string{"loc", "--output", "raw", "-a", testDir})
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		output := buf.String()

		// With -a (short form), should count all files
		if !strings.Contains(output, "Files: 3") {
			t.Errorf("expected 'Files: 3' with -a flag, got: %s", output)
		}
	})
}

// TestCLIDefaultShowsSrcTestBreakdown tests that default output shows src/test breakdown
func TestCLIDefaultShowsSrcTestBreakdown(t *testing.T) {
	testDir := t.TempDir()

	// Create a source file and a test file
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
	cmd := buildCommand()
	cmd.Writer = &buf

	err = cmd.Run(context.Background(), []string{"loc", testDir})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	output := buf.String()

	// Default output should show Total row with src/test sub-rows
	if !strings.Contains(output, "Total") {
		t.Errorf("expected 'Total' row in default output, got: %s", output)
	}

	// Should have src and test sub-rows (with elbow prefixes)
	if !strings.Contains(output, "src") {
		t.Errorf("expected 'src' in default output, got: %s", output)
	}
	if !strings.Contains(output, "test") {
		t.Errorf("expected 'test' in default output, got: %s", output)
	}
}

// TestCLIJSONOutputWithSrcTestBreakdown tests JSON output with src/test breakdown
func TestCLIJSONOutputWithSrcTestBreakdown(t *testing.T) {
	testDir := t.TempDir()

	// Create a source file and a test file
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
	cmd := buildCommand()
	cmd.Writer = &buf

	err = cmd.Run(context.Background(), []string{"loc", "--output", "json", testDir})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	output := buf.String()

	// Parse JSON output
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected valid JSON, got error: %v", err)
	}

	// Default JSON should include src and test stats
	if _, ok := result["src"]; !ok {
		t.Error("expected 'src' field in JSON output by default")
	}
	if _, ok := result["test"]; !ok {
		t.Error("expected 'test' field in JSON output by default")
	}

	// Check that src contains the expected structure
	src, ok := result["src"].(map[string]interface{})
	if !ok {
		t.Error("expected 'src' to be an object")
	} else {
		if _, ok := src["files"]; !ok {
			t.Error("expected 'files' field in src object")
		}
		if _, ok := src["code"]; !ok {
			t.Error("expected 'code' field in src object")
		}
	}
}

// TestCLIJSONOutputWithCombinedFlag tests JSON output with --combined flag
func TestCLIJSONOutputWithCombinedFlag(t *testing.T) {
	testDir := t.TempDir()

	// Create a source file and a test file
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
	cmd := buildCommand()
	cmd.Writer = &buf

	err = cmd.Run(context.Background(), []string{"loc", "--output", "json", "--combined", testDir})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	output := buf.String()

	// Parse JSON output
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("expected valid JSON, got error: %v", err)
	}

	// With --combined, JSON should NOT include src and test stats
	if _, ok := result["src"]; ok {
		t.Error("did not expect 'src' field in JSON output with --combined")
	}
	if _, ok := result["test"]; ok {
		t.Error("did not expect 'test' field in JSON output with --combined")
	}

	// But should still have total
	if _, ok := result["total"]; !ok {
		t.Error("expected 'total' field in JSON output")
	}
}

// TestCLIHelpShowsNewFlags tests that help output shows --combined and --all flags
func TestCLIHelpShowsNewFlags(t *testing.T) {
	var buf bytes.Buffer

	cmd := buildCommand()
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"loc", "--help"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	output := buf.String()

	// Check that new flags are present in help
	if !strings.Contains(output, "--combined") {
		t.Errorf("expected --combined flag in help output, got: %s", output)
	}
	if !strings.Contains(output, "--all") {
		t.Errorf("expected --all flag in help output, got: %s", output)
	}
	if !strings.Contains(output, "-a") {
		t.Errorf("expected -a alias in help output, got: %s", output)
	}
}

// TestCLINoTestsFlag tests the --no-tests flag
func TestCLINoTestsFlag(t *testing.T) {
	testDir := t.TempDir()

	err := os.WriteFile(filepath.Join(testDir, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(testDir, "main_test.go"), []byte("package main\n\nfunc TestMain() {}\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	cmd := buildCommand()
	cmd.Writer = &buf

	err = cmd.Run(context.Background(), []string{"loc", "--output", "raw", "--no-tests", testDir})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Files: 1") {
		t.Errorf("expected 'Files: 1' with --no-tests, got: %s", output)
	}
}

// TestCLITestsOnlyFlag tests the --tests-only flag
func TestCLITestsOnlyFlag(t *testing.T) {
	testDir := t.TempDir()

	err := os.WriteFile(filepath.Join(testDir, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(testDir, "main_test.go"), []byte("package main\n\nfunc TestMain() {}\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	cmd := buildCommand()
	cmd.Writer = &buf

	err = cmd.Run(context.Background(), []string{"loc", "--output", "raw", "--tests-only", testDir})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Files: 1") {
		t.Errorf("expected 'Files: 1' with --tests-only, got: %s", output)
	}
}

// TestCLINoTestsAndTestsOnlyMutuallyExclusive tests that --no-tests and --tests-only cannot be used together
func TestCLINoTestsAndTestsOnlyMutuallyExclusive(t *testing.T) {
	testDir := t.TempDir()

	err := os.WriteFile(filepath.Join(testDir, "main.go"), []byte("package main\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	cmd := buildCommand()
	cmd.Writer = &buf

	err = cmd.Run(context.Background(), []string{"loc", "--no-tests", "--tests-only", testDir})
	if err == nil {
		t.Error("expected error when using --no-tests and --tests-only together")
	}
	if err != nil && !strings.Contains(err.Error(), "mutually exclusive") {
		t.Errorf("expected 'mutually exclusive' error, got: %v", err)
	}
}

// TestCLILangFilter tests the --lang flag
func TestCLILangFilter(t *testing.T) {
	testDir := t.TempDir()

	err := os.WriteFile(filepath.Join(testDir, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(testDir, "app.py"), []byte("print('hi')\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(testDir, "index.js"), []byte("console.log('hi')\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	cmd := buildCommand()
	cmd.Writer = &buf

	err = cmd.Run(context.Background(), []string{"loc", "--output", "raw", "--lang", "Go", testDir})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Files: 1") {
		t.Errorf("expected 'Files: 1' with --lang Go, got: %s", output)
	}
}

// TestCLILangFilterCommaSeparated tests --lang with comma-separated values
func TestCLILangFilterCommaSeparated(t *testing.T) {
	testDir := t.TempDir()

	err := os.WriteFile(filepath.Join(testDir, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(testDir, "app.py"), []byte("print('hi')\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(testDir, "index.js"), []byte("console.log('hi')\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	cmd := buildCommand()
	cmd.Writer = &buf

	err = cmd.Run(context.Background(), []string{"loc", "--output", "raw", "--lang", "Go,Python", testDir})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Files: 2") {
		t.Errorf("expected 'Files: 2' with --lang Go,Python, got: %s", output)
	}
}

// TestCLIExcludeLangFilter tests the --exclude-lang flag
func TestCLIExcludeLangFilter(t *testing.T) {
	testDir := t.TempDir()

	err := os.WriteFile(filepath.Join(testDir, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(testDir, "app.py"), []byte("print('hi')\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(testDir, "index.js"), []byte("console.log('hi')\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	cmd := buildCommand()
	cmd.Writer = &buf

	err = cmd.Run(context.Background(), []string{"loc", "--output", "raw", "--exclude-lang", "Go", testDir})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Files: 2") {
		t.Errorf("expected 'Files: 2' with --exclude-lang Go, got: %s", output)
	}
}

// TestCLIHelpShowsFilterFlags tests that help output includes the new filter flags
func TestCLIHelpShowsFilterFlags(t *testing.T) {
	var buf bytes.Buffer
	cmd := buildCommand()
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"loc", "--help"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	output := buf.String()
	for _, flag := range []string{"--no-tests", "--tests-only", "--lang", "--exclude-lang"} {
		if !strings.Contains(output, flag) {
			t.Errorf("expected %q in help output", flag)
		}
	}
}
