package loc

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestNewCounter(t *testing.T) {
	t.Run("sets default workers", func(t *testing.T) {
		config := &Config{}
		counter := NewCounter(config)
		if counter.config.Workers <= 0 {
			t.Error("expected Workers to be set to a positive value")
		}
	})

	t.Run("respects configured workers", func(t *testing.T) {
		config := &Config{Workers: 4}
		counter := NewCounter(config)
		if counter.config.Workers != 4 {
			t.Errorf("expected Workers=4, got %d", counter.config.Workers)
		}
	})
}

func TestCountEmptyFileList(t *testing.T) {
	counter := NewCounter(&Config{Workers: 2})
	summary, err := counter.Count([]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.TotalFiles != 0 {
		t.Errorf("expected 0 files, got %d", summary.TotalFiles)
	}
}

func TestCountSingleFile(t *testing.T) {
	// Create a temp directory
	tmpDir, err := os.MkdirTemp("", "loc-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a simple Go file
	goFile := filepath.Join(tmpDir, "test.go")
	content := `package main

// This is a comment
func main() {
	println("hello")
}
`
	if err := os.WriteFile(goFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	counter := NewCounter(&Config{Workers: 1})
	summary, err := counter.Count([]string{goFile})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if summary.TotalFiles != 1 {
		t.Errorf("expected 1 file, got %d", summary.TotalFiles)
	}
	if summary.TotalLines != 6 {
		t.Errorf("expected 6 lines, got %d", summary.TotalLines)
	}
	if summary.TotalBlanks != 1 {
		t.Errorf("expected 1 blank line, got %d", summary.TotalBlanks)
	}
	if summary.TotalComments != 1 {
		t.Errorf("expected 1 comment line, got %d", summary.TotalComments)
	}
	if summary.TotalCode != 4 {
		t.Errorf("expected 4 code lines, got %d", summary.TotalCode)
	}
}

func TestBlankLineDetection(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "loc-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// File with various blank lines
	goFile := filepath.Join(tmpDir, "blanks.go")
	content := `package main

func main() {

	x := 1

}

`
	if err := os.WriteFile(goFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	counter := NewCounter(&Config{Workers: 1})
	summary, err := counter.Count([]string{goFile})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if summary.TotalBlanks != 4 {
		t.Errorf("expected 4 blank lines, got %d", summary.TotalBlanks)
	}
}

func TestSingleLineCommentDetection(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "loc-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name     string
		filename string
		content  string
		comments int
		code     int
	}{
		{
			name:     "Go comments",
			filename: "test.go",
			content: `package main
// Comment 1
func main() {} // inline comment
// Comment 2
`,
			comments: 2, // two pure comment lines
			code:     2, // package line + func line with inline comment (mixed counts as code)
		},
		{
			name:     "Python comments",
			filename: "test.py",
			content: `# Comment 1
def main():
    pass  # inline comment
# Comment 2
`,
			comments: 2,
			code:     2,
		},
		{
			name:     "Shell comments",
			filename: "test.sh",
			content: `#!/bin/bash
# Comment
echo "hello"
`,
			comments: 2, // shebang starts with # so treated as comment
			code:     1,
		},
		{
			name:     "SQL comments",
			filename: "test.sql",
			content: `-- Comment
SELECT * FROM users;
-- Another comment
`,
			comments: 2,
			code:     1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			file := filepath.Join(tmpDir, tc.filename)
			if err := os.WriteFile(file, []byte(tc.content), 0644); err != nil {
				t.Fatal(err)
			}

			counter := NewCounter(&Config{Workers: 1})
			summary, err := counter.Count([]string{file})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if summary.TotalComments != tc.comments {
				t.Errorf("expected %d comment lines, got %d", tc.comments, summary.TotalComments)
			}
			if summary.TotalCode != tc.code {
				t.Errorf("expected %d code lines, got %d", tc.code, summary.TotalCode)
			}
		})
	}
}

func TestBlockCommentDetection(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "loc-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name     string
		filename string
		content  string
		comments int
		code     int
	}{
		{
			name:     "C-style block comment",
			filename: "test.go",
			content: `package main
/*
 * Multi-line comment
 */
func main() {}
`,
			comments: 3, // /*, *, */
			code:     2,
		},
		{
			name:     "Block comment on single line",
			filename: "test.go",
			content: `package main
/* single line block comment */
func main() {}
`,
			comments: 1,
			code:     2,
		},
		{
			name:     "Python docstring",
			filename: "test.py",
			content: `"""
Module docstring
"""
def main():
    pass
`,
			comments: 3, // """, Module docstring, """
			code:     2,
		},
		{
			name:     "HTML comment",
			filename: "test.html",
			content: `<!DOCTYPE html>
<!-- Comment -->
<html>
</html>
`,
			comments: 1,
			code:     3,
		},
		{
			name:     "CSS comment",
			filename: "test.css",
			content: `.class {
/* style comment */
    color: red;
}
`,
			comments: 1,
			code:     3,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			file := filepath.Join(tmpDir, tc.filename)
			if err := os.WriteFile(file, []byte(tc.content), 0644); err != nil {
				t.Fatal(err)
			}

			counter := NewCounter(&Config{Workers: 1})
			summary, err := counter.Count([]string{file})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if summary.TotalComments != tc.comments {
				t.Errorf("expected %d comment lines, got %d", tc.comments, summary.TotalComments)
			}
			if summary.TotalCode != tc.code {
				t.Errorf("expected %d code lines, got %d", tc.code, summary.TotalCode)
			}
		})
	}
}

