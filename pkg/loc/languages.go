package loc

import (
	"path/filepath"
	"strings"
)

// languages is the registry of all supported programming languages
var languages = []*Language{
	{
		Name:              "Go",
		Extensions:        []string{".go"},
		LineComment:       "//",
		BlockCommentStart: "/*",
		BlockCommentEnd:   "*/",
	},
	{
		Name:              "Templ",
		Extensions:        []string{".templ"},
		LineComment:       "//",
		BlockCommentStart: "/*",
		BlockCommentEnd:   "*/",
	},
	{
		Name:              "JavaScript",
		Extensions:        []string{".js", ".jsx", ".mjs", ".cjs"},
		LineComment:       "//",
		BlockCommentStart: "/*",
		BlockCommentEnd:   "*/",
	},
	{
		Name:              "TypeScript",
		Extensions:        []string{".ts", ".tsx"},
		LineComment:       "//",
		BlockCommentStart: "/*",
		BlockCommentEnd:   "*/",
	},
	{
		Name:              "Python",
		Extensions:        []string{".py", ".pyw"},
		LineComment:       "#",
		BlockCommentStart: `"""`,
		BlockCommentEnd:   `"""`,
	},
	{
		Name:              "Rust",
		Extensions:        []string{".rs"},
		LineComment:       "//",
		BlockCommentStart: "/*",
		BlockCommentEnd:   "*/",
	},
	{
		Name:              "C",
		Extensions:        []string{".c", ".h"},
		LineComment:       "//",
		BlockCommentStart: "/*",
		BlockCommentEnd:   "*/",
	},
	{
		Name:              "C++",
		Extensions:        []string{".cpp", ".hpp", ".cc", ".cxx"},
		LineComment:       "//",
		BlockCommentStart: "/*",
		BlockCommentEnd:   "*/",
	},
	{
		Name:              "Java",
		Extensions:        []string{".java"},
		LineComment:       "//",
		BlockCommentStart: "/*",
		BlockCommentEnd:   "*/",
	},
	{
		Name:              "Kotlin",
		Extensions:        []string{".kt", ".kts"},
		LineComment:       "//",
		BlockCommentStart: "/*",
		BlockCommentEnd:   "*/",
	},
	{
		Name:              "Ruby",
		Extensions:        []string{".rb", ".rake"},
		LineComment:       "#",
		BlockCommentStart: "=begin",
		BlockCommentEnd:   "=end",
	},
	{
		Name:        "ERB",
		Extensions:  []string{".erb"},
		LineComment: "<%#",
		// No standard block comment
	},
	{
		Name:              "PHP",
		Extensions:        []string{".php"},
		LineComment:       "//",
		BlockCommentStart: "/*",
		BlockCommentEnd:   "*/",
	},
	{
		Name:              "Blade",
		Extensions:        []string{".blade.php"},
		LineComment:       "{{--",
		BlockCommentStart: "{{--",
		BlockCommentEnd:   "--}}",
	},
	{
		Name:              "Swift",
		Extensions:        []string{".swift"},
		LineComment:       "//",
		BlockCommentStart: "/*",
		BlockCommentEnd:   "*/",
	},
	{
		Name:        "Shell",
		Extensions:  []string{".sh", ".bash", ".zsh"},
		LineComment: "#",
		// No standard block comment
	},
	{
		Name:              "SQL",
		Extensions:        []string{".sql"},
		LineComment:       "--",
		BlockCommentStart: "/*",
		BlockCommentEnd:   "*/",
	},
	{
		Name:              "HTML",
		Extensions:        []string{".html"},
		LineComment:       "",
		BlockCommentStart: "<!--",
		BlockCommentEnd:   "-->",
	},
	{
		Name:              "CSS",
		Extensions:        []string{".css"},
		LineComment:       "",
		BlockCommentStart: "/*",
		BlockCommentEnd:   "*/",
	},
	{
		Name:              "SCSS",
		Extensions:        []string{".scss", ".sass"},
		LineComment:       "//",
		BlockCommentStart: "/*",
		BlockCommentEnd:   "*/",
	},
	{
		Name:              "Less",
		Extensions:        []string{".less"},
		LineComment:       "//",
		BlockCommentStart: "/*",
		BlockCommentEnd:   "*/",
	},
	{
		Name:              "Vue",
		Extensions:        []string{".vue"},
		LineComment:       "//",
		BlockCommentStart: "/*",
		BlockCommentEnd:   "*/",
	},
	{
		Name:              "Svelte",
		Extensions:        []string{".svelte"},
		LineComment:       "//",
		BlockCommentStart: "/*",
		BlockCommentEnd:   "*/",
	},
	{
		Name:        "Elixir",
		Extensions:  []string{".ex", ".exs"},
		LineComment: "#",
		// No standard block comment
	},
	{
		Name:        "EEx",
		Extensions:  []string{".eex", ".heex", ".leex"},
		LineComment: "<%#",
		// No standard block comment
	},
	{
		Name:              "Haskell",
		Extensions:        []string{".hs", ".lhs"},
		LineComment:       "--",
		BlockCommentStart: "{-",
		BlockCommentEnd:   "-}",
	},
	{
		Name:              "OCaml",
		Extensions:        []string{".ml", ".mli"},
		LineComment:       "",
		BlockCommentStart: "(*",
		BlockCommentEnd:   "*)",
	},
	{
		Name:              "F#",
		Extensions:        []string{".fs", ".fsx"},
		LineComment:       "//",
		BlockCommentStart: "(*",
		BlockCommentEnd:   "*)",
	},
	{
		Name:        "Zig",
		Extensions:  []string{".zig"},
		LineComment: "//",
		// No standard block comment
	},
	{
		Name:              "Nim",
		Extensions:        []string{".nim"},
		LineComment:       "#",
		BlockCommentStart: "#[",
		BlockCommentEnd:   "]#",
	},
	{
		Name:              "Julia",
		Extensions:        []string{".jl"},
		LineComment:       "#",
		BlockCommentStart: "#=",
		BlockCommentEnd:   "=#",
	},
	{
		Name:              "Dart",
		Extensions:        []string{".dart"},
		LineComment:       "//",
		BlockCommentStart: "/*",
		BlockCommentEnd:   "*/",
	},
	{
		Name:              "Lua",
		Extensions:        []string{".lua"},
		LineComment:       "--",
		BlockCommentStart: "--[[",
		BlockCommentEnd:   "]]",
	},
	{
		Name:        "R",
		Extensions:  []string{".r", ".R"},
		LineComment: "#",
		// No standard block comment
	},
	{
		Name:              "Scala",
		Extensions:        []string{".scala", ".sc"},
		LineComment:       "//",
		BlockCommentStart: "/*",
		BlockCommentEnd:   "*/",
	},
	{
		Name:        "Clojure",
		Extensions:  []string{".clj", ".cljs", ".cljc", ".edn"},
		LineComment: ";",
		// No standard block comment
	},
	{
		Name:        "Erlang",
		Extensions:  []string{".erl", ".hrl"},
		LineComment: "%",
		// No standard block comment
	},
}

// extensionMap provides fast lookup from extension to language
var extensionMap map[string]*Language

func init() {
	extensionMap = make(map[string]*Language)
	for _, lang := range languages {
		for _, ext := range lang.Extensions {
			extensionMap[ext] = lang
		}
	}
}

// GetLanguageByExtension returns the Language for a file extension (with leading dot).
// Returns nil if the extension is not recognized.
func GetLanguageByExtension(ext string) *Language {
	// Normalize extension to lowercase
	ext = strings.ToLower(ext)
	return extensionMap[ext]
}

// IsSourceFile returns true if the file extension is a known source file.
func IsSourceFile(path string) bool {
	// Handle compound extensions like .blade.php
	base := filepath.Base(path)
	if strings.HasSuffix(strings.ToLower(base), ".blade.php") {
		return true
	}
	ext := strings.ToLower(filepath.Ext(path))
	if ext == "" {
		return false
	}
	return extensionMap[ext] != nil
}

// GetAllLanguages returns a copy of all registered languages.
func GetAllLanguages() []*Language {
	result := make([]*Language, len(languages))
	copy(result, languages)
	return result
}
