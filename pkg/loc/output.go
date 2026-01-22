package loc

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// OutputFormat represents the output format type
type OutputFormat string

const (
	FormatPretty OutputFormat = "pretty"
	FormatJSON   OutputFormat = "json"
	FormatRaw    OutputFormat = "raw"
)

// OutputConfig holds configuration for output formatting
type OutputConfig struct {
	Format     OutputFormat
	ByLanguage bool
	ByDir      bool
	ByPackage  bool
}

// FormatOutput formats a Summary according to the given OutputConfig
func FormatOutput(summary *Summary, config *OutputConfig) string {
	switch config.Format {
	case FormatJSON:
		return formatJSON(summary)
	case FormatRaw:
		return formatRaw(summary, config)
	default:
		return formatPretty(summary, config)
	}
}

// formatNumber formats an integer with commas for thousands separators
func formatNumber(n int) string {
	// Handle negative numbers
	if n < 0 {
		return "-" + formatNumber(-n)
	}

	str := fmt.Sprintf("%d", n)
	if len(str) <= 3 {
		return str
	}

	// Insert commas from right to left
	var result strings.Builder
	for i, c := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result.WriteByte(',')
		}
		result.WriteRune(c)
	}
	return result.String()
}

// formatPretty formats the summary as pretty terminal output with box-drawing characters
func formatPretty(summary *Summary, config *OutputConfig) string {
	var sb strings.Builder

	// Header
	sb.WriteString("\n  LOC - Lines of Code Counter\n")
	sb.WriteString("  ")
	sb.WriteString(strings.Repeat("\u2500", 60)) // horizontal line
	sb.WriteString("\n")

	// Show breakdowns if requested
	if config.ByLanguage && len(summary.ByLanguage) > 0 {
		sb.WriteString("\n")
		sb.WriteString(formatLanguageTable(summary))
	}

	if config.ByDir && len(summary.ByDirectory) > 0 {
		sb.WriteString("\n")
		sb.WriteString(formatDirectoryTable(summary))
	}

	if config.ByPackage && len(summary.ByPackage) > 0 {
		sb.WriteString("\n")
		sb.WriteString(formatPackageTable(summary))
	}

	// If any breakdowns were shown, add a separator before the total
	if config.ByLanguage || config.ByDir || config.ByPackage {
		sb.WriteString("\n  ")
		sb.WriteString(strings.Repeat("\u2500", 60))
		sb.WriteString("\n")
	}

	// Total summary line
	sb.WriteString(fmt.Sprintf("  Total: %s files | %s lines | %s code | %s comments\n",
		formatNumber(summary.TotalFiles),
		formatNumber(summary.TotalLines),
		formatNumber(summary.TotalCode),
		formatNumber(summary.TotalComments)))

	return sb.String()
}

