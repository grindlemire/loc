package loc

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// Built-in exclusions from the design
var (
	// Binary file extensions to exclude
	binaryExtensions = map[string]bool{
		".exe":   true,
		".dll":   true,
		".so":    true,
		".dylib": true,
		".bin":   true,
		".o":     true,
		".a":     true,
		".wasm":  true,
	}

	// Compiled/minified JS suffixes
	compiledJSSuffixes = []string{".min.js", ".bundle.js", ".map"}

	// Generated file patterns (suffix matching)
	generatedSuffixes = []string{".pb.go", "_generated.go", ".g.dart", "_templ.go"}

	// Directories to always exclude
	excludedDirs = map[string]bool{
		"node_modules": true,
		"vendor":       true,
		".git":         true,
		"dist":         true,
		"build":        true,
		"__pycache__":  true,
		"target":       true,
		"_build":       true,
	}

	// Lock files to exclude
	lockFiles = map[string]bool{
		"package-lock.json": true,
		"yarn.lock":         true,
		"go.sum":            true,
		"Cargo.lock":        true,
		"mix.lock":          true,
	}
)

// Pattern represents a single gitignore pattern
type Pattern struct {
	pattern   string // The original pattern (for debugging)
	negation  bool   // True if pattern starts with !
	dirOnly   bool   // True if pattern ends with /
	segments  []string
	absolute  bool   // True if pattern starts with /
	baseDir   string // Directory containing the .gitignore file (relative to root)
}

// GitIgnore holds parsed gitignore patterns and provides ignore checking
type GitIgnore struct {
	root     string
	patterns []*Pattern
}

// LoadGitIgnore finds and parses all .gitignore files starting from root
func LoadGitIgnore(root string) (*GitIgnore, error) {
	// Check if root exists
	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, os.ErrNotExist
	}

	g := &GitIgnore{
		root:     root,
		patterns: make([]*Pattern, 0),
	}

	// Walk the directory tree looking for .gitignore files
	err = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			// Skip directories we can't read
			if d != nil && d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Get relative path from root
		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}

		// Skip built-in excluded directories
		if d.IsDir() {
			dirName := d.Name()
			if excludedDirs[dirName] {
				return filepath.SkipDir
			}
		}

		// Parse .gitignore files
		if d.Name() == ".gitignore" && !d.IsDir() {
			baseDir := filepath.Dir(relPath)
			if baseDir == "." {
				baseDir = ""
			}
			patterns, err := parseGitIgnoreFile(path, baseDir)
			if err != nil {
				// Skip unparseable gitignore files
				return nil
			}
			g.patterns = append(g.patterns, patterns...)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return g, nil
}

