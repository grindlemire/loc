package loc

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
)

// Buffer pool for efficient file reading
var bufferPool = sync.Pool{
	New: func() any {
		// 32KB buffer is efficient for most file systems
		buf := make([]byte, 32*1024)
		return &buf
	},
}

// binaryCheckSize is the number of bytes to read for binary file detection
const binaryCheckSize = 8192

// isTestFile determines if a file is a test file based on naming conventions
// across different languages
func isTestFile(path string) bool {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	nameWithoutExt := strings.TrimSuffix(base, ext)
	lowerBase := strings.ToLower(base)
	lowerName := strings.ToLower(nameWithoutExt)

	// Go: *_test.go
	if strings.HasSuffix(lowerBase, "_test.go") {
		return true
	}

	// JavaScript/TypeScript: *.test.js, *.spec.js, *.test.ts, *.spec.ts, etc.
	for _, testExt := range []string{".test.js", ".spec.js", ".test.ts", ".spec.ts",
		".test.jsx", ".spec.jsx", ".test.tsx", ".spec.tsx",
		".test.mjs", ".spec.mjs", ".test.cjs", ".spec.cjs"} {
		if strings.HasSuffix(lowerBase, testExt) {
			return true
		}
	}

	// Python: test_*.py, *_test.py
	if ext == ".py" {
		if strings.HasPrefix(lowerName, "test_") || strings.HasSuffix(lowerName, "_test") {
			return true
		}
	}

	// Ruby: *_spec.rb, *_test.rb
	if ext == ".rb" {
		if strings.HasSuffix(lowerName, "_spec") || strings.HasSuffix(lowerName, "_test") {
			return true
		}
	}

	// Java/Kotlin: *Test.java, *Tests.java, *Test.kt, *Tests.kt
	if ext == ".java" || ext == ".kt" {
		if strings.HasSuffix(nameWithoutExt, "Test") || strings.HasSuffix(nameWithoutExt, "Tests") {
			return true
		}
	}

	// PHP: *Test.php
	if ext == ".php" {
		if strings.HasSuffix(nameWithoutExt, "Test") || strings.HasSuffix(nameWithoutExt, "Tests") {
			return true
		}
	}

	// Rust: Check if in tests/ directory
	if ext == ".rs" && strings.Contains(filepath.ToSlash(path), "/tests/") {
		return true
	}

	// C#: *Test.cs, *Tests.cs
	if ext == ".cs" {
		if strings.HasSuffix(nameWithoutExt, "Test") || strings.HasSuffix(nameWithoutExt, "Tests") {
			return true
		}
	}

	// Swift: *Tests.swift
	if ext == ".swift" {
		if strings.HasSuffix(nameWithoutExt, "Tests") || strings.HasSuffix(nameWithoutExt, "Test") {
			return true
		}
	}

	return false
}

// categorizeFile determines the category of a file (src, test, or other)
func categorizeFile(path string) FileCategory {
	if isTestFile(path) {
		return FileCategoryTest
	}
	if IsOtherFile(path) {
		return FileCategoryOther
	}
	return FileCategorySrc
}

// Counter performs parallel line counting on source files
type Counter struct {
	config *Config
}

// NewCounter creates a new Counter with the provided configuration
func NewCounter(config *Config) *Counter {
	// Set default worker count to CPU count if not specified
	if config.Workers <= 0 {
		config.Workers = runtime.NumCPU()
	}
	return &Counter{
		config: config,
	}
}

// Count processes the given files in parallel and returns aggregated results
func (c *Counter) Count(files []string) (*Summary, error) {
	if len(files) == 0 {
		return newSummary(), nil
	}

	// Create channels for work distribution and result collection
	jobs := make(chan string, len(files))
	results := make(chan *FileResult, len(files))
	var errorCount atomic.Int64

	// Start worker pool
	var wg sync.WaitGroup
	numWorkers := c.config.Workers
	if numWorkers > len(files) {
		numWorkers = len(files)
	}

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range jobs {
				result, err := c.countFile(path)
				if err != nil {
					errorCount.Add(1)
					continue
				}
				if result != nil {
					results <- result
				}
			}
		}()
	}

	// Send all files to workers
	for _, file := range files {
		jobs <- file
	}
	close(jobs)

	// Wait for all workers to finish, then close results channel
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results and aggregate
	summary := newSummary()
	for result := range results {
		summary.addResult(result)
	}

	summary.Errors = int(errorCount.Load())

	return summary, nil
}