// formatLanguageTable creates a pretty table for language breakdown
func formatLanguageTable(summary *Summary) string {
	var sb strings.Builder

	// Get sorted language names
	languages := make([]string, 0, len(summary.ByLanguage))
	for lang := range summary.ByLanguage {
		languages = append(languages, lang)
	}
	sort.Strings(languages)

	// Calculate column widths
	langWidth := 10 // minimum width for "Language"
	filesWidth := 5
	linesWidth := 7
	codeWidth := 6
	commentsWidth := 8

	for _, lang := range languages {
		stats := summary.ByLanguage[lang]
		if len(lang) > langWidth {
			langWidth = len(lang)
		}
		if w := len(formatNumber(stats.Files)); w > filesWidth {
			filesWidth = w
		}
		if w := len(formatNumber(stats.Lines)); w > linesWidth {
			linesWidth = w
		}
		if w := len(formatNumber(stats.Code)); w > codeWidth {
			codeWidth = w
		}
		if w := len(formatNumber(stats.Comments)); w > commentsWidth {
			commentsWidth = w
		}
	}

	sb.WriteString("  By Language\n")

	// Top border
	sb.WriteString("  \u250c") // top-left corner
	sb.WriteString(strings.Repeat("\u2500", langWidth+2))
	sb.WriteString("\u252c") // top-tee
	sb.WriteString(strings.Repeat("\u2500", filesWidth+2))
	sb.WriteString("\u252c")
	sb.WriteString(strings.Repeat("\u2500", linesWidth+2))
	sb.WriteString("\u252c")
	sb.WriteString(strings.Repeat("\u2500", codeWidth+2))
	sb.WriteString("\u252c")
	sb.WriteString(strings.Repeat("\u2500", commentsWidth+2))
	sb.WriteString("\u2510\n") // top-right corner

	// Header row
	sb.WriteString(fmt.Sprintf("  \u2502 %-*s \u2502 %*s \u2502 %*s \u2502 %*s \u2502 %*s \u2502\n",
		langWidth, "Language",
		filesWidth, "Files",
		linesWidth, "Lines",
		codeWidth, "Code",
		commentsWidth, "Comments"))

	// Header separator
	sb.WriteString("  \u251c") // left-tee
	sb.WriteString(strings.Repeat("\u2500", langWidth+2))
	sb.WriteString("\u253c") // cross
	sb.WriteString(strings.Repeat("\u2500", filesWidth+2))
	sb.WriteString("\u253c")
	sb.WriteString(strings.Repeat("\u2500", linesWidth+2))
	sb.WriteString("\u253c")
	sb.WriteString(strings.Repeat("\u2500", codeWidth+2))
	sb.WriteString("\u253c")
	sb.WriteString(strings.Repeat("\u2500", commentsWidth+2))
	sb.WriteString("\u2524\n") // right-tee

	// Data rows
	for _, lang := range languages {
		stats := summary.ByLanguage[lang]
		sb.WriteString(fmt.Sprintf("  \u2502 %-*s \u2502 %*s \u2502 %*s \u2502 %*s \u2502 %*s \u2502\n",
			langWidth, lang,
			filesWidth, formatNumber(stats.Files),
			linesWidth, formatNumber(stats.Lines),
			codeWidth, formatNumber(stats.Code),
			commentsWidth, formatNumber(stats.Comments)))
	}

	// Bottom border
	sb.WriteString("  \u2514") // bottom-left corner
	sb.WriteString(strings.Repeat("\u2500", langWidth+2))
	sb.WriteString("\u2534") // bottom-tee
	sb.WriteString(strings.Repeat("\u2500", filesWidth+2))
	sb.WriteString("\u2534")
	sb.WriteString(strings.Repeat("\u2500", linesWidth+2))
	sb.WriteString("\u2534")
	sb.WriteString(strings.Repeat("\u2500", codeWidth+2))
	sb.WriteString("\u2534")
	sb.WriteString(strings.Repeat("\u2500", commentsWidth+2))
	sb.WriteString("\u2518\n") // bottom-right corner

	return sb.String()
}

