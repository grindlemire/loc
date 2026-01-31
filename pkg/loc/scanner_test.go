package loc

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// Helper to create test file structure
func createTestFiles(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for path, content := range files {
		fullPath := filepath.Join(root, path)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", fullPath, err)
		}
	}
}

// Helper to get relative paths from scan result
func getRelativePaths(t *testing.T, root string, files []string) []string {
	t.Helper()
	var result []string
	for _, f := range files {
		rel, err := filepath.Rel(root, f)
		if err != nil {
			t.Fatalf("Failed to get relative path: %v", err)
		}
		// Normalize to forward slashes for comparison
		result = append(result, filepath.ToSlash(rel))
	}
	sort.Strings(result)
	return result
}

func TestScannerDefaultExclusions(t *testing.T) {
	tempDir := t.TempDir()

	// Create test file structure with various files
	files := map[string]string{
		// Valid source files
		"main.go":           "package main",
		"src/app.go":        "package src",
		"lib/utils.py":      "def foo(): pass",
		"web/index.js":      "console.log('hi')",
		// Binary files (should be excluded)
		"bin/program.exe":   "",
		"lib/native.so":     "",
		"lib/native.dylib":  "",
		// Generated files (should be excluded)
		"api/types.pb.go":        "package api",
		"gen/model_generated.go": "package gen",
		"ui/widget_templ.go":     "package ui",
		// Minified/bundled JS (should be excluded)
		"dist/app.min.js":    "",
		"dist/app.bundle.js": "",
		"dist/app.js.map":    "",
		// Lock files (should be excluded)
		"package-lock.json": "{}",
		"yarn.lock":         "",
		"go.sum":            "",
		// Files in excluded directories (should be excluded)
		"node_modules/pkg/index.js":   "",
		"vendor/lib/lib.go":           "",
		".git/config":                 "",
		"build/output.js":             "",
		"dist/bundle.js":              "",
		"__pycache__/cache.pyc":       "",
		"target/debug/main.rs":        "",
		"_build/dev/lib.ex":           "",
	}
	createTestFiles(t, tempDir, files)

	config := &Config{}
	scanner := NewScanner(config)

	result, err := scanner.Scan(tempDir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	relPaths := getRelativePaths(t, tempDir, result)

	// Expected files (only valid source files not in excluded dirs/patterns)
	expected := []string{
		"lib/utils.py",
		"main.go",
		"src/app.go",
		"web/index.js",
	}
	sort.Strings(expected)

	if len(relPaths) != len(expected) {
		t.Errorf("Expected %d files, got %d", len(expected), len(relPaths))
		t.Errorf("Expected: %v", expected)
		t.Errorf("Got: %v", relPaths)
		return
	}

	for i, path := range expected {
		if relPaths[i] != path {
			t.Errorf("Expected %s at position %d, got %s", path, i, relPaths[i])
		}
	}
}

func TestScannerIncludePatterns(t *testing.T) {
	tempDir := t.TempDir()

	files := map[string]string{
		"main.go":      "package main",
		"test.go":      "package main",
		"main_test.go": "package main",
		"app.py":       "print('hi')",
		"readme.md":    "# README",
		"config.json":  "{}",
	}
	createTestFiles(t, tempDir, files)

	// Include only test files
	config := &Config{
		Include: []string{"*_test.go"},
	}
	scanner := NewScanner(config)

	result, err := scanner.Scan(tempDir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	relPaths := getRelativePaths(t, tempDir, result)

	expected := []string{"main_test.go"}

	if len(relPaths) != len(expected) {
		t.Errorf("Expected %d files, got %d", len(expected), len(relPaths))
		t.Errorf("Expected: %v", expected)
		t.Errorf("Got: %v", relPaths)
		return
	}

	for i, path := range expected {
		if relPaths[i] != path {
			t.Errorf("Expected %s, got %s", path, relPaths[i])
		}
	}
}

func TestScannerExcludePatterns(t *testing.T) {
	tempDir := t.TempDir()

	files := map[string]string{
		"main.go":      "package main",
		"main_test.go": "package main",
		"utils.go":     "package main",
		"utils_test.go": "package main",
		"app.py":       "print('hi')",
	}
	createTestFiles(t, tempDir, files)

	// Exclude test files
	config := &Config{
		Exclude: []string{"*_test.go"},
	}
	scanner := NewScanner(config)

	result, err := scanner.Scan(tempDir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	relPaths := getRelativePaths(t, tempDir, result)

	expected := []string{"app.py", "main.go", "utils.go"}
	sort.Strings(expected)

	if len(relPaths) != len(expected) {
		t.Errorf("Expected %d files, got %d", len(expected), len(relPaths))
		t.Errorf("Expected: %v", expected)
		t.Errorf("Got: %v", relPaths)
		return
	}

	for i, path := range expected {
		if relPaths[i] != path {
			t.Errorf("Expected %s at position %d, got %s", path, i, relPaths[i])
		}
	}
}

func TestScannerGitignoreIntegration(t *testing.T) {
	tempDir := t.TempDir()

	// Create .gitignore
	gitignore := `
*.log
temp/
`
	files := map[string]string{
		".gitignore":     gitignore,
		"main.go":        "package main",
		"debug.log":      "debug output",
		"app.log":        "app output",
		"temp/cache.go":  "package temp",
		"src/utils.go":   "package src",
	}
	createTestFiles(t, tempDir, files)

	config := &Config{}
	scanner := NewScanner(config)

	result, err := scanner.Scan(tempDir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	relPaths := getRelativePaths(t, tempDir, result)

	expected := []string{"main.go", "src/utils.go"}
	sort.Strings(expected)

	if len(relPaths) != len(expected) {
		t.Errorf("Expected %d files, got %d", len(expected), len(relPaths))
		t.Errorf("Expected: %v", expected)
		t.Errorf("Got: %v", relPaths)
		return
	}

	for i, path := range expected {
		if relPaths[i] != path {
			t.Errorf("Expected %s at position %d, got %s", path, i, relPaths[i])
		}
	}
}

func TestScannerNoGitignoreFlag(t *testing.T) {
	tempDir := t.TempDir()

	// Create .gitignore that would normally exclude test.go
	gitignore := `
test.go
`
	files := map[string]string{
		".gitignore": gitignore,
		"main.go":    "package main",
		"test.go":    "package main",
	}
	createTestFiles(t, tempDir, files)

	// With NoGitignore=true, test.go should be included
	config := &Config{
		NoGitignore: true,
	}
	scanner := NewScanner(config)

	result, err := scanner.Scan(tempDir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	relPaths := getRelativePaths(t, tempDir, result)

	expected := []string{"main.go", "test.go"}
	sort.Strings(expected)

	if len(relPaths) != len(expected) {
		t.Errorf("Expected %d files, got %d", len(expected), len(relPaths))
		t.Errorf("Expected: %v", expected)
		t.Errorf("Got: %v", relPaths)
		return
	}
}

func TestScannerExcludeDirectory(t *testing.T) {
	tempDir := t.TempDir()

	files := map[string]string{
		"main.go":          "package main",
		"tests/test1.go":   "package tests",
		"tests/test2.go":   "package tests",
		"src/app.go":       "package src",
		"src/tests/t.go":   "package tests",
	}
	createTestFiles(t, tempDir, files)

	// Exclude tests directory
	config := &Config{
		Exclude: []string{"tests/"},
	}
	scanner := NewScanner(config)

	result, err := scanner.Scan(tempDir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	relPaths := getRelativePaths(t, tempDir, result)

	expected := []string{"main.go", "src/app.go"}
	sort.Strings(expected)

	if len(relPaths) != len(expected) {
		t.Errorf("Expected %d files, got %d", len(expected), len(relPaths))
		t.Errorf("Expected: %v", expected)
		t.Errorf("Got: %v", relPaths)
		return
	}
}

func TestScannerSingleFile(t *testing.T) {
	tempDir := t.TempDir()

	files := map[string]string{
		"main.go": "package main",
		"app.py":  "print('hi')",
	}
	createTestFiles(t, tempDir, files)

	config := &Config{}
	scanner := NewScanner(config)

	// Scan a single file
	singleFile := filepath.Join(tempDir, "main.go")
	result, err := scanner.Scan(singleFile)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("Expected 1 file, got %d", len(result))
		return
	}

	if result[0] != singleFile {
		t.Errorf("Expected %s, got %s", singleFile, result[0])
	}
}

func TestScannerNonSourceFile(t *testing.T) {
	tempDir := t.TempDir()

	files := map[string]string{
		"readme.md":   "# README",
		"config.json": "{}",
	}
	createTestFiles(t, tempDir, files)

	config := &Config{}
	scanner := NewScanner(config)

	// Scan a non-source file
	nonSource := filepath.Join(tempDir, "readme.md")
	result, err := scanner.Scan(nonSource)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Expected 0 files for non-source file, got %d", len(result))
	}
}

func TestScannerEmptyDirectory(t *testing.T) {
	tempDir := t.TempDir()

	config := &Config{}
	scanner := NewScanner(config)

	result, err := scanner.Scan(tempDir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Expected 0 files for empty directory, got %d", len(result))
	}
}

func TestScannerNonexistentPath(t *testing.T) {
	config := &Config{}
	scanner := NewScanner(config)

	_, err := scanner.Scan("/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Error("Expected error for nonexistent path")
	}
}

func TestScannerSymlinks(t *testing.T) {
	tempDir := t.TempDir()

	// Create actual files
	files := map[string]string{
		"src/main.go": "package main",
		"src/app.go":  "package main",
	}
	createTestFiles(t, tempDir, files)

	// Create symlink to a file
	srcFile := filepath.Join(tempDir, "src", "main.go")
	linkFile := filepath.Join(tempDir, "link.go")
	if err := os.Symlink(srcFile, linkFile); err != nil {
		t.Skipf("Symlinks not supported: %v", err)
	}

	// Create symlink to a directory
	srcDir := filepath.Join(tempDir, "src")
	linkDir := filepath.Join(tempDir, "linked_src")
	if err := os.Symlink(srcDir, linkDir); err != nil {
		t.Skipf("Directory symlinks not supported: %v", err)
	}

	config := &Config{}
	scanner := NewScanner(config)

	result, err := scanner.Scan(tempDir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	// WalkDir by default follows symlinks to files but not directories
	// We expect src/main.go, src/app.go, and link.go
	// linked_src/ is a symlink to a directory, WalkDir does not follow it by default
	relPaths := getRelativePaths(t, tempDir, result)

	// At minimum, we should have the original files
	if len(relPaths) < 2 {
		t.Errorf("Expected at least 2 files, got %d: %v", len(relPaths), relPaths)
	}

	// Check that source files are included
	hasMain := false
	hasApp := false
	for _, p := range relPaths {
		if p == "src/main.go" {
			hasMain = true
		}
		if p == "src/app.go" {
			hasApp = true
		}
	}

	if !hasMain {
		t.Error("Expected src/main.go to be included")
	}
	if !hasApp {
		t.Error("Expected src/app.go to be included")
	}
}

func TestScannerCombinedIncludeExclude(t *testing.T) {
	tempDir := t.TempDir()

	files := map[string]string{
		"main.go":       "package main",
		"main_test.go":  "package main",
		"utils.go":      "package main",
		"utils_test.go": "package main",
		"app.go":        "package main",
	}
	createTestFiles(t, tempDir, files)

	// Include only .go files, but exclude test files
	config := &Config{
		Include: []string{"*.go"},
		Exclude: []string{"*_test.go"},
	}
	scanner := NewScanner(config)

	result, err := scanner.Scan(tempDir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	relPaths := getRelativePaths(t, tempDir, result)

	expected := []string{"app.go", "main.go", "utils.go"}
	sort.Strings(expected)

	if len(relPaths) != len(expected) {
		t.Errorf("Expected %d files, got %d", len(expected), len(relPaths))
		t.Errorf("Expected: %v", expected)
		t.Errorf("Got: %v", relPaths)
		return
	}

	for i, path := range expected {
		if relPaths[i] != path {
			t.Errorf("Expected %s at position %d, got %s", path, i, relPaths[i])
		}
	}
}

func TestScannerNestedDirectoryExclusion(t *testing.T) {
	tempDir := t.TempDir()

	files := map[string]string{
		"main.go":                      "package main",
		"pkg/lib.go":                   "package pkg",
		"pkg/node_modules/dep/index.js": "module.exports = {}",
		"web/node_modules/pkg/app.js":   "console.log('hi')",
		"src/vendor/lib.go":            "package vendor",
	}
	createTestFiles(t, tempDir, files)

	config := &Config{}
	scanner := NewScanner(config)

	result, err := scanner.Scan(tempDir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	relPaths := getRelativePaths(t, tempDir, result)

	// Only main.go and pkg/lib.go should be included
	// node_modules and vendor are built-in exclusions
	expected := []string{"main.go", "pkg/lib.go"}
	sort.Strings(expected)

	if len(relPaths) != len(expected) {
		t.Errorf("Expected %d files, got %d", len(expected), len(relPaths))
		t.Errorf("Expected: %v", expected)
		t.Errorf("Got: %v", relPaths)
	}
}

func TestScannerDoubleStarPattern(t *testing.T) {
	tempDir := t.TempDir()

	files := map[string]string{
		"main.go":               "package main",
		"tests/unit/test1.go":   "package tests",
		"tests/integration/t.go": "package tests",
		"src/app.go":            "package src",
		"src/tests/helper.go":   "package tests",
	}
	createTestFiles(t, tempDir, files)

	// Exclude all test directories anywhere in the tree
	config := &Config{
		Exclude: []string{"**/tests/**"},
	}
	scanner := NewScanner(config)

	result, err := scanner.Scan(tempDir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	relPaths := getRelativePaths(t, tempDir, result)

	expected := []string{"main.go", "src/app.go"}
	sort.Strings(expected)

	if len(relPaths) != len(expected) {
		t.Errorf("Expected %d files, got %d", len(expected), len(relPaths))
		t.Errorf("Expected: %v", expected)
		t.Errorf("Got: %v", relPaths)
	}
}

func TestMatchGlobPattern(t *testing.T) {
	tests := []struct {
		pattern string
		name    string
		match   bool
	}{
		// Simple patterns
		{"*.go", "main.go", true},
		{"*.go", "main.py", false},
		{"*_test.go", "main_test.go", true},
		{"*_test.go", "main.go", false},

		// Question mark
		{"test?.go", "test1.go", true},
		{"test?.go", "test12.go", false},

		// Exact match
		{"main.go", "main.go", true},
		{"main.go", "other.go", false},

		// Path patterns with **
		{"**/test.go", "src/test.go", true},
		{"**/test.go", "test.go", true},
		{"src/**/*.go", "src/pkg/main.go", true},
		{"src/**/*.go", "other/main.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.name, func(t *testing.T) {
			result := matchGlobPattern(tt.pattern, tt.name)
			if result != tt.match {
				t.Errorf("matchGlobPattern(%q, %q) = %v, want %v", tt.pattern, tt.name, result, tt.match)
			}
		})
	}
}

func TestScannerNoTestsFilter(t *testing.T) {
	tempDir := t.TempDir()

	files := map[string]string{
		"main.go":       "package main",
		"main_test.go":  "package main",
		"utils.go":      "package main",
		"utils_test.go": "package main",
		"app.py":        "print('hi')",
		"test_app.py":   "import unittest",
	}
	createTestFiles(t, tempDir, files)

	config := &Config{
		NoTests: true,
	}
	scanner := NewScanner(config)

	result, err := scanner.Scan(tempDir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	relPaths := getRelativePaths(t, tempDir, result)

	expected := []string{"app.py", "main.go", "utils.go"}
	sort.Strings(expected)

	if len(relPaths) != len(expected) {
		t.Errorf("Expected %d files, got %d", len(expected), len(relPaths))
		t.Errorf("Expected: %v", expected)
		t.Errorf("Got: %v", relPaths)
		return
	}

	for i, path := range expected {
		if relPaths[i] != path {
			t.Errorf("Expected %s at position %d, got %s", path, i, relPaths[i])
		}
	}
}

func TestScannerTestsOnlyFilter(t *testing.T) {
	tempDir := t.TempDir()

	files := map[string]string{
		"main.go":       "package main",
		"main_test.go":  "package main",
		"utils.go":      "package main",
		"utils_test.go": "package main",
		"app.py":        "print('hi')",
		"test_app.py":   "import unittest",
	}
	createTestFiles(t, tempDir, files)

	config := &Config{
		TestsOnly: true,
	}
	scanner := NewScanner(config)

	result, err := scanner.Scan(tempDir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	relPaths := getRelativePaths(t, tempDir, result)

	expected := []string{"main_test.go", "test_app.py", "utils_test.go"}
	sort.Strings(expected)

	if len(relPaths) != len(expected) {
		t.Errorf("Expected %d files, got %d", len(expected), len(relPaths))
		t.Errorf("Expected: %v", expected)
		t.Errorf("Got: %v", relPaths)
		return
	}

	for i, path := range expected {
		if relPaths[i] != path {
			t.Errorf("Expected %s at position %d, got %s", path, i, relPaths[i])
		}
	}
}

func TestScannerLanguageFilter(t *testing.T) {
	tempDir := t.TempDir()

	files := map[string]string{
		"main.go":      "package main",
		"main_test.go": "package main",
		"utils.go":     "package main",
		"app.py":       "print('hi')",
		"test_app.py":  "import unittest",
		"index.js":     "console.log('hi')",
	}
	createTestFiles(t, tempDir, files)

	t.Run("single language", func(t *testing.T) {
		config := &Config{
			Languages: []string{"Go"},
		}
		scanner := NewScanner(config)

		result, err := scanner.Scan(tempDir)
		if err != nil {
			t.Fatalf("Scan failed: %v", err)
		}

		relPaths := getRelativePaths(t, tempDir, result)

		expected := []string{"main.go", "main_test.go", "utils.go"}
		sort.Strings(expected)

		if len(relPaths) != len(expected) {
			t.Errorf("Expected %d files, got %d", len(expected), len(relPaths))
			t.Errorf("Expected: %v", expected)
			t.Errorf("Got: %v", relPaths)
		}
	})

	t.Run("multiple languages", func(t *testing.T) {
		config := &Config{
			Languages: []string{"Go", "Python"},
		}
		scanner := NewScanner(config)

		result, err := scanner.Scan(tempDir)
		if err != nil {
			t.Fatalf("Scan failed: %v", err)
		}

		relPaths := getRelativePaths(t, tempDir, result)

		expected := []string{"app.py", "main.go", "main_test.go", "test_app.py", "utils.go"}
		sort.Strings(expected)

		if len(relPaths) != len(expected) {
			t.Errorf("Expected %d files, got %d", len(expected), len(relPaths))
			t.Errorf("Expected: %v", expected)
			t.Errorf("Got: %v", relPaths)
		}
	})

	t.Run("case insensitive", func(t *testing.T) {
		config := &Config{
			Languages: []string{"go"},
		}
		scanner := NewScanner(config)

		result, err := scanner.Scan(tempDir)
		if err != nil {
			t.Fatalf("Scan failed: %v", err)
		}

		relPaths := getRelativePaths(t, tempDir, result)

		expected := []string{"main.go", "main_test.go", "utils.go"}
		sort.Strings(expected)

		if len(relPaths) != len(expected) {
			t.Errorf("Expected %d files, got %d", len(expected), len(relPaths))
			t.Errorf("Expected: %v", expected)
			t.Errorf("Got: %v", relPaths)
		}
	})
}

func TestScannerExcludeLanguageFilter(t *testing.T) {
	tempDir := t.TempDir()

	files := map[string]string{
		"main.go":  "package main",
		"utils.go": "package main",
		"app.py":   "print('hi')",
		"index.js": "console.log('hi')",
	}
	createTestFiles(t, tempDir, files)

	config := &Config{
		ExcludeLangs: []string{"Go"},
	}
	scanner := NewScanner(config)

	result, err := scanner.Scan(tempDir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	relPaths := getRelativePaths(t, tempDir, result)

	expected := []string{"app.py", "index.js"}
	sort.Strings(expected)

	if len(relPaths) != len(expected) {
		t.Errorf("Expected %d files, got %d", len(expected), len(relPaths))
		t.Errorf("Expected: %v", expected)
		t.Errorf("Got: %v", relPaths)
	}
}

func TestScannerCombinedFilters(t *testing.T) {
	tempDir := t.TempDir()

	files := map[string]string{
		"main.go":       "package main",
		"main_test.go":  "package main",
		"utils.go":      "package main",
		"utils_test.go": "package main",
		"app.py":        "print('hi')",
		"test_app.py":   "import unittest",
		"index.js":      "console.log('hi')",
	}
	createTestFiles(t, tempDir, files)

	// Filter: Go only, no tests
	config := &Config{
		Languages: []string{"Go"},
		NoTests:   true,
	}
	scanner := NewScanner(config)

	result, err := scanner.Scan(tempDir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	relPaths := getRelativePaths(t, tempDir, result)

	expected := []string{"main.go", "utils.go"}
	sort.Strings(expected)

	if len(relPaths) != len(expected) {
		t.Errorf("Expected %d files, got %d", len(expected), len(relPaths))
		t.Errorf("Expected: %v", expected)
		t.Errorf("Got: %v", relPaths)
	}
}

func TestScannerIncludeNonSourceFiles(t *testing.T) {
	tempDir := t.TempDir()

	files := map[string]string{
		"main.go":     "package main",
		"readme.md":   "# README",
		"config.json": "{}",
		"data.yaml":   "key: value",
	}
	createTestFiles(t, tempDir, files)

	// Include patterns can include non-source files
	config := &Config{
		Include: []string{"*.md", "*.json"},
	}
	scanner := NewScanner(config)

	result, err := scanner.Scan(tempDir)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	relPaths := getRelativePaths(t, tempDir, result)

	expected := []string{"config.json", "readme.md"}
	sort.Strings(expected)

	if len(relPaths) != len(expected) {
		t.Errorf("Expected %d files, got %d", len(expected), len(relPaths))
		t.Errorf("Expected: %v", expected)
		t.Errorf("Got: %v", relPaths)
	}
}
