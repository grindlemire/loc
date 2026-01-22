package loc

// Language defines a programming language with its comment syntax
type Language struct {
	Name              string   // e.g., "Go", "Python", "Templ"
	Extensions        []string // e.g., [".go"], [".py", ".pyw"], [".templ"]
	LineComment       string   // e.g., "//", "#"
	BlockCommentStart string   // e.g., "/*", "\"\"\""
	BlockCommentEnd   string   // e.g., "*/", "\"\"\""
}

// FileResult holds line counting results for a single file
type FileResult struct {
	Path         string
	Language     string
	Package      string // Derived from directory or language-specific detection
	Lines        int    // Total lines
	CodeLines    int    // Non-blank, non-comment lines
	BlankLines   int
	CommentLines int
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
}

// LanguageStats holds aggregated statistics for a single language
type LanguageStats struct {
	Language string
	Files    int
	Lines    int
	Code     int
	Blanks   int
	Comments int
}

// DirectoryStats holds aggregated statistics for a single directory
type DirectoryStats struct {
	Path     string
	Files    int
	Lines    int
	Code     int
	Blanks   int
	Comments int
}

// PackageStats holds aggregated statistics for a single package
type PackageStats struct {
	Package  string
	Files    int
	Lines    int
	Code     int
	Blanks   int
	Comments int
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
}