// formatDirectoryTable creates a pretty table for directory breakdown
func formatDirectoryTable(summary *Summary) string {
	var sb strings.Builder

	// Get sorted directory paths
	dirs := make([]string, 0, len(summary.ByDirectory))
	for dir := range summary.ByDirectory {
		dirs = append(dirs, dir)
	}
	sort.Strings(dirs)

	// Calculate column widths
	dirWidth := 9 // minimum width for "Directory"
	filesWidth := 5
	linesWidth := 7

	for _, dir := range dirs {
		stats := summary.ByDirectory[dir]
		if len(dir) > dirWidth {
			dirWidth = len(dir)
		}
		if w := len(formatNumber(stats.Files)); w > filesWidth {
			filesWidth = w
		}
		if w := len(formatNumber(stats.Lines)); w > linesWidth {
			linesWidth = w
		}
	}

	sb.WriteString("  By Directory\n")

	// Top border
	sb.WriteString("  \u250c")
	sb.WriteString(strings.Repeat("\u2500", dirWidth+2))
	sb.WriteString("\u252c")
	sb.WriteString(strings.Repeat("\u2500", filesWidth+2))
	sb.WriteString("\u252c")
	sb.WriteString(strings.Repeat("\u2500", linesWidth+2))
	sb.WriteString("\u2510\n")

	// Header row
	sb.WriteString(fmt.Sprintf("  \u2502 %-*s \u2502 %*s \u2502 %*s \u2502\n",
		dirWidth, "Directory",
		filesWidth, "Files",
		linesWidth, "Lines"))

	// Header separator
	sb.WriteString("  \u251c")
	sb.WriteString(strings.Repeat("\u2500", dirWidth+2))
	sb.WriteString("\u253c")
	sb.WriteString(strings.Repeat("\u2500", filesWidth+2))
	sb.WriteString("\u253c")
	sb.WriteString(strings.Repeat("\u2500", linesWidth+2))
	sb.WriteString("\u2524\n")

	// Data rows
	for _, dir := range dirs {
		stats := summary.ByDirectory[dir]
		sb.WriteString(fmt.Sprintf("  \u2502 %-*s \u2502 %*s \u2502 %*s \u2502\n",
			dirWidth, dir,
			filesWidth, formatNumber(stats.Files),
			linesWidth, formatNumber(stats.Lines)))
	}

	// Bottom border
	sb.WriteString("  \u2514")
	sb.WriteString(strings.Repeat("\u2500", dirWidth+2))
	sb.WriteString("\u2534")
	sb.WriteString(strings.Repeat("\u2500", filesWidth+2))
	sb.WriteString("\u2534")
	sb.WriteString(strings.Repeat("\u2500", linesWidth+2))
	sb.WriteString("\u2518\n")

	return sb.String()
}

// formatPackageTable creates a pretty table for package breakdown
func formatPackageTable(summary *Summary) string {
	var sb strings.Builder

	// Get sorted package names
	packages := make([]string, 0, len(summary.ByPackage))
	for pkg := range summary.ByPackage {
		packages = append(packages, pkg)
	}
	sort.Strings(packages)

	// Calculate column widths
	pkgWidth := 7 // minimum width for "Package"
	filesWidth := 5
	linesWidth := 7
	codeWidth := 6
	commentsWidth := 8

	for _, pkg := range packages {
		stats := summary.ByPackage[pkg]
		if len(pkg) > pkgWidth {
			pkgWidth = len(pkg)
		}
		if w := len(formatNumber(stats.Files)); w > filesWidth {
			filesWidth = w
		}
		if w := len(formatNumber(stats.Lines)); w > linesWidth {
			linesWidth = w
		}
		if w := len(formatNumber(stats.Code)); w > codeWidth {
			codeWidth = w
		}
		if w := len(formatNumber(stats.Comments)); w > commentsWidth {
			commentsWidth = w
		}
	}

	sb.WriteString("  By Package\n")

	// Top border
	sb.WriteString("  \u250c")
	sb.WriteString(strings.Repeat("\u2500", pkgWidth+2))
	sb.WriteString("\u252c")
	sb.WriteString(strings.Repeat("\u2500", filesWidth+2))
	sb.WriteString("\u252c")
	sb.WriteString(strings.Repeat("\u2500", linesWidth+2))
	sb.WriteString("\u252c")
	sb.WriteString(strings.Repeat("\u2500", codeWidth+2))
	sb.WriteString("\u252c")
	sb.WriteString(strings.Repeat("\u2500", commentsWidth+2))
	sb.WriteString("\u2510\n")

	// Header row
	sb.WriteString(fmt.Sprintf("  \u2502 %-*s \u2502 %*s \u2502 %*s \u2502 %*s \u2502 %*s \u2502\n",
		pkgWidth, "Package",
		filesWidth, "Files",
		linesWidth, "Lines",
		codeWidth, "Code",
		commentsWidth, "Comments"))

	// Header separator
	sb.WriteString("  \u251c")
	sb.WriteString(strings.Repeat("\u2500", pkgWidth+2))
	sb.WriteString("\u253c")
	sb.WriteString(strings.Repeat("\u2500", filesWidth+2))
	sb.WriteString("\u253c")
	sb.WriteString(strings.Repeat("\u2500", linesWidth+2))
	sb.WriteString("\u253c")
	sb.WriteString(strings.Repeat("\u2500", codeWidth+2))
	sb.WriteString("\u253c")
	sb.WriteString(strings.Repeat("\u2500", commentsWidth+2))
	sb.WriteString("\u2524\n")

	// Data rows
	for _, pkg := range packages {
		stats := summary.ByPackage[pkg]
		sb.WriteString(fmt.Sprintf("  \u2502 %-*s \u2502 %*s \u2502 %*s \u2502 %*s \u2502 %*s \u2502\n",
			pkgWidth, pkg,
			filesWidth, formatNumber(stats.Files),
			linesWidth, formatNumber(stats.Lines),
			codeWidth, formatNumber(stats.Code),
			commentsWidth, formatNumber(stats.Comments)))
	}

	// Bottom border
	sb.WriteString("  \u2514")
	sb.WriteString(strings.Repeat("\u2500", pkgWidth+2))
	sb.WriteString("\u2534")
	sb.WriteString(strings.Repeat("\u2500", filesWidth+2))
	sb.WriteString("\u2534")
	sb.WriteString(strings.Repeat("\u2500", linesWidth+2))
	sb.WriteString("\u2534")
	sb.WriteString(strings.Repeat("\u2500", codeWidth+2))
	sb.WriteString("\u2534")
	sb.WriteString(strings.Repeat("\u2500", commentsWidth+2))
	sb.WriteString("\u2518\n")

	return sb.String()
}

