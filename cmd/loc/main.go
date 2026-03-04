package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/alethi-co/loc/pkg/loc"
	"github.com/mattn/go-isatty"
	"github.com/urfave/cli/v3"
)

// Version is set at build time
var Version = "dev"

func main() {
	cmd := buildCommand()
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func buildCommand() *cli.Command {
	return &cli.Command{
		Name:    "loc",
		Usage:   "A fast, parallel lines-of-code counter for software projects",
		Version: Version,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "by-language",
				Aliases: []string{"l"},
				Usage:   "Show breakdown by language",
			},
			&cli.BoolFlag{
				Name:    "by-dir",
				Aliases: []string{"d"},
				Usage:   "Show breakdown by directory",
			},
			&cli.BoolFlag{
				Name:    "by-package",
				Aliases: []string{"p"},
				Usage:   "Show breakdown by package",
			},
			&cli.BoolFlag{
				Name:    "code-only",
				Aliases: []string{"c"},
				Usage:   "Only count code lines (exclude blanks/comments)",
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Value:   "pretty",
				Usage:   "Output format: pretty, json, raw",
			},
			&cli.StringSliceFlag{
				Name:    "include",
				Aliases: []string{"i"},
				Usage:   "Additional glob patterns to include",
			},
			&cli.StringSliceFlag{
				Name:    "exclude",
				Aliases: []string{"e"},
				Usage:   "Glob patterns to exclude",
			},
			&cli.BoolFlag{
				Name:  "no-gitignore",
				Usage: "Don't respect .gitignore files",
			},
			&cli.IntFlag{
				Name:    "workers",
				Aliases: []string{"w"},
				Value:   runtime.NumCPU(),
				Usage:   "Number of parallel workers",
			},
			&cli.BoolFlag{
				Name:  "no-color",
				Usage: "Disable colored output",
			},
			&cli.BoolFlag{
				Name:  "combined",
				Usage: "Combine src/test/other into single totals (disable breakdown)",
			},
			&cli.BoolFlag{
				Name:    "all",
				Aliases: []string{"a"},
				Usage:   "Include non-source files (config, markdown, etc.)",
			},
			&cli.BoolFlag{
				Name:  "no-tests",
				Usage: "Exclude test files from counting",
			},
			&cli.BoolFlag{
				Name:  "tests-only",
				Usage: "Only count test files",
			},
			&cli.StringSliceFlag{
				Name:  "lang",
				Usage: "Include only specified languages (comma-separated or repeated, e.g. --lang Go,Python)",
			},
			&cli.StringSliceFlag{
				Name:  "exclude-lang",
				Usage: "Exclude specified languages (comma-separated or repeated, e.g. --exclude-lang Go)",
			},
		},
		Action: run,
	}
}

func run(ctx context.Context, cmd *cli.Command) error {
	// Get path argument (default to current directory)
	path := "."
	if cmd.NArg() > 0 {
		path = cmd.Args().Get(0)
	}

	// Validate mutually exclusive flags
	if cmd.Bool("no-tests") && cmd.Bool("tests-only") {
		return fmt.Errorf("--no-tests and --tests-only are mutually exclusive")
	}

	// Build config from flags
	config := &loc.Config{
		Workers:      int(cmd.Int("workers")),
		OutputFormat: cmd.String("output"),
		ByLanguage:   cmd.Bool("by-language"),
		ByDirectory:  cmd.Bool("by-dir"),
		ByPackage:    cmd.Bool("by-package"),
		CodeOnly:     cmd.Bool("code-only"),
		NoGitignore:  cmd.Bool("no-gitignore"),
		Include:      cmd.StringSlice("include"),
		Exclude:      cmd.StringSlice("exclude"),
		Combined:     cmd.Bool("combined"),
		All:          cmd.Bool("all"),
		NoTests:      cmd.Bool("no-tests"),
		TestsOnly:    cmd.Bool("tests-only"),
		Languages:    parseCommaSeparated(cmd.StringSlice("lang")),
		ExcludeLangs: parseCommaSeparated(cmd.StringSlice("exclude-lang")),
	}

	// Validate output format
	switch config.OutputFormat {
	case "pretty", "json", "raw":
		// valid
	default:
		return fmt.Errorf("invalid output format: %s (must be pretty, json, or raw)", config.OutputFormat)
	}

	// Create scanner and scan for files
	scanner := loc.NewScanner(config)
	files, err := scanner.Scan(path)
	if err != nil {
		return fmt.Errorf("scanning failed: %w", err)
	}

	// Create counter and count lines
	counter := loc.NewCounter(config)
	summary, err := counter.Count(files)
	if err != nil {
		return fmt.Errorf("counting failed: %w", err)
	}

	// Convert directory paths to relative paths for cleaner output
	absPath, err := filepath.Abs(path)
	if err == nil {
		summary.MakePathsRelative(absPath)
	}

	// Determine if colors should be disabled:
	// 1. --no-color flag is set, OR
	// 2. NO_COLOR environment variable is set (any non-empty value), OR
	// 3. stdout is not a TTY
	noColor := cmd.Bool("no-color") ||
		os.Getenv("NO_COLOR") != "" ||
		!isatty.IsTerminal(os.Stdout.Fd())

	// Format and print output
	outputConfig := &loc.OutputConfig{
		Format:     loc.OutputFormat(config.OutputFormat),
		ByLanguage: config.ByLanguage,
		ByDir:      config.ByDirectory,
		ByPackage:  config.ByPackage,
		NoColor:    noColor,
		Combined:   config.Combined,
		ShowAll:    config.All,
		CodeOnly:   config.CodeOnly,
	}

	output := loc.FormatOutput(summary, outputConfig)

	// Write to the command's writer (allows for testing with custom writers)
	writer := cmd.Writer
	if writer == nil {
		writer = os.Stdout
	}
	fmt.Fprint(writer, output)

	return nil
}

// parseCommaSeparated takes a string slice (from repeated flags) and splits
// each element on commas, returning a flat list of trimmed, non-empty values.
// This allows both --lang Go --lang Python and --lang Go,Python.
func parseCommaSeparated(values []string) []string {
	var result []string
	for _, v := range values {
		for _, part := range strings.Split(v, ",") {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
	}
	return result
}