func TestMixedCodeAndCommentLines(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "loc-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test that lines with both code and comments count as code
	goFile := filepath.Join(tmpDir, "mixed.go")
	content := `package main
x := 1 // inline comment
y := 2 /* block */ + 3
`
	if err := os.WriteFile(goFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	counter := NewCounter(&Config{Workers: 1})
	summary, err := counter.Count([]string{goFile})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// All three lines have code, even if some also have comments
	if summary.TotalCode != 3 {
		t.Errorf("expected 3 code lines, got %d", summary.TotalCode)
	}
	if summary.TotalComments != 0 {
		t.Errorf("expected 0 pure comment lines, got %d", summary.TotalComments)
	}
}

func TestStringsContainingCommentMarkers(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "loc-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test that comment markers inside strings are not treated as comments
	goFile := filepath.Join(tmpDir, "strings.go")
	content := `package main
s1 := "// not a comment"
s2 := "/* also not */ a comment"
s3 := ` + "`// backtick string`" + `
`
	if err := os.WriteFile(goFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	counter := NewCounter(&Config{Workers: 1})
	summary, err := counter.Count([]string{goFile})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// All lines are code, no comments
	if summary.TotalCode != 4 {
		t.Errorf("expected 4 code lines, got %d", summary.TotalCode)
	}
	if summary.TotalComments != 0 {
		t.Errorf("expected 0 comment lines, got %d", summary.TotalComments)
	}
}

func TestParallelCountingConsistency(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "loc-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create multiple files
	var files []string
	for i := 0; i < 20; i++ {
		file := filepath.Join(tmpDir, "file"+string(rune('a'+i))+".go")
		content := `package main
// Comment
func main() {}
`
		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
		files = append(files, file)
	}

	// Run counting multiple times with different worker counts
	// All should produce the same result
	var results []*Summary
	for _, workers := range []int{1, 2, 4, 8} {
		counter := NewCounter(&Config{Workers: workers})
		summary, err := counter.Count(files)
		if err != nil {
			t.Fatalf("unexpected error with %d workers: %v", workers, err)
		}
		results = append(results, summary)
	}

	// All results should be identical
	expected := results[0]
	for i, summary := range results[1:] {
		if summary.TotalFiles != expected.TotalFiles {
			t.Errorf("worker count %d: files mismatch %d vs %d", i+2, summary.TotalFiles, expected.TotalFiles)
		}
		if summary.TotalLines != expected.TotalLines {
			t.Errorf("worker count %d: lines mismatch %d vs %d", i+2, summary.TotalLines, expected.TotalLines)
		}
		if summary.TotalCode != expected.TotalCode {
			t.Errorf("worker count %d: code mismatch %d vs %d", i+2, summary.TotalCode, expected.TotalCode)
		}
		if summary.TotalComments != expected.TotalComments {
			t.Errorf("worker count %d: comments mismatch %d vs %d", i+2, summary.TotalComments, expected.TotalComments)
		}
	}
}

func TestConcurrentSafety(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "loc-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create multiple files
	var files []string
	for i := 0; i < 50; i++ {
		file := filepath.Join(tmpDir, "file"+string(rune('a'+i%26))+string(rune('0'+i/26))+".go")
		content := `package main
// Comment line
func foo() {
	x := 1
}
`
		if err := os.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
		files = append(files, file)
	}

	// Run multiple counts concurrently
	var wg sync.WaitGroup
	errors := make(chan error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter := NewCounter(&Config{Workers: 4})
			summary, err := counter.Count(files)
			if err != nil {
				errors <- err
				return
			}
			if summary.TotalFiles != 50 {
				t.Errorf("expected 50 files, got %d", summary.TotalFiles)
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("concurrent error: %v", err)
	}
}

func TestAggregationByLanguage(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "loc-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create files in different languages
	goFile := filepath.Join(tmpDir, "main.go")
	pyFile := filepath.Join(tmpDir, "main.py")
	jsFile := filepath.Join(tmpDir, "main.js")

	os.WriteFile(goFile, []byte("package main\nfunc main() {}\n"), 0644)
	os.WriteFile(pyFile, []byte("def main():\n    pass\n"), 0644)
	os.WriteFile(jsFile, []byte("function main() {}\n"), 0644)

	counter := NewCounter(&Config{Workers: 2})
	summary, err := counter.Count([]string{goFile, pyFile, jsFile})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(summary.ByLanguage) != 3 {
		t.Errorf("expected 3 languages, got %d", len(summary.ByLanguage))
	}

	goStats := summary.ByLanguage["Go"]
	if goStats == nil || goStats.Files != 1 {
		t.Errorf("expected 1 Go file, got %v", goStats)
	}

	pyStats := summary.ByLanguage["Python"]
	if pyStats == nil || pyStats.Files != 1 {
		t.Errorf("expected 1 Python file, got %v", pyStats)
	}

	jsStats := summary.ByLanguage["JavaScript"]
	if jsStats == nil || jsStats.Files != 1 {
		t.Errorf("expected 1 JavaScript file, got %v", jsStats)
	}
}

func TestAggregationByDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "loc-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create directories
	dir1 := filepath.Join(tmpDir, "pkg", "api")
	dir2 := filepath.Join(tmpDir, "pkg", "core")
	os.MkdirAll(dir1, 0755)
	os.MkdirAll(dir2, 0755)

	file1 := filepath.Join(dir1, "api.go")
	file2 := filepath.Join(dir2, "core.go")
	file3 := filepath.Join(dir2, "util.go")

	os.WriteFile(file1, []byte("package api\n"), 0644)
	os.WriteFile(file2, []byte("package core\n"), 0644)
	os.WriteFile(file3, []byte("package core\n"), 0644)

	counter := NewCounter(&Config{Workers: 2})
	summary, err := counter.Count([]string{file1, file2, file3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(summary.ByDirectory) != 2 {
		t.Errorf("expected 2 directories, got %d", len(summary.ByDirectory))
	}

	dir1Stats := summary.ByDirectory[dir1]
	if dir1Stats == nil || dir1Stats.Files != 1 {
		t.Errorf("expected 1 file in api dir, got %v", dir1Stats)
	}

	dir2Stats := summary.ByDirectory[dir2]
	if dir2Stats == nil || dir2Stats.Files != 2 {
		t.Errorf("expected 2 files in core dir, got %v", dir2Stats)
	}
}

func TestAggregationByPackage(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "loc-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create directories that will be used as package names
	apiDir := filepath.Join(tmpDir, "api")
	coreDir := filepath.Join(tmpDir, "core")
	os.MkdirAll(apiDir, 0755)
	os.MkdirAll(coreDir, 0755)

	file1 := filepath.Join(apiDir, "handler.go")
	file2 := filepath.Join(coreDir, "service.go")
	file3 := filepath.Join(coreDir, "model.go")

	os.WriteFile(file1, []byte("package api\n"), 0644)
	os.WriteFile(file2, []byte("package core\n"), 0644)
	os.WriteFile(file3, []byte("package core\n"), 0644)

	counter := NewCounter(&Config{Workers: 2})
	summary, err := counter.Count([]string{file1, file2, file3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(summary.ByPackage) != 2 {
		t.Errorf("expected 2 packages, got %d", len(summary.ByPackage))
	}

	apiStats := summary.ByPackage["api"]
	if apiStats == nil || apiStats.Files != 1 {
		t.Errorf("expected 1 file in api package, got %v", apiStats)
	}

	coreStats := summary.ByPackage["core"]
	if coreStats == nil || coreStats.Files != 2 {
		t.Errorf("expected 2 files in core package, got %v", coreStats)
	}
}

func TestUnreadableFileSkipped(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "loc-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a readable file
	goodFile := filepath.Join(tmpDir, "good.go")
	os.WriteFile(goodFile, []byte("package main\n"), 0644)

	// Non-existent file
	badFile := filepath.Join(tmpDir, "nonexistent.go")

	counter := NewCounter(&Config{Workers: 1})
	summary, err := counter.Count([]string{goodFile, badFile})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should only count the good file
	if summary.TotalFiles != 1 {
		t.Errorf("expected 1 file, got %d", summary.TotalFiles)
	}
}

func TestUnknownLanguage(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "loc-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a file with unknown extension
	file := filepath.Join(tmpDir, "test.xyz")
	content := `line 1
line 2
// looks like a comment but unknown language

line 5
`
	os.WriteFile(file, []byte(content), 0644)

	counter := NewCounter(&Config{Workers: 1})
	summary, err := counter.Count([]string{file})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Unknown language treats all non-blank as code
	if summary.TotalLines != 5 {
		t.Errorf("expected 5 lines, got %d", summary.TotalLines)
	}
	if summary.TotalBlanks != 1 {
		t.Errorf("expected 1 blank, got %d", summary.TotalBlanks)
	}
	if summary.TotalCode != 4 {
		t.Errorf("expected 4 code lines, got %d", summary.TotalCode)
	}
	if summary.TotalComments != 0 {
		t.Errorf("expected 0 comments (unknown language), got %d", summary.TotalComments)
	}
}

func TestNestedBlockComments(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "loc-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test that we handle nested-looking block comments
	// Note: most languages don't support truly nested block comments
	goFile := filepath.Join(tmpDir, "nested.go")
	content := `package main
/*
 * /* inner looks nested but isn't */
 */
func main() {}
`
	os.WriteFile(goFile, []byte(content), 0644)

	counter := NewCounter(&Config{Workers: 1})
	summary, err := counter.Count([]string{goFile})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The third line ends the block comment at first */
	// Fourth line has just */ which is not valid but treated as code
	if summary.TotalCode != 3 { // package, */, func
		t.Errorf("expected 3 code lines, got %d", summary.TotalCode)
	}
}

func TestCodeAfterBlockCommentEnd(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "loc-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	goFile := filepath.Join(tmpDir, "after.go")
	content := `package main
/* comment */ func main() {}
`
	os.WriteFile(goFile, []byte(content), 0644)

	counter := NewCounter(&Config{Workers: 1})
	summary, err := counter.Count([]string{goFile})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Second line is mixed (has both comment and code)
	if summary.TotalCode != 2 {
		t.Errorf("expected 2 code lines, got %d", summary.TotalCode)
	}
	if summary.TotalComments != 0 {
		t.Errorf("expected 0 pure comment lines, got %d", summary.TotalComments)
	}
}

func TestCodeBeforeBlockCommentStart(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "loc-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	goFile := filepath.Join(tmpDir, "before.go")
	content := `package main
x := 1 /* comment
continues here
*/ y := 2
`
	os.WriteFile(goFile, []byte(content), 0644)

	counter := NewCounter(&Config{Workers: 1})
	summary, err := counter.Count([]string{goFile})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Line 1: package (code)
	// Line 2: x := 1 /* comment (mixed = code)
	// Line 3: continues here (comment)
	// Line 4: */ y := 2 (mixed = code)
	if summary.TotalCode != 3 {
		t.Errorf("expected 3 code lines, got %d", summary.TotalCode)
	}
	if summary.TotalComments != 1 {
		t.Errorf("expected 1 comment line, got %d", summary.TotalComments)
	}
}

func TestMakePathsRelative(t *testing.T) {
	tests := []struct {
		name     string
		root     string
		absPaths map[string]*DirectoryStats
		expected map[string]string // key -> expected relative path
	}{
		{
			name: "basic relative paths",
			root: "/Users/joel/projects/loc",
			absPaths: map[string]*DirectoryStats{
				"/Users/joel/projects/loc/pkg/loc": {
					Path:      "/Users/joel/projects/loc/pkg/loc",
					BaseStats: BaseStats{Files: 10},
				},
				"/Users/joel/projects/loc/cmd/loc": {
					Path:      "/Users/joel/projects/loc/cmd/loc",
					BaseStats: BaseStats{Files: 3},
				},
			},
			expected: map[string]string{
				"pkg/loc": "pkg/loc",
				"cmd/loc": "cmd/loc",
			},
		},
		{
			name: "root directory becomes dot",
			root: "/Users/joel/projects/loc",
			absPaths: map[string]*DirectoryStats{
				"/Users/joel/projects/loc": {
					Path:      "/Users/joel/projects/loc",
					BaseStats: BaseStats{Files: 2},
				},
			},
			expected: map[string]string{
				".": ".",
			},
		},
		{
			name: "nested directories",
			root: "/home/user/project",
			absPaths: map[string]*DirectoryStats{
				"/home/user/project/src/api/v1": {
					Path:      "/home/user/project/src/api/v1",
					BaseStats: BaseStats{Files: 5},
				},
				"/home/user/project/src/api/v2": {
					Path:      "/home/user/project/src/api/v2",
					BaseStats: BaseStats{Files: 3},
				},
			},
			expected: map[string]string{
				"src/api/v1": "src/api/v1",
				"src/api/v2": "src/api/v2",
			},
		},
		{
			name:     "empty map",
			root:     "/home/user/project",
			absPaths: map[string]*DirectoryStats{},
			expected: map[string]string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			summary := &Summary{
				ByDirectory: tc.absPaths,
			}

			summary.MakePathsRelative(tc.root)

			// Check that we have the expected number of entries
			if len(summary.ByDirectory) != len(tc.expected) {
				t.Errorf("expected %d directories, got %d", len(tc.expected), len(summary.ByDirectory))
			}

			// Check that each expected path exists and has the correct Path field
			for expectedKey, expectedPath := range tc.expected {
				stats, ok := summary.ByDirectory[expectedKey]
				if !ok {
					t.Errorf("expected key %q not found in ByDirectory", expectedKey)
					continue
				}
				if stats.Path != expectedPath {
					t.Errorf("expected Path field to be %q, got %q", expectedPath, stats.Path)
				}
			}
		})
	}
}

func TestMakePathsRelativePreservesStats(t *testing.T) {
	summary := &Summary{
		ByDirectory: map[string]*DirectoryStats{
			"/home/user/project/pkg/api": {
				Path: "/home/user/project/pkg/api",
				BaseStats: BaseStats{
					Files:    10,
					Lines:    500,
					Code:     400,
					Blanks:   50,
					Comments: 50,
				},
			},
		},
	}

	summary.MakePathsRelative("/home/user/project")

	// Check that stats are preserved
	stats, ok := summary.ByDirectory["pkg/api"]
	if !ok {
		t.Fatal("expected pkg/api in ByDirectory")
	}

	if stats.Files != 10 {
		t.Errorf("expected Files=10, got %d", stats.Files)
	}
	if stats.Lines != 500 {
		t.Errorf("expected Lines=500, got %d", stats.Lines)
	}
	if stats.Code != 400 {
		t.Errorf("expected Code=400, got %d", stats.Code)
	}
	if stats.Blanks != 50 {
		t.Errorf("expected Blanks=50, got %d", stats.Blanks)
	}
	if stats.Comments != 50 {
		t.Errorf("expected Comments=50, got %d", stats.Comments)
	}
}

// TestCategorizeFile tests the categorizeFile function
func TestCategorizeFile(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected FileCategory
	}{
		// Source files
		{"Go source file", "main.go", FileCategorySrc},
		{"Python source file", "app.py", FileCategorySrc},
		{"JavaScript source file", "index.js", FileCategorySrc},
		{"TypeScript source file", "component.tsx", FileCategorySrc},
		{"Java source file", "App.java", FileCategorySrc},

		// Test files
		{"Go test file", "main_test.go", FileCategoryTest},
		{"Python test file prefix", "test_app.py", FileCategoryTest},
		{"Python test file suffix", "app_test.py", FileCategoryTest},
		{"JavaScript test file", "app.test.js", FileCategoryTest},
		{"TypeScript spec file", "component.spec.tsx", FileCategoryTest},
		{"Java test file", "AppTest.java", FileCategoryTest},
		{"Ruby spec file", "model_spec.rb", FileCategoryTest},

		// Other files (config, docs, build)
		{"YAML config file", "config.yaml", FileCategoryOther},
		{"JSON config file", "package.json", FileCategoryOther},
		{"TOML config file", "Cargo.toml", FileCategoryOther},
		{"Markdown doc file", "README.md", FileCategoryOther},
		{"XML file", "pom.xml", FileCategoryOther},
		{"Makefile", "Makefile", FileCategoryOther},
		{"Dockerfile", "Dockerfile", FileCategoryOther},
		{"INI config file", "config.ini", FileCategoryOther},
		{"Text file", "notes.txt", FileCategoryOther},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := categorizeFile(tc.path)
			if result != tc.expected {
				t.Errorf("categorizeFile(%q) = %v, want %v", tc.path, result, tc.expected)
			}
		})
	}
}

// TestIsOtherFile tests the IsOtherFile function
func TestIsOtherFile(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		// Other files should return true
		{"config.yaml", true},
		{"config.yml", true},
		{"package.json", true},
		{"Cargo.toml", true},
		{"README.md", true},
		{"docs.markdown", true},
		{"config.ini", true},
		{"settings.cfg", true},
		{"pom.xml", true},
		{"app.properties", true},
		{"notes.txt", true},
		{"guide.rst", true},
		{"build.mk", true},
		{"Makefile", true},
		{"Dockerfile", true},

		// Source files should return false
		{"main.go", false},
		{"app.py", false},
		{"index.js", false},
		{"component.tsx", false},
		{"App.java", false},
		{"lib.rs", false},
		{"main.c", false},

		// Test files should return false
		{"main_test.go", false},
		{"test_app.py", false},
		{"app.test.js", false},
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			result := IsOtherFile(tc.path)
			if result != tc.expected {
				t.Errorf("IsOtherFile(%q) = %v, want %v", tc.path, result, tc.expected)
			}
		})
	}
}