// jsonOutput represents the JSON output structure
type jsonOutput struct {
	Total       jsonTotal            `json:"total"`
	ByLanguage  []jsonLanguageStats  `json:"byLanguage"`
	ByDirectory []jsonDirectoryStats `json:"byDirectory"`
	ByPackage   []jsonPackageStats   `json:"byPackage"`
}

type jsonTotal struct {
	Files    int `json:"files"`
	Lines    int `json:"lines"`
	Code     int `json:"code"`
	Blanks   int `json:"blanks"`
	Comments int `json:"comments"`
}

type jsonLanguageStats struct {
	Language string `json:"language"`
	Files    int    `json:"files"`
	Lines    int    `json:"lines"`
	Code     int    `json:"code"`
	Blanks   int    `json:"blanks"`
	Comments int    `json:"comments"`
}

type jsonDirectoryStats struct {
	Path     string `json:"path"`
	Files    int    `json:"files"`
	Lines    int    `json:"lines"`
	Code     int    `json:"code"`
	Blanks   int    `json:"blanks"`
	Comments int    `json:"comments"`
}

type jsonPackageStats struct {
	Package  string `json:"package"`
	Files    int    `json:"files"`
	Lines    int    `json:"lines"`
	Code     int    `json:"code"`
	Blanks   int    `json:"blanks"`
	Comments int    `json:"comments"`
}

