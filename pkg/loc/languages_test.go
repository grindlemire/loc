package loc

import (
	"testing"
)

func TestGetLanguageByExtension(t *testing.T) {
	tests := []struct {
		ext      string
		wantName string
		wantNil  bool
	}{
		// Go
		{".go", "Go", false},
		{".GO", "Go", false}, // case insensitive
		// Templ
		{".templ", "Templ", false},
		// JavaScript
		{".js", "JavaScript", false},
		{".jsx", "JavaScript", false},
		{".mjs", "JavaScript", false},
		{".cjs", "JavaScript", false},
		// TypeScript
		{".ts", "TypeScript", false},
		{".tsx", "TypeScript", false},
		// Python
		{".py", "Python", false},
		{".pyw", "Python", false},
		// Rust
		{".rs", "Rust", false},
		// C
		{".c", "C", false},
		{".h", "C", false},
		// C++
		{".cpp", "C++", false},
		{".hpp", "C++", false},
		{".cc", "C++", false},
		{".cxx", "C++", false},
		// Java
		{".java", "Java", false},
		// Kotlin
		{".kt", "Kotlin", false},
		{".kts", "Kotlin", false},
		// Ruby
		{".rb", "Ruby", false},
		{".rake", "Ruby", false},
		// ERB
		{".erb", "ERB", false},
		// PHP
		{".php", "PHP", false},
		// Blade (note: compound extension handled separately)
		{".blade.php", "Blade", false},
		// Swift
		{".swift", "Swift", false},
		// Shell
		{".sh", "Shell", false},
		{".bash", "Shell", false},
		{".zsh", "Shell", false},
		// SQL
		{".sql", "SQL", false},
		// HTML
		{".html", "HTML", false},
		// CSS
		{".css", "CSS", false},
		// SCSS/Sass
		{".scss", "SCSS", false},
		{".sass", "SCSS", false},
		// Less
		{".less", "Less", false},
		// Vue
		{".vue", "Vue", false},
		// Svelte
		{".svelte", "Svelte", false},
		// Elixir
		{".ex", "Elixir", false},
		{".exs", "Elixir", false},
		// EEx
		{".eex", "EEx", false},
		{".heex", "EEx", false},
		{".leex", "EEx", false},
		// Haskell
		{".hs", "Haskell", false},
		{".lhs", "Haskell", false},
		// OCaml
		{".ml", "OCaml", false},
		{".mli", "OCaml", false},
		// F#
		{".fs", "F#", false},
		{".fsx", "F#", false},
		// Zig
		{".zig", "Zig", false},
		// Nim
		{".nim", "Nim", false},
		// Julia
		{".jl", "Julia", false},
		// Dart
		{".dart", "Dart", false},
		// Lua
		{".lua", "Lua", false},
		// R
		{".r", "R", false},
		{".R", "R", false},
		// Scala
		{".scala", "Scala", false},
		{".sc", "Scala", false},
		// Clojure
		{".clj", "Clojure", false},
		{".cljs", "Clojure", false},
		{".cljc", "Clojure", false},
		{".edn", "Clojure", false},
		// Erlang
		{".erl", "Erlang", false},
		{".hrl", "Erlang", false},
		// Unknown extensions
		{".unknown", "", true},
		{".xyz", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.ext, func(t *testing.T) {
			got := GetLanguageByExtension(tt.ext)
			if tt.wantNil {
				if got != nil {
					t.Errorf("GetLanguageByExtension(%q) = %v, want nil", tt.ext, got)
				}
				return
			}
			if got == nil {
				t.Errorf("GetLanguageByExtension(%q) = nil, want %q", tt.ext, tt.wantName)
				return
			}
			if got.Name != tt.wantName {
				t.Errorf("GetLanguageByExtension(%q).Name = %q, want %q", tt.ext, got.Name, tt.wantName)
			}
		})
	}
}