// TestSummaryAddResultCategories tests that addResult correctly categorizes files
func TestSummaryAddResultCategories(t *testing.T) {
	t.Run("source file adds to src stats", func(t *testing.T) {
		summary := newSummary()
		result := &FileResult{
			Path:         "/project/main.go",
			Language:     "Go",
			Package:      "main",
			Lines:        10,
			CodeLines:    8,
			BlankLines:   1,
			CommentLines: 1,
			Category:     FileCategorySrc,
		}

		summary.addResult(result)

		// Check total stats
		if summary.TotalFiles != 1 {
			t.Errorf("expected TotalFiles=1, got %d", summary.TotalFiles)
		}
		if summary.TotalCode != 8 {
			t.Errorf("expected TotalCode=8, got %d", summary.TotalCode)
		}

		// Check src-specific stats
		if summary.SrcFiles != 1 {
			t.Errorf("expected SrcFiles=1, got %d", summary.SrcFiles)
		}
		if summary.SrcCode != 8 {
			t.Errorf("expected SrcCode=8, got %d", summary.SrcCode)
		}
		if summary.SrcLines != 10 {
			t.Errorf("expected SrcLines=10, got %d", summary.SrcLines)
		}

		// Test files should be 0
		if summary.TestFiles != 0 {
			t.Errorf("expected TestFiles=0, got %d", summary.TestFiles)
		}

		// Other files should be 0
		if summary.OtherFiles != 0 {
			t.Errorf("expected OtherFiles=0, got %d", summary.OtherFiles)
		}
	})

	t.Run("test file adds to test stats", func(t *testing.T) {
		summary := newSummary()
		result := &FileResult{
			Path:         "/project/main_test.go",
			Language:     "Go",
			Package:      "main",
			Lines:        20,
			CodeLines:    15,
			BlankLines:   3,
			CommentLines: 2,
			Category:     FileCategoryTest,
		}

		summary.addResult(result)

		// Check total stats
		if summary.TotalFiles != 1 {
			t.Errorf("expected TotalFiles=1, got %d", summary.TotalFiles)
		}
		if summary.TotalCode != 15 {
			t.Errorf("expected TotalCode=15, got %d", summary.TotalCode)
		}

		// Check test-specific stats
		if summary.TestFiles != 1 {
			t.Errorf("expected TestFiles=1, got %d", summary.TestFiles)
		}
		if summary.TestCode != 15 {
			t.Errorf("expected TestCode=15, got %d", summary.TestCode)
		}
		if summary.TestLines != 20 {
			t.Errorf("expected TestLines=20, got %d", summary.TestLines)
		}

		// Src files should be 0
		if summary.SrcFiles != 0 {
			t.Errorf("expected SrcFiles=0, got %d", summary.SrcFiles)
		}
	})

	t.Run("other file adds to other stats", func(t *testing.T) {
		summary := newSummary()
		result := &FileResult{
			Path:         "/project/config.yaml",
			Language:     "YAML",
			Package:      "project",
			Lines:        30,
			CodeLines:    25,
			BlankLines:   5,
			CommentLines: 0,
			Category:     FileCategoryOther,
		}

		summary.addResult(result)

		// Check total stats
		if summary.TotalFiles != 1 {
			t.Errorf("expected TotalFiles=1, got %d", summary.TotalFiles)
		}
		if summary.TotalCode != 25 {
			t.Errorf("expected TotalCode=25, got %d", summary.TotalCode)
		}

		// Check other-specific stats
		if summary.OtherFiles != 1 {
			t.Errorf("expected OtherFiles=1, got %d", summary.OtherFiles)
		}
		if summary.OtherCode != 25 {
			t.Errorf("expected OtherCode=25, got %d", summary.OtherCode)
		}
		if summary.OtherLines != 30 {
			t.Errorf("expected OtherLines=30, got %d", summary.OtherLines)
		}

		// Src and test files should be 0
		if summary.SrcFiles != 0 {
			t.Errorf("expected SrcFiles=0, got %d", summary.SrcFiles)
		}
		if summary.TestFiles != 0 {
			t.Errorf("expected TestFiles=0, got %d", summary.TestFiles)
		}
	})

	t.Run("mixed file categories", func(t *testing.T) {
		summary := newSummary()

		// Add a source file
		summary.addResult(&FileResult{
			Path:      "/project/main.go",
			Language:  "Go",
			Package:   "main",
			Lines:     10,
			CodeLines: 8,
			Category:  FileCategorySrc,
		})

		// Add a test file
		summary.addResult(&FileResult{
			Path:      "/project/main_test.go",
			Language:  "Go",
			Package:   "main",
			Lines:     20,
			CodeLines: 15,
			Category:  FileCategoryTest,
		})

		// Add an other file
		summary.addResult(&FileResult{
			Path:      "/project/config.yaml",
			Language:  "YAML",
			Package:   "project",
			Lines:     5,
			CodeLines: 4,
			Category:  FileCategoryOther,
		})

		// Check total stats
		if summary.TotalFiles != 3 {
			t.Errorf("expected TotalFiles=3, got %d", summary.TotalFiles)
		}
		if summary.TotalCode != 27 { // 8 + 15 + 4
			t.Errorf("expected TotalCode=27, got %d", summary.TotalCode)
		}
		if summary.TotalLines != 35 { // 10 + 20 + 5
			t.Errorf("expected TotalLines=35, got %d", summary.TotalLines)
		}

		// Check category-specific stats
		if summary.SrcFiles != 1 {
			t.Errorf("expected SrcFiles=1, got %d", summary.SrcFiles)
		}
		if summary.SrcCode != 8 {
			t.Errorf("expected SrcCode=8, got %d", summary.SrcCode)
		}
		if summary.TestFiles != 1 {
			t.Errorf("expected TestFiles=1, got %d", summary.TestFiles)
		}
		if summary.TestCode != 15 {
			t.Errorf("expected TestCode=15, got %d", summary.TestCode)
		}
		if summary.OtherFiles != 1 {
			t.Errorf("expected OtherFiles=1, got %d", summary.OtherFiles)
		}
		if summary.OtherCode != 4 {
			t.Errorf("expected OtherCode=4, got %d", summary.OtherCode)
		}
	})
}

