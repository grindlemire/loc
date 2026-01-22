# LOC - Lines of Code Counter

A fast, parallel lines-of-code counter for software projects. LOC efficiently counts lines of code, comments, and blank lines across multiple programming languages.

## Features

- **Fast parallel processing** - Uses worker pools to count files concurrently
- **Language-aware** - Distinguishes between code, comments, and blank lines for 30+ languages
- **Gitignore support** - Respects `.gitignore` patterns by default
- **Multiple output formats** - Pretty tables, JSON, or raw text
- **Flexible filtering** - Include/exclude files with glob patterns
- **Smart detection** - Automatically skips binary files

## Installation

```bash
go install github.com/alethi-co/loc/cmd@latest
```

Or build from source:

```bash
git clone https://github.com/alethi-co/loc.git
cd loc
go build -o loc ./cmd
```

## Usage

### Basic Usage

Count lines in the current directory:

```bash
loc
```

Count lines in a specific directory:

```bash
loc ./src
```

Count lines in a specific file:

```bash
loc main.go
```

### Output Formats

**Pretty output (default):**

```bash
loc
```

```
  LOC - Lines of Code Counter
  ────────────────────────────────────────────────────────────
  Total: 42 files | 3,847 lines | 2,918 code | 412 comments
```

**JSON output:**

```bash
loc -o json
```

```json
{
  "total": {
    "files": 42,
    "lines": 3847,
    "code": 2918,
    "blanks": 517,
    "comments": 412
  },
  "byLanguage": [...],
  "byDirectory": [...],
  "byPackage": [...]
}
```

**Raw output (for scripting):**

```bash
loc -o raw
```

```
Files: 42
Lines: 3847
Code: 2918
Blanks: 517
Comments: 412
```

### Breakdown Options

**By language:**

```bash
loc -l
```

```
  By Language
  ┌────────────┬───────┬─────────┬────────┬──────────┐
  │ Language   │ Files │   Lines │   Code │ Comments │
  ├────────────┼───────┼─────────┼────────┼──────────┤
  │ Go         │    35 │   2,941 │  2,234 │      312 │
  │ JavaScript │     5 │     678 │    512 │       78 │
  │ YAML       │     2 │     228 │    172 │       22 │
  └────────────┴───────┴─────────┴────────┴──────────┘
```

**By directory:**

```bash
loc -d
```

**By package:**

```bash
loc -p
```

### Filtering

**Include only specific patterns:**

```bash
loc -i "*.go"           # Only Go files
loc -i "*.go" -i "*.js" # Go and JavaScript files
```

**Exclude patterns:**

```bash
loc -e "vendor/"        # Exclude vendor directory
loc -e "*_test.go"      # Exclude test files
loc -e "*.min.js"       # Exclude minified files
```

**Ignore .gitignore:**

```bash
loc --no-gitignore
```

### Performance Tuning

**Set worker count:**

```bash
loc -w 8    # Use 8 parallel workers
```

By default, LOC uses the number of CPU cores.

### Code-Only Mode

Count only code lines (exclude blanks and comments from the total):

```bash
loc -c
```

## Flag Reference

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--by-language` | `-l` | Show breakdown by language | false |
| `--by-dir` | `-d` | Show breakdown by directory | false |
| `--by-package` | `-p` | Show breakdown by package | false |
| `--code-only` | `-c` | Only count code lines | false |
| `--output` | `-o` | Output format: pretty, json, raw | pretty |
| `--include` | `-i` | Glob patterns to include (can be repeated) | - |
| `--exclude` | `-e` | Glob patterns to exclude (can be repeated) | - |
| `--no-gitignore` | - | Don't respect .gitignore files | false |
| `--workers` | `-w` | Number of parallel workers | CPU count |
| `--version` | `-v` | Show version | - |
| `--help` | `-h` | Show help | - |

## Supported Languages

LOC supports 30+ programming languages including:

- Go, Rust, C, C++, Java, Kotlin, Swift, Objective-C
- Python, Ruby, Perl, PHP, JavaScript, TypeScript
- HTML, CSS, SCSS, SASS
- SQL, Shell/Bash
- YAML, JSON, TOML, XML
- Markdown, Dockerfile
- And more...

## Built-in Exclusions

LOC automatically excludes common non-source directories:

- `node_modules`, `vendor`, `.git`
- `__pycache__`, `.venv`, `venv`
- `dist`, `build`, `target`
- `coverage`, `.nyc_output`
- And more...

## Performance

LOC is designed for speed:

- Parallel file processing with configurable worker pools
- Buffer pooling for efficient memory usage
- Streaming file reading for large files
- Binary file detection to skip non-text files

Typical performance: 10,000+ files in under 2 seconds.

## License

MIT License