// countFile counts lines in a single file using streaming/buffered reading.
// Returns nil result (not error) for files that should be skipped (binary, permission denied).
func (c *Counter) countFile(path string) (*FileResult, error) {
	file, err := os.Open(path)
	if err != nil {
		// Permission denied or other errors - skip file gracefully
		if os.IsPermission(err) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	// Check if file is binary by looking for null bytes in first 8KB
	isBinary, err := c.isBinaryFile(file)
	if err != nil {
		return nil, err
	}
	if isBinary {
		return nil, nil // Skip binary files
	}

	// Seek back to beginning after binary check
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	// Determine language from filename or extension
	lang := GetLanguageForFile(path)
	langName := "Unknown"
	if lang != nil {
		langName = lang.Name
	}

	// Derive package name from parent directory
	pkg := filepath.Base(filepath.Dir(path))

	// Determine file category
	category := categorizeFile(path)

	result := &FileResult{
		Path:     path,
		Language: langName,
		Package:  pkg,
		Category: category,
	}

	// Count lines using streaming reader with buffer pool
	err = c.countLinesStreaming(file, lang, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// isBinaryFile checks if a file is binary by looking for null bytes in the first 8KB
func (c *Counter) isBinaryFile(file *os.File) (bool, error) {
	buf := make([]byte, binaryCheckSize)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return false, err
	}
	if n == 0 {
		return false, nil // Empty file is not binary
	}

	// Check for null bytes (common indicator of binary content)
	return bytes.Contains(buf[:n], []byte{0}), nil
}

// countLinesStreaming counts lines using buffered reading for memory efficiency.
// Handles mixed line endings (LF, CRLF, CR) and files without trailing newline.
func (c *Counter) countLinesStreaming(file *os.File, lang *Language, result *FileResult) error {
	// Get buffer from pool
	bufPtr := bufferPool.Get().(*[]byte)
	buf := *bufPtr
	defer bufferPool.Put(bufPtr)

	var lineBuilder strings.Builder
	inBlockComment := false
	lastCharWasCR := false
	hasContent := false

	for {
		n, err := file.Read(buf)
		if n > 0 {
			hasContent = true
			for i := 0; i < n; i++ {
				ch := buf[i]

				// Handle line endings: LF, CRLF, CR
				if ch == '\n' {
					// LF or CRLF ending
					if lastCharWasCR {
						// CRLF - we already processed CR, skip LF
						lastCharWasCR = false
						continue
					}
					c.processLine(lineBuilder.String(), lang, result, &inBlockComment)
					lineBuilder.Reset()
					lastCharWasCR = false
				} else if ch == '\r' {
					// CR - could be CRLF or old Mac CR-only
					c.processLine(lineBuilder.String(), lang, result, &inBlockComment)
					lineBuilder.Reset()
					lastCharWasCR = true
				} else {
					if lastCharWasCR {
						// Previous CR was standalone (old Mac style)
						lastCharWasCR = false
					}
					lineBuilder.WriteByte(ch)
				}
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	// Handle file without trailing newline
	if lineBuilder.Len() > 0 || lastCharWasCR {
		if lineBuilder.Len() > 0 {
			c.processLine(lineBuilder.String(), lang, result, &inBlockComment)
		}
	} else if !hasContent {
		// Empty file - no lines to count
		return nil
	}

	return nil
}

// processLine classifies and counts a single line
func (c *Counter) processLine(line string, lang *Language, result *FileResult, inBlockComment *bool) {
	result.Lines++

	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		result.BlankLines++
		return
	}

	lineType := c.classifyLine(trimmed, lang, inBlockComment)

	switch lineType {
	case lineTypeBlank:
		result.BlankLines++
	case lineTypeComment:
		result.CommentLines++
	case lineTypeCode:
		result.CodeLines++
	case lineTypeMixed:
		// Mixed lines count as code (contain both code and comments)
		result.CodeLines++
	}
}

// lineType represents the classification of a source line
type lineType int

const (
	lineTypeBlank   lineType = iota
	lineTypeComment          // pure comment line
	lineTypeCode             // pure code line
	lineTypeMixed            // line with both code and comment
)

// classifyLine determines the type of a non-empty line
func (c *Counter) classifyLine(trimmed string, lang *Language, inBlockComment *bool) lineType {
	if lang == nil {
		// Unknown language - treat all non-blank as code
		return lineTypeCode
	}

	// If we're inside a block comment
	if *inBlockComment {
		// Check if block comment ends on this line
		if lang.BlockCommentEnd != "" {
			if endIdx := strings.Index(trimmed, lang.BlockCommentEnd); endIdx >= 0 {
				*inBlockComment = false
				// Check if there's code after the block comment end
				afterComment := strings.TrimSpace(trimmed[endIdx+len(lang.BlockCommentEnd):])
				if afterComment != "" && !c.isLineComment(afterComment, lang) {
					return lineTypeMixed
				}
			}
		}
		return lineTypeComment
	}

	// Check for block comment start
	if lang.BlockCommentStart != "" {
		if startIdx := c.findBlockCommentStart(trimmed, lang); startIdx >= 0 {
			beforeComment := strings.TrimSpace(trimmed[:startIdx])

			// Check if block comment ends on the same line
			afterStart := trimmed[startIdx+len(lang.BlockCommentStart):]
			if endIdx := strings.Index(afterStart, lang.BlockCommentEnd); endIdx >= 0 {
				// Block comment is contained within this line
				afterComment := strings.TrimSpace(afterStart[endIdx+len(lang.BlockCommentEnd):])
				if beforeComment != "" || afterComment != "" {
					return lineTypeMixed
				}
				return lineTypeComment
			}

			// Block comment continues to next line
			*inBlockComment = true
			if beforeComment != "" {
				return lineTypeMixed
			}
			return lineTypeComment
		}
	}

	// Check for single-line comment
	if lang.LineComment != "" {
		if commentIdx := c.findLineComment(trimmed, lang); commentIdx >= 0 {
			if commentIdx == 0 {
				return lineTypeComment
			}
			// There's code before the comment
			return lineTypeMixed
		}
	}

	// No comment found - pure code
	return lineTypeCode
}

// findBlockCommentStart finds the start of a block comment, avoiding false positives in strings
func (c *Counter) findBlockCommentStart(line string, lang *Language) int {
	if lang.BlockCommentStart == "" {
		return -1
	}

	return c.findCommentMarker(line, lang.BlockCommentStart)
}

// findLineComment finds a line comment marker, avoiding false positives in strings
func (c *Counter) findLineComment(line string, lang *Language) int {
	if lang.LineComment == "" {
		return -1
	}

	return c.findCommentMarker(line, lang.LineComment)
}

// findCommentMarker finds a comment marker in a line, avoiding false positives in strings
func (c *Counter) findCommentMarker(line, marker string) int {
	// Special case: if marker is a quote-based marker (like """), check at position 0 first
	// before any string state tracking confuses us
	if strings.HasPrefix(line, marker) {
		return 0
	}

	// Track whether we're in a string to avoid false positives
	inSingleQuote := false
	inDoubleQuote := false
	inBacktick := false
	// Track triple-quote strings for Python
	inTripleDoubleQuote := false
	inTripleSingleQuote := false

	escaped := false
	for i := 0; i < len(line); i++ {
		// Check for escape sequences
		if escaped {
			escaped = false
			continue
		}
		if line[i] == '\\' && (inSingleQuote || inDoubleQuote) {
			escaped = true
			continue
		}

		// Check for triple-quote strings first (Python)
		if i+2 < len(line) {
			substr := line[i : i+3]
			if substr == `"""` {
				if !inSingleQuote && !inBacktick && !inTripleSingleQuote {
					if inTripleDoubleQuote {
						inTripleDoubleQuote = false
						i += 2
						continue
					}
					// Check if this is the comment marker
					if strings.HasPrefix(line[i:], marker) && !inDoubleQuote {
						return i
					}
					inTripleDoubleQuote = true
					i += 2
					continue
				}
			}
			if substr == `'''` {
				if !inDoubleQuote && !inBacktick && !inTripleDoubleQuote {
					if inTripleSingleQuote {
						inTripleSingleQuote = false
						i += 2
						continue
					}
					inTripleSingleQuote = true
					i += 2
					continue
				}
			}
		}

		// Skip if in a triple-quote string
		if inTripleDoubleQuote || inTripleSingleQuote {
			continue
		}

		ch := line[i]

		// Toggle string state
		switch ch {
		case '\'':
			if !inDoubleQuote && !inBacktick {
				inSingleQuote = !inSingleQuote
			}
		case '"':
			if !inSingleQuote && !inBacktick {
				inDoubleQuote = !inDoubleQuote
			}
		case '`':
			if !inSingleQuote && !inDoubleQuote {
				inBacktick = !inBacktick
			}
		}

		// If not in any string, check for comment marker
		if !inSingleQuote && !inDoubleQuote && !inBacktick {
			if strings.HasPrefix(line[i:], marker) {
				return i
			}
		}
	}

	return -1
}

// isLineComment checks if a line is a complete line comment
func (c *Counter) isLineComment(trimmed string, lang *Language) bool {
	if lang == nil || lang.LineComment == "" {
		return false
	}
	return strings.HasPrefix(trimmed, lang.LineComment)
}

// newSummary creates a new empty Summary
func newSummary() *Summary {
	return &Summary{
		ByLanguage:  make(map[string]*LanguageStats),
		ByDirectory: make(map[string]*DirectoryStats),
		ByPackage:   make(map[string]*PackageStats),
	}
}

// MakePathsRelative converts absolute directory paths in ByDirectory to paths
// relative to the given root directory. If a directory equals the root, it
// becomes "." (current directory).
func (s *Summary) MakePathsRelative(root string) {
	if len(s.ByDirectory) == 0 {
		return
	}

	// Create new map with relative paths
	newByDirectory := make(map[string]*DirectoryStats, len(s.ByDirectory))

	for absPath, stats := range s.ByDirectory {
		relPath, err := filepath.Rel(root, absPath)
		if err != nil {
			// If we can't make it relative, keep the original
			relPath = absPath
		}

		// Normalize to forward slashes for consistent display
		relPath = filepath.ToSlash(relPath)

		// Update the stats Path field as well
		stats.Path = relPath
		newByDirectory[relPath] = stats
	}

	s.ByDirectory = newByDirectory
}

// addResult adds a FileResult to the Summary
func (s *Summary) addResult(r *FileResult) {
	s.TotalFiles++
	s.TotalLines += r.Lines
	s.TotalCode += r.CodeLines
	s.TotalBlanks += r.BlankLines
	s.TotalComments += r.CommentLines

	// Update stats based on file category
	switch r.Category {
	case FileCategoryTest:
		s.TestFiles++
		s.TestLines += r.Lines
		s.TestCode += r.CodeLines
		s.TestBlanks += r.BlankLines
		s.TestComments += r.CommentLines
	case FileCategoryOther:
		s.OtherFiles++
		s.OtherLines += r.Lines
		s.OtherCode += r.CodeLines
		s.OtherBlanks += r.BlankLines
		s.OtherComments += r.CommentLines
	default: // FileCategorySrc
		s.SrcFiles++
		s.SrcLines += r.Lines
		s.SrcCode += r.CodeLines
		s.SrcBlanks += r.BlankLines
		s.SrcComments += r.CommentLines
	}

	// Update language stats
	langStats, ok := s.ByLanguage[r.Language]
	if !ok {
		langStats = &LanguageStats{Language: r.Language}
		s.ByLanguage[r.Language] = langStats
	}
	langStats.AddFile(r)

	// Update directory stats
	dir := filepath.Dir(r.Path)
	dirStats, ok := s.ByDirectory[dir]
	if !ok {
		dirStats = &DirectoryStats{Path: dir}
		s.ByDirectory[dir] = dirStats
	}
	dirStats.AddFile(r)

	// Update package stats
	pkgStats, ok := s.ByPackage[r.Package]
	if !ok {
		pkgStats = &PackageStats{Package: r.Package}
		s.ByPackage[r.Package] = pkgStats
	}
	pkgStats.AddFile(r)
}