// formatJSON formats the summary as JSON (always includes full data)
func formatJSON(summary *Summary) string {
	output := jsonOutput{
		Total: jsonTotal{
			Files:    summary.TotalFiles,
			Lines:    summary.TotalLines,
			Code:     summary.TotalCode,
			Blanks:   summary.TotalBlanks,
			Comments: summary.TotalComments,
		},
		ByLanguage:  make([]jsonLanguageStats, 0, len(summary.ByLanguage)),
		ByDirectory: make([]jsonDirectoryStats, 0, len(summary.ByDirectory)),
		ByPackage:   make([]jsonPackageStats, 0, len(summary.ByPackage)),
	}

	// Sort and add language stats
	languages := make([]string, 0, len(summary.ByLanguage))
	for lang := range summary.ByLanguage {
		languages = append(languages, lang)
	}
	sort.Strings(languages)
	for _, lang := range languages {
		stats := summary.ByLanguage[lang]
		output.ByLanguage = append(output.ByLanguage, jsonLanguageStats{
			Language: stats.Language,
			Files:    stats.Files,
			Lines:    stats.Lines,
			Code:     stats.Code,
			Blanks:   stats.Blanks,
			Comments: stats.Comments,
		})
	}

	// Sort and add directory stats
	dirs := make([]string, 0, len(summary.ByDirectory))
	for dir := range summary.ByDirectory {
		dirs = append(dirs, dir)
	}
	sort.Strings(dirs)
	for _, dir := range dirs {
		stats := summary.ByDirectory[dir]
		output.ByDirectory = append(output.ByDirectory, jsonDirectoryStats{
			Path:     stats.Path,
			Files:    stats.Files,
			Lines:    stats.Lines,
			Code:     stats.Code,
			Blanks:   stats.Blanks,
			Comments: stats.Comments,
		})
	}

	// Sort and add package stats
	packages := make([]string, 0, len(summary.ByPackage))
	for pkg := range summary.ByPackage {
		packages = append(packages, pkg)
	}
	sort.Strings(packages)
	for _, pkg := range packages {
		stats := summary.ByPackage[pkg]
		output.ByPackage = append(output.ByPackage, jsonPackageStats{
			Package:  stats.Package,
			Files:    stats.Files,
			Lines:    stats.Lines,
			Code:     stats.Code,
			Blanks:   stats.Blanks,
			Comments: stats.Comments,
		})
	}

	data, _ := json.MarshalIndent(output, "", "  ")
	return string(data) + "\n"
}

// formatRaw formats the summary as labeled, one stat per line
func formatRaw(summary *Summary, config *OutputConfig) string {
	var sb strings.Builder

	// Basic summary (always shown)
	sb.WriteString(fmt.Sprintf("Files: %d\n", summary.TotalFiles))
	sb.WriteString(fmt.Sprintf("Lines: %d\n", summary.TotalLines))
	sb.WriteString(fmt.Sprintf("Code: %d\n", summary.TotalCode))
	sb.WriteString(fmt.Sprintf("Blanks: %d\n", summary.TotalBlanks))
	sb.WriteString(fmt.Sprintf("Comments: %d\n", summary.TotalComments))

	// Add language breakdown if requested
	if config.ByLanguage && len(summary.ByLanguage) > 0 {
		sb.WriteString("\n")
		languages := make([]string, 0, len(summary.ByLanguage))
		for lang := range summary.ByLanguage {
			languages = append(languages, lang)
		}
		sort.Strings(languages)
		for _, lang := range languages {
			stats := summary.ByLanguage[lang]
			sb.WriteString(fmt.Sprintf("%s: %d files, %d lines, %d code\n",
				lang, stats.Files, stats.Lines, stats.Code))
		}
	}

	// Add directory breakdown if requested
	if config.ByDir && len(summary.ByDirectory) > 0 {
		sb.WriteString("\n")
		dirs := make([]string, 0, len(summary.ByDirectory))
		for dir := range summary.ByDirectory {
			dirs = append(dirs, dir)
		}
		sort.Strings(dirs)
		for _, dir := range dirs {
			stats := summary.ByDirectory[dir]
			sb.WriteString(fmt.Sprintf("%s: %d files, %d lines\n",
				dir, stats.Files, stats.Lines))
		}
	}

	// Add package breakdown if requested
	if config.ByPackage && len(summary.ByPackage) > 0 {
		sb.WriteString("\n")
		packages := make([]string, 0, len(summary.ByPackage))
		for pkg := range summary.ByPackage {
			packages = append(packages, pkg)
		}
		sort.Strings(packages)
		for _, pkg := range packages {
			stats := summary.ByPackage[pkg]
			sb.WriteString(fmt.Sprintf("%s: %d files, %d lines, %d code\n",
				pkg, stats.Files, stats.Lines, stats.Code))
		}
	}

	return sb.String()
}