// TestLanguageStatsSubBreakdown tests that language stats track src/test/other breakdown
func TestLanguageStatsSubBreakdown(t *testing.T) {
	summary := newSummary()

	// Add Go source file
	summary.addResult(&FileResult{
		Path:      "/project/main.go",
		Language:  "Go",
		Package:   "main",
		Lines:     100,
		CodeLines: 80,
		Category:  FileCategorySrc,
	})

	// Add Go test file
	summary.addResult(&FileResult{
		Path:      "/project/main_test.go",
		Language:  "Go",
		Package:   "main",
		Lines:     50,
		CodeLines: 40,
		Category:  FileCategoryTest,
	})

	// Check language stats
	goStats := summary.ByLanguage["Go"]
	if goStats == nil {
		t.Fatal("expected Go language stats")
	}

	// Total for language
	if goStats.Files != 2 {
		t.Errorf("expected Go Files=2, got %d", goStats.Files)
	}
	if goStats.Code != 120 { // 80 + 40
		t.Errorf("expected Go Code=120, got %d", goStats.Code)
	}

	// Src sub-breakdown
	if goStats.SrcFiles != 1 {
		t.Errorf("expected Go SrcFiles=1, got %d", goStats.SrcFiles)
	}
	if goStats.SrcCode != 80 {
		t.Errorf("expected Go SrcCode=80, got %d", goStats.SrcCode)
	}

	// Test sub-breakdown
	if goStats.TestFiles != 1 {
		t.Errorf("expected Go TestFiles=1, got %d", goStats.TestFiles)
	}
	if goStats.TestCode != 40 {
		t.Errorf("expected Go TestCode=40, got %d", goStats.TestCode)
	}

	// Other should be 0
	if goStats.OtherFiles != 0 {
		t.Errorf("expected Go OtherFiles=0, got %d", goStats.OtherFiles)
	}
}

