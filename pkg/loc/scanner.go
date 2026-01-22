package loc

import (
	"os"
	"path/filepath"
	"strings"
)

// Scanner walks directories and finds source files to count
type Scanner struct {
	config    *Config
	gitignore *GitIgnore
}

// NewScanner creates a new Scanner with the provided configuration
func NewScanner(config *Config) *Scanner {
	return &Scanner{
		config: config,
	}
}

// Scan walks the directory tree starting from root and returns a list of source files to count.
// It applies built-in exclusions, gitignore patterns (unless disabled), and user-specified
// include/exclude patterns.
func (s *Scanner) Scan(root string) ([]string, error) {
	// Resolve absolute path
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	// Check if root exists
	info, err := os.Stat(absRoot)
	if err != nil {
		return nil, err
	}

	// If root is a file, check if it's a source file and return it
	if !info.IsDir() {
		if s.isIncluded(absRoot, absRoot) {
			return []string{absRoot}, nil
		}
		return []string{}, nil
	}

	// Load gitignore patterns unless disabled
	if !s.config.NoGitignore {
		s.gitignore, err = LoadGitIgnore(absRoot)
		if err != nil {
			// If we can't load gitignore, continue without it
			s.gitignore = nil
		}
	}

	var files []string

	err = filepath.WalkDir(absRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			// Skip directories we can't read
			if d != nil && d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Get relative path from root for pattern matching
		relPath, err := filepath.Rel(absRoot, path)
		if err != nil {
			return nil
		}

		// Normalize to forward slashes for consistent matching
		relPath = filepath.ToSlash(relPath)

		// Skip the root directory itself
		if relPath == "." {
			return nil
		}

		// Handle directories
		if d.IsDir() {
			// Check if this directory should be skipped
			if s.shouldSkipDir(d.Name(), relPath) {
				return filepath.SkipDir
			}
			return nil
		}

		// Handle files
		if s.isIncluded(path, relPath) {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

// shouldSkipDir returns true if the directory should be skipped entirely
func (s *Scanner) shouldSkipDir(name string, relPath string) bool {
	// Always skip built-in excluded directories
	if IsExcludedDir(name) {
		return true
	}

	// Check gitignore patterns
	if s.gitignore != nil && s.gitignore.ShouldIgnore(relPath) {
		return true
	}

	// Check user exclude patterns for directories
	for _, pattern := range s.config.Exclude {
		// Handle directory patterns (ending with /)
		dirPattern := strings.TrimSuffix(pattern, "/")
		if matchGlobPattern(dirPattern, name) || matchGlobPattern(dirPattern, relPath) {
			return true
		}
	}

	return false
}

// isIncluded returns true if the file should be included in the count
func (s *Scanner) isIncluded(absPath string, relPath string) bool {
	base := filepath.Base(absPath)

	// If user specified include patterns, ONLY include files that match
	if len(s.config.Include) > 0 {
		included := false
		for _, pattern := range s.config.Include {
			if matchGlobPattern(pattern, base) || matchGlobPattern(pattern, relPath) {
				included = true
				break
			}
		}
		if !included {
			return false
		}
	} else {
		// No include patterns: only include recognized source files
		if !IsSourceFile(absPath) {
			return false
		}
	}

	// Check built-in exclusions (via gitignore's shouldIgnoreBuiltin)
	// This is done through ShouldIgnore which checks builtins first
	if s.gitignore != nil {
		if s.gitignore.ShouldIgnore(relPath) {
			return false
		}
	} else {
		// If no gitignore loaded, still check built-in exclusions
		g := &GitIgnore{patterns: []*Pattern{}}
		if g.ShouldIgnore(relPath) {
			return false
		}
	}

	// Check user exclude patterns
	for _, pattern := range s.config.Exclude {
		if matchGlobPattern(pattern, base) || matchGlobPattern(pattern, relPath) {
			return false
		}
	}

	return true
}

// matchGlobPattern matches a glob pattern against a name or path
// Supports *, **, and ? wildcards
func matchGlobPattern(pattern, name string) bool {
	// Handle ** for matching any path depth
	if strings.Contains(pattern, "**") {
		return matchDoubleStarPattern(pattern, name)
	}

	// Simple glob matching using the existing globMatch function
	return globMatch(pattern, name)
}

// matchDoubleStarPattern handles patterns with ** (match any directory depth)
func matchDoubleStarPattern(pattern, path string) bool {
	// Split pattern and path into segments
	patternSegs := strings.Split(pattern, "/")
	pathSegs := strings.Split(path, "/")

	return matchSegments(patternSegs, pathSegs, 0, 0)
}

// matchSegments recursively matches pattern segments against path segments
func matchSegments(pattern, path []string, pi, pai int) bool {
	for pi < len(pattern) && pai < len(path) {
		if pattern[pi] == "**" {
			// ** can match zero or more path segments
			// Try matching zero segments
			if matchSegments(pattern, path, pi+1, pai) {
				return true
			}
			// Try matching one or more segments
			return matchSegments(pattern, path, pi, pai+1)
		}

		if !globMatch(pattern[pi], path[pai]) {
			return false
		}
		pi++
		pai++
	}

	// Handle remaining ** at end of pattern
	for pi < len(pattern) && pattern[pi] == "**" {
		pi++
	}

	return pi == len(pattern) && pai == len(path)
}
