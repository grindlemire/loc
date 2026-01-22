package loc

import (
	"bufio"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

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
	errors := make(chan error, len(files))

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
					errors <- err
					continue
				}
				results <- result
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
		close(errors)
	}()

	// Collect results and aggregate
	summary := newSummary()
	for result := range results {
		summary.addResult(result)
	}

	return summary, nil
}

// countFile counts lines in a single file
func (c *Counter) countFile(path string) (*FileResult, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Determine language from extension
	ext := filepath.Ext(path)
	// Handle compound extensions like .blade.php
	base := filepath.Base(path)
	if strings.HasSuffix(strings.ToLower(base), ".blade.php") {
		ext = ".blade.php"
	}

	lang := GetLanguageByExtension(ext)
	langName := "Unknown"
	if lang != nil {
		langName = lang.Name
	}

	// Derive package name from parent directory
	pkg := filepath.Base(filepath.Dir(path))

	result := &FileResult{
		Path:     path,
		Language: langName,
		Package:  pkg,
	}

	scanner := bufio.NewScanner(file)
	inBlockComment := false

	for scanner.Scan() {
		line := scanner.Text()
		result.Lines++

		// Check if blank line
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			result.BlankLines++
			continue
		}

		// Count this line based on comment status
		lineType := c.classifyLine(trimmed, lang, &inBlockComment)

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

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return result, nil
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

	for i := 0; i < len(line); i++ {
		// Check for escape sequences
		if i > 0 && line[i-1] == '\\' {
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

// addResult adds a FileResult to the Summary
func (s *Summary) addResult(r *FileResult) {
	s.TotalFiles++
	s.TotalLines += r.Lines
	s.TotalCode += r.CodeLines
	s.TotalBlanks += r.BlankLines
	s.TotalComments += r.CommentLines

	// Update language stats
	langStats, ok := s.ByLanguage[r.Language]
	if !ok {
		langStats = &LanguageStats{Language: r.Language}
		s.ByLanguage[r.Language] = langStats
	}
	langStats.Files++
	langStats.Lines += r.Lines
	langStats.Code += r.CodeLines
	langStats.Blanks += r.BlankLines
	langStats.Comments += r.CommentLines

	// Update directory stats
	dir := filepath.Dir(r.Path)
	dirStats, ok := s.ByDirectory[dir]
	if !ok {
		dirStats = &DirectoryStats{Path: dir}
		s.ByDirectory[dir] = dirStats
	}
	dirStats.Files++
	dirStats.Lines += r.Lines
	dirStats.Code += r.CodeLines
	dirStats.Blanks += r.BlankLines
	dirStats.Comments += r.CommentLines

	// Update package stats
	pkgStats, ok := s.ByPackage[r.Package]
	if !ok {
		pkgStats = &PackageStats{Package: r.Package}
		s.ByPackage[r.Package] = pkgStats
	}
	pkgStats.Files++
	pkgStats.Lines += r.Lines
	pkgStats.Code += r.CodeLines
	pkgStats.Blanks += r.BlankLines
	pkgStats.Comments += r.CommentLines
}