func TestIsSourceFile(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		// Known source files
		{"main.go", true},
		{"app.py", true},
		{"index.js", true},
		{"component.tsx", true},
		{"style.css", true},
		{"template.html", true},
		{"query.sql", true},
		{"script.sh", true},
		// Compound extension
		{"view.blade.php", true},
		{"VIEW.BLADE.PHP", true}, // case insensitive
		// Path with directories
		{"/path/to/file.go", true},
		{"./relative/path/file.rs", true},
		// Unknown extensions
		{"readme.md", false},
		{"data.json", false},
		{"config.yaml", false},
		{"image.png", false},
		{"noextension", false},
		{"", false},
		// Edge cases
		{".go", true},      // just extension
		{".hidden.go", true}, // hidden file with extension
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := IsSourceFile(tt.path)
			if got != tt.want {
				t.Errorf("IsSourceFile(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestAllDocumentedExtensionsRegistered(t *testing.T) {
	// All extensions documented in the design
	documentedExtensions := map[string]string{
		// Go
		".go": "Go",
		// Templ
		".templ": "Templ",
		// JavaScript
		".js":  "JavaScript",
		".jsx": "JavaScript",
		".mjs": "JavaScript",
		".cjs": "JavaScript",
		// TypeScript
		".ts":  "TypeScript",
		".tsx": "TypeScript",
		// Python
		".py":  "Python",
		".pyw": "Python",
		// Rust
		".rs": "Rust",
		// C
		".c": "C",
		".h": "C",
		// C++
		".cpp": "C++",
		".hpp": "C++",
		".cc":  "C++",
		".cxx": "C++",
		// Java
		".java": "Java",
		// Kotlin
		".kt":  "Kotlin",
		".kts": "Kotlin",
		// Ruby
		".rb":   "Ruby",
		".rake": "Ruby",
		// ERB
		".erb": "ERB",
		// PHP
		".php": "PHP",
		// Blade (compound extension)
		".blade.php": "Blade",
		// Swift
		".swift": "Swift",
		// Shell
		".sh":   "Shell",
		".bash": "Shell",
		".zsh":  "Shell",
		// SQL
		".sql": "SQL",
		// HTML
		".html": "HTML",
		// CSS
		".css": "CSS",
		// SCSS/Sass
		".scss": "SCSS",
		".sass": "SCSS",
		// Less
		".less": "Less",
		// Vue
		".vue": "Vue",
		// Svelte
		".svelte": "Svelte",
		// Elixir
		".ex":  "Elixir",
		".exs": "Elixir",
		// EEx
		".eex":  "EEx",
		".heex": "EEx",
		".leex": "EEx",
		// Haskell
		".hs":  "Haskell",
		".lhs": "Haskell",
		// OCaml
		".ml":  "OCaml",
		".mli": "OCaml",
		// F#
		".fs":  "F#",
		".fsx": "F#",
		// Zig
		".zig": "Zig",
		// Nim
		".nim": "Nim",
		// Julia
		".jl": "Julia",
		// Dart
		".dart": "Dart",
		// Lua
		".lua": "Lua",
		// R
		".r": "R",
		".R": "R",
		// Scala
		".scala": "Scala",
		".sc":    "Scala",
		// Clojure
		".clj":  "Clojure",
		".cljs": "Clojure",
		".cljc": "Clojure",
		".edn":  "Clojure",
		// Erlang
		".erl": "Erlang",
		".hrl": "Erlang",
	}

	for ext, expectedLang := range documentedExtensions {
		t.Run(ext, func(t *testing.T) {
			lang := GetLanguageByExtension(ext)
			if lang == nil {
				t.Errorf("Extension %q is not registered, expected language %q", ext, expectedLang)
				return
			}
			if lang.Name != expectedLang {
				t.Errorf("Extension %q maps to %q, expected %q", ext, lang.Name, expectedLang)
			}
		})
	}
}

func TestLanguageCommentSyntax(t *testing.T) {
	// Verify comment syntax for select languages
	tests := []struct {
		ext               string
		lineComment       string
		blockCommentStart string
		blockCommentEnd   string
	}{
		{".go", "//", "/*", "*/"},
		{".py", "#", `"""`, `"""`},
		{".rb", "#", "=begin", "=end"},
		{".html", "", "<!--", "-->"},
		{".css", "", "/*", "*/"},
		{".hs", "--", "{-", "-}"},
		{".ml", "", "(*", "*)"},
		{".nim", "#", "#[", "]#"},
		{".jl", "#", "#=", "=#"},
		{".lua", "--", "--[[", "]]"},
		{".sql", "--", "/*", "*/"},
		{".clj", ";", "", ""},
		{".erl", "%", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.ext, func(t *testing.T) {
			lang := GetLanguageByExtension(tt.ext)
			if lang == nil {
				t.Fatalf("Extension %q not registered", tt.ext)
			}
			if lang.LineComment != tt.lineComment {
				t.Errorf("LineComment = %q, want %q", lang.LineComment, tt.lineComment)
			}
			if lang.BlockCommentStart != tt.blockCommentStart {
				t.Errorf("BlockCommentStart = %q, want %q", lang.BlockCommentStart, tt.blockCommentStart)
			}
			if lang.BlockCommentEnd != tt.blockCommentEnd {
				t.Errorf("BlockCommentEnd = %q, want %q", lang.BlockCommentEnd, tt.blockCommentEnd)
			}
		})
	}
}

func TestGetAllLanguages(t *testing.T) {
	langs := GetAllLanguages()
	if len(langs) == 0 {
		t.Error("GetAllLanguages() returned empty slice")
	}

	// Verify it returns a copy, not the original
	originalLen := len(langs)
	langs[0] = nil
	newLangs := GetAllLanguages()
	if newLangs[0] == nil {
		t.Error("GetAllLanguages() should return a copy, not the original slice")
	}
	if len(newLangs) != originalLen {
		t.Errorf("GetAllLanguages() length changed, got %d, want %d", len(newLangs), originalLen)
	}
}