// parseGitIgnoreFile parses a single .gitignore file and returns its patterns
func parseGitIgnoreFile(path string, baseDir string) ([]*Pattern, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var patterns []*Pattern
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		pattern := parseLine(line, baseDir)
		if pattern != nil {
			patterns = append(patterns, pattern)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return patterns, nil
}

// parseLine parses a single line from a .gitignore file
func parseLine(line string, baseDir string) *Pattern {
	// Trim trailing whitespace (but not leading - leading spaces are significant)
	line = strings.TrimRight(line, " \t\r\n")

	// Skip empty lines and comments
	if line == "" || strings.HasPrefix(line, "#") {
		return nil
	}

	p := &Pattern{
		pattern: line,
		baseDir: baseDir,
	}

	// Check for negation
	if strings.HasPrefix(line, "!") {
		p.negation = true
		line = line[1:]
	}

	// Check for directory-only pattern
	if strings.HasSuffix(line, "/") {
		p.dirOnly = true
		line = strings.TrimSuffix(line, "/")
	}

	// Check for absolute pattern (starts with /)
	if strings.HasPrefix(line, "/") {
		p.absolute = true
		line = line[1:]
	}

	// Also treat patterns with a slash anywhere (except at end) as anchored to baseDir
	// e.g., "foo/bar" should only match "foo/bar" relative to .gitignore location
	if strings.Contains(line, "/") {
		p.absolute = true
	}

	// Split into segments for matching
	p.segments = strings.Split(line, "/")

	return p
}

// ShouldIgnore returns true if the path should be ignored
func (g *GitIgnore) ShouldIgnore(path string) bool {
	// Normalize path separators
	path = filepath.ToSlash(path)

	// Remove leading ./ if present
	path = strings.TrimPrefix(path, "./")

	// Check built-in exclusions first
	if g.shouldIgnoreBuiltin(path) {
		return true
	}

	// Check gitignore patterns (later patterns override earlier ones)
	ignored := false
	for _, p := range g.patterns {
		if p.matches(path) {
			ignored = !p.negation
		}
	}

	return ignored
}

// shouldIgnoreBuiltin checks if a path matches built-in exclusions
func (g *GitIgnore) shouldIgnoreBuiltin(path string) bool {
	// Get the base name and extension
	base := filepath.Base(path)
	ext := strings.ToLower(filepath.Ext(path))

	// Check binary extensions
	if binaryExtensions[ext] {
		return true
	}

	// Check compiled/minified JS suffixes
	lowerBase := strings.ToLower(base)
	for _, suffix := range compiledJSSuffixes {
		if strings.HasSuffix(lowerBase, suffix) {
			return true
		}
	}

	// Check generated file suffixes
	for _, suffix := range generatedSuffixes {
		if strings.HasSuffix(lowerBase, suffix) {
			return true
		}
	}

	// Check lock files
	if lockFiles[base] {
		return true
	}

	// Check if any path component is an excluded directory
	parts := strings.Split(path, "/")
	for _, part := range parts {
		if excludedDirs[part] {
			return true
		}
	}

	return false
}

// matches checks if a pattern matches a given path
func (p *Pattern) matches(path string) bool {
	// Split path into segments
	pathSegments := strings.Split(path, "/")

	// If pattern is anchored to a baseDir, the path must start with baseDir
	searchPath := pathSegments
	if p.baseDir != "" {
		baseParts := strings.Split(p.baseDir, "/")
		if len(pathSegments) < len(baseParts) {
			return false
		}
		for i, bp := range baseParts {
			if pathSegments[i] != bp {
				return false
			}
		}
		// Remove baseDir from searchPath for matching
		searchPath = pathSegments[len(baseParts):]
	}

	patternSegs := p.segments

	if p.absolute {
		// Anchored pattern: must match from the start of searchPath
		return matchFromStart(patternSegs, searchPath, p.dirOnly)
	}

	// Non-anchored pattern: can match at any level
	// For single-segment patterns (like "*.log" or "debug"), try to match any segment
	if len(patternSegs) == 1 {
		// Single segment pattern - match against any path segment
		for i, seg := range searchPath {
			if matchSegment(patternSegs[0], seg) {
				// If dirOnly, this segment must not be the last one (it must be a directory)
				if p.dirOnly {
					if i < len(searchPath)-1 {
						return true
					}
					// Could also be exact match for a directory name
					if i == len(searchPath)-1 {
						return true
					}
				} else {
					return true
				}
			}
		}
		return false
	}

	// Multi-segment pattern without leading slash - try at each position
	for i := 0; i <= len(searchPath)-len(patternSegs); i++ {
		if matchFromStart(patternSegs, searchPath[i:], p.dirOnly) {
			return true
		}
	}

	return false
}

// matchFromStart matches pattern segments against path segments starting from index 0
func matchFromStart(pattern []string, path []string, dirOnly bool) bool {
	if len(pattern) > len(path) {
		return false
	}

	// Match each pattern segment against corresponding path segment
	for i, seg := range pattern {
		if !matchSegment(seg, path[i]) {
			return false
		}
	}

	// If dirOnly and pattern length equals path length, it matches (could be a dir)
	// If dirOnly and pattern is prefix of path, it matches (path is inside that dir)
	// If not dirOnly and pattern length equals path length, it's an exact match
	// If not dirOnly and pattern is prefix, it still matches (file inside matched dir)

	return true
}

// matchSegment matches a single pattern segment against a path segment
// Supports *, **, and ? wildcards
func matchSegment(pattern, name string) bool {
	if pattern == "**" {
		return true
	}
	return globMatch(pattern, name)
}

// globMatch implements glob matching with * and ? wildcards
func globMatch(pattern, name string) bool {
	// Handle empty cases
	if pattern == "" {
		return name == ""
	}

	pi := 0 // pattern index
	ni := 0 // name index
	starPI := -1
	starNI := -1

	for ni < len(name) {
		if pi < len(pattern) {
			switch pattern[pi] {
			case '*':
				// Try to match zero or more characters
				starPI = pi
				starNI = ni
				pi++
				continue
			case '?':
				// Match exactly one character
				pi++
				ni++
				continue
			default:
				if pattern[pi] == name[ni] {
					pi++
					ni++
					continue
				}
			}
		}

		// No match - backtrack to last star if possible
		if starPI >= 0 {
			pi = starPI + 1
			starNI++
			ni = starNI
			continue
		}

		return false
	}

	// Check remaining pattern is all stars
	for pi < len(pattern) {
		if pattern[pi] != '*' {
			return false
		}
		pi++
	}

	return true
}

// IsExcludedDir returns true if the directory name is in the built-in exclusion list
func IsExcludedDir(name string) bool {
	return excludedDirs[name]
}