// TestDirectoryStatsSubBreakdown tests that directory stats track src/test/other breakdown
func TestDirectoryStatsSubBreakdown(t *testing.T) {
	summary := newSummary()

	// Add source file in pkg directory
	summary.addResult(&FileResult{
		Path:      "/project/pkg/api.go",
		Language:  "Go",
		Package:   "pkg",
		Lines:     100,
		CodeLines: 80,
		Category:  FileCategorySrc,
	})

	// Add test file in same directory
	summary.addResult(&FileResult{
		Path:      "/project/pkg/api_test.go",
		Language:  "Go",
		Package:   "pkg",
		Lines:     50,
		CodeLines: 40,
		Category:  FileCategoryTest,
	})

	// Check directory stats
	dirStats := summary.ByDirectory["/project/pkg"]
	if dirStats == nil {
		t.Fatal("expected /project/pkg directory stats")
	}

	// Total for directory
	if dirStats.Files != 2 {
		t.Errorf("expected dir Files=2, got %d", dirStats.Files)
	}
	if dirStats.Code != 120 { // 80 + 40
		t.Errorf("expected dir Code=120, got %d", dirStats.Code)
	}

	// Src sub-breakdown
	if dirStats.SrcFiles != 1 {
		t.Errorf("expected dir SrcFiles=1, got %d", dirStats.SrcFiles)
	}
	if dirStats.SrcCode != 80 {
		t.Errorf("expected dir SrcCode=80, got %d", dirStats.SrcCode)
	}

	// Test sub-breakdown
	if dirStats.TestFiles != 1 {
		t.Errorf("expected dir TestFiles=1, got %d", dirStats.TestFiles)
	}
	if dirStats.TestCode != 40 {
		t.Errorf("expected dir TestCode=40, got %d", dirStats.TestCode)
	}
}

