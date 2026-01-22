# LOC

A fast, parallel lines-of-code counter for 30+ languages.

```bash
go install github.com/alethi-co/loc/cmd/loc@latest
```

## Quick Start

```bash
loc              # Count current directory
loc ./src        # Count specific directory
loc -l           # Breakdown by language
loc -o json      # JSON output
```

## Output Example

```
╭─────────┬───────┬───────┬──────────╮
│         │ Files │ Code  │ Comments │
├─────────┼───────┼───────┼──────────┤
│ Total   │    14 │ 6,716 │      915 │
│ +- src  │     8 │ 2,651 │      462 │
│ `- test │     6 │ 4,065 │      453 │
╰─────────┴───────┴───────┴──────────╯
```

With `-l` (by language):

```
By Language
╭──────────┬───────┬───────┬──────────┬────────╮
│ Language │ Files │ Code  │ Comments │   %    │
├──────────┼───────┼───────┼──────────┼────────┤
│ Go       │    14 │ 6,716 │      915 │ 100.0% │
│ +- src   │     8 │ 2,651 │      462 │  39.5% │
│ `- test  │     6 │ 4,065 │      453 │  60.5% │
╰──────────┴───────┴───────┴──────────┴────────╯
```

## Options

| Flag | Description |
|------|-------------|
| `-l` | Breakdown by language |
| `-d` | Breakdown by directory |
| `-p` | Breakdown by package |
| `-c` | Count only code lines (no blanks/comments) |
| `-o pretty\|json\|raw` | Output format (default: pretty) |
| `-i "*.go"` | Include only matching files |
| `-e "vendor/"` | Exclude matching files |
| `-w 8` | Worker count (default: CPU cores) |
| `--no-gitignore` | Don't respect .gitignore |

Multiple `-i` and `-e` flags can be combined.

## Features

- **Parallel processing** with configurable workers
- **Language-aware** counting of code, comments, and blanks
- **Gitignore support** enabled by default
- **Auto-skips** binaries and common non-source dirs (`node_modules`, `vendor`, `.git`, etc.)

## Supported Languages

Go, Rust, C/C++, Java, Kotlin, Swift, Python, Ruby, PHP, JavaScript, TypeScript, HTML, CSS, SCSS, SQL, Shell, YAML, JSON, TOML, XML, Markdown, Dockerfile, and more.

## License

MIT
