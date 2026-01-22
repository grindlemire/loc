package loc

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuiltinExclusions(t *testing.T) {
	// Create a minimal GitIgnore with no custom patterns
	g := &GitIgnore{
		root:     ".",
		patterns: []*Pattern{},
	}

	tests := []struct {
		path   string
		ignore bool
		desc   string
	}{
		// Binary extensions
		{"program.exe", true, "binary .exe"},
		{"lib.dll", true, "binary .dll"},
		{"lib.so", true, "binary .so"},
		{"lib.dylib", true, "binary .dylib"},
		{"data.bin", true, "binary .bin"},
		{"main.o", true, "binary .o"},
		{"archive.a", true, "binary .a"},
		{"module.wasm", true, "binary .wasm"},

		// Compiled/minified JS
		{"jquery.min.js", true, "minified JS"},
		{"app.bundle.js", true, "bundled JS"},
		{"styles.css.map", true, "source map"},
		{"app.js.map", true, "JS source map"},

		// Generated files
		{"proto.pb.go", true, "protobuf generated"},
		{"types_generated.go", true, "generated go"},
		{"model.g.dart", true, "generated dart"},
		{"component_templ.go", true, "templ generated"},

		// Lock files
		{"package-lock.json", true, "npm lock"},
		{"yarn.lock", true, "yarn lock"},
		{"go.sum", true, "go checksum"},
		{"Cargo.lock", true, "cargo lock"},
		{"mix.lock", true, "mix lock"},

		// Excluded directories
		{"node_modules/package/index.js", true, "node_modules dir"},
		{"vendor/lib/lib.go", true, "vendor dir"},
		{".git/config", true, ".git dir"},
		{"dist/bundle.js", true, "dist dir"},
		{"build/output.js", true, "build dir"},
		{"__pycache__/module.pyc", true, "pycache dir"},
		{"target/debug/main", true, "target dir"},
		{"_build/dev/lib", true, "_build dir"},

		// Valid source files
		{"main.go", false, "go source"},
		{"app.js", false, "js source"},
		{"script.py", false, "python source"},
		{"lib.rs", false, "rust source"},
		{"src/main.go", false, "nested go source"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			result := g.ShouldIgnore(tt.path)
			if result != tt.ignore {
				t.Errorf("ShouldIgnore(%q) = %v, want %v", tt.path, result, tt.ignore)
			}
		})
	}
}

func TestGitIgnorePatterns(t *testing.T) {
	// Create test directory structure
	tempDir := t.TempDir()

	// Create .gitignore file
	gitignore := `
# Comment line
*.log
debug/
!important.log
temp*.txt
build.sh
`
	err := os.WriteFile(filepath.Join(tempDir, ".gitignore"), []byte(gitignore), 0644)
	if err != nil {
		t.Fatalf("Failed to create .gitignore: %v", err)
	}

	g, err := LoadGitIgnore(tempDir)
	if err != nil {
		t.Fatalf("LoadGitIgnore failed: %v", err)
	}

	tests := []struct {
		path   string
		ignore bool
		desc   string
	}{
		// Basic patterns
		{"test.log", true, "*.log pattern"},
		{"app.log", true, "*.log pattern 2"},
		{"nested/test.log", true, "nested *.log"},

		// Directory patterns
		{"debug/file.txt", true, "debug/ directory"},
		{"debug/nested/file.txt", true, "debug/ nested"},

		// Negation patterns
		{"important.log", false, "!important.log negation"},

		// Wildcard patterns
		{"temp123.txt", true, "temp*.txt pattern"},
		{"temporary.txt", true, "temp*.txt pattern 2"},

		// Exact filename
		{"build.sh", true, "exact filename match"},
		{"src/build.sh", true, "nested exact filename"},

		// Files that should NOT be ignored
		{"main.go", false, "normal go file"},
		{"README.md", false, "readme file"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			result := g.ShouldIgnore(tt.path)
			if result != tt.ignore {
				t.Errorf("ShouldIgnore(%q) = %v, want %v", tt.path, result, tt.ignore)
			}
		})
	}
}