// TestPackageStatsSubBreakdown tests that package stats track src/test/other breakdown
func TestPackageStatsSubBreakdown(t *testing.T) {
	summary := newSummary()

	// Add source file
	summary.addResult(&FileResult{
		Path:      "/project/pkg/api.go",
		Language:  "Go",
		Package:   "api",
		Lines:     100,
		CodeLines: 80,
		Category:  FileCategorySrc,
	})

	// Add test file in same package
	summary.addResult(&FileResult{
		Path:      "/project/pkg/api_test.go",
		Language:  "Go",
		Package:   "api",
		Lines:     50,
		CodeLines: 40,
		Category:  FileCategoryTest,
	})

	// Check package stats
	pkgStats := summary.ByPackage["api"]
	if pkgStats == nil {
		t.Fatal("expected api package stats")
	}

	// Total for package
	if pkgStats.Files != 2 {
		t.Errorf("expected pkg Files=2, got %d", pkgStats.Files)
	}
	if pkgStats.Code != 120 { // 80 + 40
		t.Errorf("expected pkg Code=120, got %d", pkgStats.Code)
	}

	// Src sub-breakdown
	if pkgStats.SrcFiles != 1 {
		t.Errorf("expected pkg SrcFiles=1, got %d", pkgStats.SrcFiles)
	}
	if pkgStats.SrcCode != 80 {
		t.Errorf("expected pkg SrcCode=80, got %d", pkgStats.SrcCode)
	}

	// Test sub-breakdown
	if pkgStats.TestFiles != 1 {
		t.Errorf("expected pkg TestFiles=1, got %d", pkgStats.TestFiles)
	}
	if pkgStats.TestCode != 40 {
		t.Errorf("expected pkg TestCode=40, got %d", pkgStats.TestCode)
	}
}

