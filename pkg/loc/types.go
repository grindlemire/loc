package loc

// Language defines a programming language with its comment syntax
type Language struct {
	Name              string   // e.g., "Go", "Python", "Templ"
	Extensions        []string // e.g., [".go"], [".py", ".pyw"], [".templ"]
	Filenames         []string // e.g., ["Makefile", "Dockerfile"] - specific filenames without extensions
	LineComment       string   // e.g., "//", "#"
	BlockCommentStart string   // e.g., "/*", "\"\"\""
	BlockCommentEnd   string   // e.g., "*/", "\"\"\""
}

// FileCategory represents the category of a source file
type FileCategory int

const (
	FileCategorySrc   FileCategory = iota // Source code files
	FileCategoryTest                      // Test files
	FileCategoryOther                     // Config, docs, etc. (only with --all)
)

// FileResult holds line counting results for a single file
type FileResult struct {
	Path         string
	Language     string
	Package      string // Derived from directory or language-specific detection
	Lines        int    // Total lines
	CodeLines    int    // Non-blank, non-comment lines
	BlankLines   int
	CommentLines int
	Category     FileCategory // File category (src, test, or other)
}

// Summary aggregates results across all counted files
type Summary struct {
	TotalFiles    int
	TotalLines    int
	TotalCode     int
	TotalBlanks   int
	TotalComments int
	ByLanguage    map[string]*LanguageStats
	ByDirectory   map[string]*DirectoryStats
	ByPackage     map[string]*PackageStats

	// Source file statistics (non-test, non-other files)
	SrcFiles    int
	SrcLines    int
	SrcCode     int
	SrcBlanks   int
	SrcComments int

	// Test file statistics
	TestFiles    int
	TestLines    int
	TestCode     int
	TestBlanks   int
	TestComments int

	// Other file statistics (config, docs, etc. - only with --all)
	OtherFiles    int
	OtherLines    int
	OtherCode     int
	OtherBlanks   int
	OtherComments int
}

// BaseStats holds common statistics fields shared across language, directory, and package stats.
// It provides an AddFile method to accumulate results from individual files.
type BaseStats struct {
	Files    int
	Lines    int
	Code     int
	Blanks   int
	Comments int

	// Source file sub-breakdown
	SrcFiles    int
	SrcLines    int
	SrcCode     int
	SrcBlanks   int
	SrcComments int

	// Test file sub-breakdown
	TestFiles    int
	TestLines    int
	TestCode     int
	TestBlanks   int
	TestComments int

	// Other file sub-breakdown
	OtherFiles    int
	OtherLines    int
	OtherCode     int
	OtherBlanks   int
	OtherComments int
}

// AddFile accumulates statistics from a FileResult into the BaseStats.
func (s *BaseStats) AddFile(r *FileResult) {
	s.Files++
	s.Lines += r.Lines
	s.Code += r.CodeLines
	s.Blanks += r.BlankLines
	s.Comments += r.CommentLines

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
}

// LanguageStats holds aggregated statistics for a single language
type LanguageStats struct {
	Language string
	BaseStats
}

// DirectoryStats holds aggregated statistics for a single directory
type DirectoryStats struct {
	Path string
	BaseStats
}

// PackageStats holds aggregated statistics for a single package
type PackageStats struct {
	Package string
	BaseStats
}

// Config holds runtime configuration options
type Config struct {
	Workers      int      // Number of parallel workers
	OutputFormat string   // "pretty", "json", "raw"
	ByLanguage   bool     // Show breakdown by language
	ByDirectory  bool     // Show breakdown by directory
	ByPackage    bool     // Show breakdown by package
	CodeOnly     bool     // Only count code lines
	NoGitignore  bool     // Don't respect .gitignore
	Include      []string // Additional glob patterns to include
	Exclude      []string // Glob patterns to exclude
	Combined     bool     // Combine src/test/other into single totals (disable breakdown)
	All          bool     // Include non-source files (config, markdown, etc.)
	NoTests      bool     // Exclude test files from counting
	TestsOnly    bool     // Only include test files
	Languages    []string // Include only files of specified languages (case-insensitive)
	ExcludeLangs []string // Exclude files of specified languages (case-insensitive)
}