func TestNestedGitIgnore(t *testing.T) {
	// Create test directory structure
	tempDir := t.TempDir()

	// Create root .gitignore
	rootGitignore := `
*.log
`
	err := os.WriteFile(filepath.Join(tempDir, ".gitignore"), []byte(rootGitignore), 0644)
	if err != nil {
		t.Fatalf("Failed to create root .gitignore: %v", err)
	}

	// Create subdirectory with its own .gitignore
	subDir := filepath.Join(tempDir, "src")
	err = os.MkdirAll(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create src dir: %v", err)
	}

	subGitignore := `
*.tmp
!keep.tmp
`
	err = os.WriteFile(filepath.Join(subDir, ".gitignore"), []byte(subGitignore), 0644)
	if err != nil {
		t.Fatalf("Failed to create src .gitignore: %v", err)
	}

	g, err := LoadGitIgnore(tempDir)
	if err != nil {
		t.Fatalf("LoadGitIgnore failed: %v", err)
	}

	tests := []struct {
		path   string
		ignore bool
		desc   string
	}{
		// Root patterns
		{"test.log", true, "root *.log"},
		{"src/test.log", true, "nested *.log from root"},

		// Nested directory patterns
		{"src/cache.tmp", true, "nested *.tmp"},
		{"src/keep.tmp", false, "nested !keep.tmp negation"},

		// tmp pattern should NOT affect root level
		{"root.tmp", false, "root level tmp not ignored"},

		// Normal files
		{"main.go", false, "normal go file"},
		{"src/main.go", false, "nested go file"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			result := g.ShouldIgnore(tt.path)
			if result != tt.ignore {
				t.Errorf("ShouldIgnore(%q) = %v, want %v", tt.path, result, tt.ignore)
			}
		})
	}
}

func TestDirectoryPatterns(t *testing.T) {
	tempDir := t.TempDir()

	gitignore := `
logs/
cache
`
	err := os.WriteFile(filepath.Join(tempDir, ".gitignore"), []byte(gitignore), 0644)
	if err != nil {
		t.Fatalf("Failed to create .gitignore: %v", err)
	}

	g, err := LoadGitIgnore(tempDir)
	if err != nil {
		t.Fatalf("LoadGitIgnore failed: %v", err)
	}

	tests := []struct {
		path   string
		ignore bool
		desc   string
	}{
		// Directory pattern (ends with /)
		{"logs/debug.txt", true, "logs/ directory content"},
		{"logs/nested/file.txt", true, "logs/ nested content"},

		// Pattern without trailing slash (matches files and dirs)
		{"cache/data.txt", true, "cache as directory"},
		{"cache", true, "cache as file"},

		// Similar names that shouldn't match
		{"mylogger/config", false, "similar name not matched"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			result := g.ShouldIgnore(tt.path)
			if result != tt.ignore {
				t.Errorf("ShouldIgnore(%q) = %v, want %v", tt.path, result, tt.ignore)
			}
		})
	}
}

func TestAbsolutePatterns(t *testing.T) {
	tempDir := t.TempDir()

	gitignore := `
/root-only.txt
/config/secret.yml
nested/specific.txt
`
	err := os.WriteFile(filepath.Join(tempDir, ".gitignore"), []byte(gitignore), 0644)
	if err != nil {
		t.Fatalf("Failed to create .gitignore: %v", err)
	}

	g, err := LoadGitIgnore(tempDir)
	if err != nil {
		t.Fatalf("LoadGitIgnore failed: %v", err)
	}

	tests := []struct {
		path   string
		ignore bool
		desc   string
	}{
		// Absolute pattern (starts with /)
		{"root-only.txt", true, "/root-only.txt at root"},
		{"subdir/root-only.txt", false, "/root-only.txt not in subdir"},

		// Absolute path pattern
		{"config/secret.yml", true, "/config/secret.yml"},
		{"other/config/secret.yml", false, "absolute pattern not matched in subdir"},

		// Pattern with slash (anchored)
		{"nested/specific.txt", true, "nested/specific.txt anchored"},
		{"other/nested/specific.txt", false, "anchored not in other dir"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			result := g.ShouldIgnore(tt.path)
			if result != tt.ignore {
				t.Errorf("ShouldIgnore(%q) = %v, want %v", tt.path, result, tt.ignore)
			}
		})
	}
}