// TestFileCategoryConstants tests that FileCategory constants have expected values
func TestFileCategoryConstants(t *testing.T) {
	// Verify the constants are defined with expected values
	if FileCategorySrc != 0 {
		t.Errorf("expected FileCategorySrc=0, got %d", FileCategorySrc)
	}
	if FileCategoryTest != 1 {
		t.Errorf("expected FileCategoryTest=1, got %d", FileCategoryTest)
	}
	if FileCategoryOther != 2 {
		t.Errorf("expected FileCategoryOther=2, got %d", FileCategoryOther)
	}
}

func TestFindCommentMarkerEscapeHandling(t *testing.T) {
	c := &Counter{}

	tests := []struct {
		name     string
		line     string
		marker   string
		expected int
	}{
		{
			name:     "double backslash before closing quote then comment",
			line:     `x = "hello\\" // comment`,
			marker:   "//",
			expected: 14, // the // after the string
		},
		{
			name:     "single escaped quote inside string no comment",
			line:     `x = "he\"llo"`,
			marker:   "//",
			expected: -1,
		},
		{
			name:     "escaped backslash at end of string then comment",
			line:     `fmt.Println("path\\") // trailing`,
			marker:   "//",
			expected: 22,
		},
		{
			name:     "triple backslash before quote (backslash + escaped quote)",
			line:     `x = "test\\\"still" // comment`,
			marker:   "//",
			expected: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := c.findCommentMarker(tt.line, tt.marker)
			if result != tt.expected {
				t.Errorf("findCommentMarker(%q, %q) = %d, want %d", tt.line, tt.marker, result, tt.expected)
			}
		})
	}
}

func TestCountTracksErrors(t *testing.T) {
	dir := t.TempDir()

	// Make a valid file
	validPath := filepath.Join(dir, "valid.go")
	err := os.WriteFile(validPath, []byte("package main\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	config := &Config{Workers: 1}
	counter := NewCounter(config)

	// Count both files — the nonexistent one will error but not crash
	summary, err := counter.Count([]string{validPath, "/nonexistent/file.go"})
	if err != nil {
		t.Fatalf("Count should not return error for per-file failures: %v", err)
	}

	// Valid file should be counted
	if summary.TotalFiles != 1 {
		t.Errorf("expected 1 file counted, got %d", summary.TotalFiles)
	}

	// Should track that 1 file had errors
	if summary.Errors != 1 {
		t.Errorf("expected 1 error tracked, got %d", summary.Errors)
	}
}
