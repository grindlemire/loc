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