func TestGlobMatch(t *testing.T) {
	tests := []struct {
		pattern string
		name    string
		match   bool
	}{
		// Basic matching
		{"*.go", "main.go", true},
		{"*.go", "test.py", false},
		{"*.go", ".go", true},

		// Question mark
		{"test?.go", "test1.go", true},
		{"test?.go", "test12.go", false},
		{"test?.go", "test.go", false},

		// Multiple wildcards
		{"*.min.*", "app.min.js", true},
		{"*.min.*", "app.min.css", true},
		{"*_test.go", "main_test.go", true},
		{"*_test.go", "test.go", false},

		// Complex patterns
		{"test*file*.txt", "testmyfile123.txt", true},
		{"test*file*.txt", "testfile.txt", true},
		{"test*file*.txt", "mytest.txt", false},

		// Exact match
		{"exact.txt", "exact.txt", true},
		{"exact.txt", "notexact.txt", false},

		// Star matching everything
		{"*", "anything", true},
		{"*", "", true},
		{"prefix*", "prefix", true},
		{"prefix*", "prefixsuffix", true},
		{"*suffix", "suffix", true},
		{"*suffix", "presuffix", true},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.name, func(t *testing.T) {
			result := globMatch(tt.pattern, tt.name)
			if result != tt.match {
				t.Errorf("globMatch(%q, %q) = %v, want %v", tt.pattern, tt.name, result, tt.match)
			}
		})
	}
}

func TestIsExcludedDir(t *testing.T) {
	tests := []struct {
		name     string
		excluded bool
	}{
		{"node_modules", true},
		{"vendor", true},
		{".git", true},
		{"dist", true},
		{"build", true},
		{"__pycache__", true},
		{"target", true},
		{"_build", true},
		{"src", false},
		{"lib", false},
		{"pkg", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsExcludedDir(tt.name)
			if result != tt.excluded {
				t.Errorf("IsExcludedDir(%q) = %v, want %v", tt.name, result, tt.excluded)
			}
		})
	}
}

func TestEmptyGitIgnore(t *testing.T) {
	tempDir := t.TempDir()

	// No .gitignore file
	g, err := LoadGitIgnore(tempDir)
	if err != nil {
		t.Fatalf("LoadGitIgnore failed: %v", err)
	}

	// Should still exclude built-in patterns
	if !g.ShouldIgnore("node_modules/pkg/index.js") {
		t.Error("Should ignore node_modules even without .gitignore")
	}

	if g.ShouldIgnore("main.go") {
		t.Error("Should not ignore main.go")
	}
}

func TestCommentAndEmptyLines(t *testing.T) {
	tempDir := t.TempDir()

	gitignore := `
# This is a comment

# Another comment
*.log

# Trailing comment
`
	err := os.WriteFile(filepath.Join(tempDir, ".gitignore"), []byte(gitignore), 0644)
	if err != nil {
		t.Fatalf("Failed to create .gitignore: %v", err)
	}

	g, err := LoadGitIgnore(tempDir)
	if err != nil {
		t.Fatalf("LoadGitIgnore failed: %v", err)
	}

	// Comments and empty lines should not affect parsing
	if !g.ShouldIgnore("test.log") {
		t.Error("Should ignore *.log")
	}

	if g.ShouldIgnore("main.go") {
		t.Error("Should not ignore main.go")
	}
}

func TestLoadGitIgnoreWithMissingRoot(t *testing.T) {
	_, err := LoadGitIgnore("/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Error("Expected error for nonexistent path")
	}
}

func TestNegationPatterns(t *testing.T) {
	tempDir := t.TempDir()

	gitignore := `
# Ignore all log files
*.log

# But keep important ones
!important.log
!critical.log

# Ignore all in temp, except keep.txt
temp/
!temp/keep.txt
`
	err := os.WriteFile(filepath.Join(tempDir, ".gitignore"), []byte(gitignore), 0644)
	if err != nil {
		t.Fatalf("Failed to create .gitignore: %v", err)
	}

	g, err := LoadGitIgnore(tempDir)
	if err != nil {
		t.Fatalf("LoadGitIgnore failed: %v", err)
	}

	tests := []struct {
		path   string
		ignore bool
		desc   string
	}{
		{"debug.log", true, "normal log ignored"},
		{"important.log", false, "important.log not ignored"},
		{"critical.log", false, "critical.log not ignored"},
		{"other.log", true, "other log ignored"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			result := g.ShouldIgnore(tt.path)
			if result != tt.ignore {
				t.Errorf("ShouldIgnore(%q) = %v, want %v", tt.path, result, tt.ignore)
			}
		})
	}
}
