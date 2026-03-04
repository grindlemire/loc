package loc

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
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
	NoColor    bool
	Combined   bool // When true, don't show src/test/other breakdown
	ShowAll    bool // When true, include "other" files in the breakdown
	CodeOnly   bool // When true, suppress Blanks and Comments columns/fields
}

// FormatOutput formats a Summary according to the given OutputConfig
func FormatOutput(summary *Summary, config *OutputConfig) string {
	switch config.Format {
	case FormatJSON:
		return formatJSON(summary, config)
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

// hasOtherFiles returns true if the summary has any "other" files (config, docs, etc.)
func hasOtherFiles(summary *Summary) bool {
	return summary.OtherFiles > 0
}

// subRowData holds the data for a single sub-row in the breakdown
type subRowData struct {
	label    string
	files    int
	code     int
	comments int
}

// renderSubRows generates sub-rows with proper elbow prefixes for src/test/other breakdown.
// It handles 2-item (src/test) vs 3-item (src/test/other) cases.
// Returns a slice of row data (label with elbow prefix, files, code, comments).
func renderSubRows(summary *Summary, showOther bool, noColor bool) []subRowData {
	var rows []subRowData

	// Determine elbow characters based on color mode
	elbowMiddle := ElbowMiddle
	elbowLast := ElbowLast
	if noColor {
		elbowMiddle = ElbowMiddleASCII
		elbowLast = ElbowLastASCII
	}

	// Always have src and test
	// If we're showing other and there are other files, use middle elbow for src and test
	// Otherwise, use middle elbow for src and last elbow for test
	hasOther := showOther && hasOtherFiles(summary)

	// Source row
	srcPrefix := elbowMiddle
	if !hasOther && summary.TestFiles == 0 {
		// Only src files exist
		srcPrefix = elbowLast
	}
	rows = append(rows, subRowData{
		label:    srcPrefix + "src",
		files:    summary.SrcFiles,
		code:     summary.SrcCode,
		comments: summary.SrcComments,
	})

	// Test row (only if there are test files or if we want to always show it)
	testPrefix := elbowMiddle
	if !hasOther {
		testPrefix = elbowLast
	}
	rows = append(rows, subRowData{
		label:    testPrefix + "test",
		files:    summary.TestFiles,
		code:     summary.TestCode,
		comments: summary.TestComments,
	})

	// Other row (only when --all flag is set and there are other files)
	if hasOther {
		rows = append(rows, subRowData{
			label:    elbowLast + "other",
			files:    summary.OtherFiles,
			code:     summary.OtherCode,
			comments: summary.OtherComments,
		})
	}

	return rows
}

// formatPretty formats the summary as pretty terminal output with styled tables
func formatPretty(summary *Summary, config *OutputConfig) string {
	// Check if any breakdown table flags are set (language, dir, package)
	hasBreakdownTables := config.ByLanguage || config.ByDir || config.ByPackage

	// If no breakdown table flags, show the summary table
	// (which handles both combined and non-combined views with elbow sub-rows)
	if !hasBreakdownTables {
		return formatSummaryTable(summary, config)
	}

	// Otherwise, show breakdown tables
	var sb strings.Builder

	// Show summary table first (with src/test/other breakdown if not combined)
	sb.WriteString(formatSummaryTable(summary, config))

	// Show breakdowns if requested
	if config.ByLanguage && len(summary.ByLanguage) > 0 {
		if sb.Len() > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(formatLanguageTable(summary, config))
	}

	if config.ByDir && len(summary.ByDirectory) > 0 {
		if sb.Len() > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(formatDirectoryTable(summary, config))
	}

	if config.ByPackage && len(summary.ByPackage) > 0 {
		if sb.Len() > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(formatPackageTable(summary, config))
	}

	return sb.String()
}

// formatSummaryTable renders the default summary table using Charmbracelet lipgloss/table
func formatSummaryTable(summary *Summary, config *OutputConfig) string {
	// Get appropriate styles based on NoColor setting
	var styles *Styles
	if config.NoColor {
		styles = NoColorStyles()
	} else {
		styles = DefaultStyles()
	}

	// When --combined is set, show simple table without breakdown
	if config.Combined {
		headers := []string{"Files", "Code"}
		row := []string{
			formatNumber(summary.TotalFiles),
			formatNumber(summary.TotalCode),
		}
		if !config.CodeOnly {
			headers = append(headers, "Comments")
			row = append(row, formatNumber(summary.TotalComments))
		}
		t := table.New().
			Border(lipgloss.RoundedBorder()).
			BorderStyle(styles.Border).
			Headers(headers...).
			Row(row...)

		// Style the header row with accent color
		t.StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return styles.Label.Padding(0, 1).Align(lipgloss.Center)
			}
			// Data rows: right-align numbers
			return styles.Number.Padding(0, 1).Align(lipgloss.Right)
		})

		return t.Render() + "\n"
	}

	// Default view: show Total row with elbow sub-rows for src/test/other
	// Build the rows
	var rows [][]string

	// Total row
	totalRow := []string{
		"Total",
		formatNumber(summary.TotalFiles),
		formatNumber(summary.TotalCode),
	}
	if !config.CodeOnly {
		totalRow = append(totalRow, formatNumber(summary.TotalComments))
	}
	rows = append(rows, totalRow)

	// Add sub-rows with elbow prefixes
	subRows := renderSubRows(summary, config.ShowAll, config.NoColor)
	for _, sr := range subRows {
		row := []string{
			sr.label,
			formatNumber(sr.files),
			formatNumber(sr.code),
		}
		if !config.CodeOnly {
			row = append(row, formatNumber(sr.comments))
		}
		rows = append(rows, row)
	}

	// Create the table with rounded borders
	headers := []string{"", "Files", "Code"}
	if !config.CodeOnly {
		headers = append(headers, "Comments")
	}
	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(styles.Border).
		Headers(headers...)

	// Add all rows
	for _, row := range rows {
		t.Row(row...)
	}

	// Style the table: header centered, first column left-aligned, rest right-aligned
	t.StyleFunc(func(row, col int) lipgloss.Style {
		if row == table.HeaderRow {
			return styles.Label.Padding(0, 1).Align(lipgloss.Center)
		}
		// First column (label) left-aligned
		if col == 0 {
			// Use TotalLabel style for the Total row (row 0)
			if row == 0 {
				return styles.TotalLabel.Padding(0, 1).Align(lipgloss.Left)
			}
			// Use SubRowLabel style for sub-rows
			return styles.SubRowLabel.Padding(0, 1).Align(lipgloss.Left)
		}
		// Numeric columns right-aligned
		if row == 0 {
			return styles.TotalNumber.Padding(0, 1).Align(lipgloss.Right)
		}
		return styles.Number.Padding(0, 1).Align(lipgloss.Right)
	})

	return t.Render() + "\n"
}

// breakdownRowType indicates the type of row in a breakdown table
type breakdownRowType int

const (
	breakdownRowMain  breakdownRowType = iota // Main language/dir/package row
	breakdownRowSub                           // Sub-row (src/test/other)
	breakdownRowTotal                         // Total row
)

// breakdownRow holds a row in the breakdown table with its type
type breakdownRow struct {
	rowType breakdownRowType
	data    []string
}

// breakdownEntry represents a single entry in a breakdown table (language, directory, or package)
type breakdownEntry struct {
	Label string
	Stats *BaseStats
}

// formatBreakdownTable is a generic formatter for language, directory, and package tables.
// It eliminates duplication across the three specific formatters.
func formatBreakdownTable(
	title string,
	headerLabel string,
	entries []breakdownEntry,
	totalCode int,
	config *OutputConfig,
) string {
	// Get appropriate styles based on NoColor setting
	var styles *Styles
	if config.NoColor {
		styles = NoColorStyles()
	} else {
		styles = DefaultStyles()
	}

	// Determine elbow characters based on color mode
	elbowMiddle := ElbowMiddle
	elbowLast := ElbowLast
	if config.NoColor {
		elbowMiddle = ElbowMiddleASCII
		elbowLast = ElbowLastASCII
	}

	// Build rows with type information
	var rows []breakdownRow
	var totalFiles, totalCodeSum, totalComments int

	// When combined, show simple rows without sub-breakdown
	showSubRows := !config.Combined

	for _, entry := range entries {
		stats := entry.Stats
		totalFiles += stats.Files
		totalCodeSum += stats.Code
		totalComments += stats.Comments

		// Calculate percentage for the main row
		var pct float64
		if totalCode > 0 {
			pct = float64(stats.Code) / float64(totalCode) * 100
		}

		// Add main row
		mainRow := []string{
			entry.Label,
			formatNumber(stats.Files),
			formatNumber(stats.Code),
		}
		if !config.CodeOnly {
			mainRow = append(mainRow, formatNumber(stats.Comments))
		}
		mainRow = append(mainRow, fmt.Sprintf("%.1f%%", pct))
		rows = append(rows, breakdownRow{
			rowType: breakdownRowMain,
			data:    mainRow,
		})

		// Add sub-rows if not combined
		if showSubRows {
			// Determine if we have "other" files for this entry
			hasOther := config.ShowAll && stats.OtherFiles > 0

			// Determine elbow prefixes
			srcPrefix := elbowMiddle
			testPrefix := elbowMiddle
			if !hasOther {
				testPrefix = elbowLast
			}

			// Source sub-row
			var srcPct float64
			if totalCode > 0 {
				srcPct = float64(stats.SrcCode) / float64(totalCode) * 100
			}
			srcRow := []string{
				srcPrefix + "src",
				formatNumber(stats.SrcFiles),
				formatNumber(stats.SrcCode),
			}
			if !config.CodeOnly {
				srcRow = append(srcRow, formatNumber(stats.SrcComments))
			}
			srcRow = append(srcRow, fmt.Sprintf("%.1f%%", srcPct))
			rows = append(rows, breakdownRow{
				rowType: breakdownRowSub,
				data:    srcRow,
			})

			// Test sub-row
			var testPct float64
			if totalCode > 0 {
				testPct = float64(stats.TestCode) / float64(totalCode) * 100
			}
			testRow := []string{
				testPrefix + "test",
				formatNumber(stats.TestFiles),
				formatNumber(stats.TestCode),
			}
			if !config.CodeOnly {
				testRow = append(testRow, formatNumber(stats.TestComments))
			}
			testRow = append(testRow, fmt.Sprintf("%.1f%%", testPct))
			rows = append(rows, breakdownRow{
				rowType: breakdownRowSub,
				data:    testRow,
			})

			// Other sub-row (only when --all flag is set and there are other files)
			if hasOther {
				var otherPct float64
				if totalCode > 0 {
					otherPct = float64(stats.OtherCode) / float64(totalCode) * 100
				}
				otherRow := []string{
					elbowLast + "other",
					formatNumber(stats.OtherFiles),
					formatNumber(stats.OtherCode),
				}
				if !config.CodeOnly {
					otherRow = append(otherRow, formatNumber(stats.OtherComments))
				}
				otherRow = append(otherRow, fmt.Sprintf("%.1f%%", otherPct))
				rows = append(rows, breakdownRow{
					rowType: breakdownRowSub,
					data:    otherRow,
				})
			}
		}
	}

	// Add totals row
	totalsRow := []string{
		"Total",
		formatNumber(totalFiles),
		formatNumber(totalCodeSum),
	}
	if !config.CodeOnly {
		totalsRow = append(totalsRow, formatNumber(totalComments))
	}
	totalsRow = append(totalsRow, "100.0%")
	rows = append(rows, breakdownRow{
		rowType: breakdownRowTotal,
		data:    totalsRow,
	})

	// Create the table with rounded borders
	bdHeaders := []string{headerLabel, "Files", "Code"}
	if !config.CodeOnly {
		bdHeaders = append(bdHeaders, "Comments")
	}
	bdHeaders = append(bdHeaders, "%")
	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(styles.Border).
		Headers(bdHeaders...)

	// Add all rows
	for _, row := range rows {
		t.Row(row.data...)
	}

	// Style the table: header centered, first column left-aligned, rest right-aligned
	t.StyleFunc(func(rowIdx, col int) lipgloss.Style {
		if rowIdx == table.HeaderRow {
			return styles.Label.Padding(0, 1).Align(lipgloss.Center)
		}
		// Get the row type
		rowType := rows[rowIdx].rowType

		// First column (label) left-aligned
		if col == 0 {
			switch rowType {
			case breakdownRowTotal:
				return styles.TotalLabel.Padding(0, 1).Align(lipgloss.Left)
			case breakdownRowSub:
				return styles.SubRowLabel.Padding(0, 1).Align(lipgloss.Left)
			default:
				return styles.Number.Padding(0, 1).Align(lipgloss.Left)
			}
		}
		// Numeric columns right-aligned
		switch rowType {
		case breakdownRowTotal:
			return styles.TotalNumber.Padding(0, 1).Align(lipgloss.Right)
		case breakdownRowSub:
			return styles.SubRowLabel.Padding(0, 1).Align(lipgloss.Right)
		default:
			return styles.Number.Padding(0, 1).Align(lipgloss.Right)
		}
	})

	return styles.SectionTitle.Render(title) + "\n" + t.Render() + "\n"
}

// formatLanguageTable creates a pretty table for language breakdown using Charmbracelet lipgloss/table
func formatLanguageTable(summary *Summary, config *OutputConfig) string {
	// Get sorted language names and build entries
	languages := make([]string, 0, len(summary.ByLanguage))
	for lang := range summary.ByLanguage {
		languages = append(languages, lang)
	}
	sort.Strings(languages)

	entries := make([]breakdownEntry, 0, len(languages))
	for _, lang := range languages {
		stats := summary.ByLanguage[lang]
		entries = append(entries, breakdownEntry{
			Label: lang,
			Stats: &stats.BaseStats,
		})
	}

	return formatBreakdownTable("By Language", "Language", entries, summary.TotalCode, config)
}

// formatDirectoryTable creates a pretty table for directory breakdown using Charmbracelet lipgloss/table
func formatDirectoryTable(summary *Summary, config *OutputConfig) string {
	// Get sorted directory paths and build entries
	dirs := make([]string, 0, len(summary.ByDirectory))
	for dir := range summary.ByDirectory {
		dirs = append(dirs, dir)
	}
	sort.Strings(dirs)

	entries := make([]breakdownEntry, 0, len(dirs))
	for _, dir := range dirs {
		stats := summary.ByDirectory[dir]
		entries = append(entries, breakdownEntry{
			Label: dir,
			Stats: &stats.BaseStats,
		})
	}

	return formatBreakdownTable("By Directory", "Directory", entries, summary.TotalCode, config)
}

// formatPackageTable creates a pretty table for package breakdown using Charmbracelet lipgloss/table
func formatPackageTable(summary *Summary, config *OutputConfig) string {
	// Get sorted package names and build entries
	packages := make([]string, 0, len(summary.ByPackage))
	for pkg := range summary.ByPackage {
		packages = append(packages, pkg)
	}
	sort.Strings(packages)

	entries := make([]breakdownEntry, 0, len(packages))
	for _, pkg := range packages {
		stats := summary.ByPackage[pkg]
		entries = append(entries, breakdownEntry{
			Label: pkg,
			Stats: &stats.BaseStats,
		})
	}

	return formatBreakdownTable("By Package", "Package", entries, summary.TotalCode, config)
}

// jsonOutput represents the JSON output structure
type jsonOutput struct {
	Total       jsonTotal            `json:"total"`
	Src         *jsonTestStats       `json:"src,omitempty"`
	Test        *jsonTestStats       `json:"test,omitempty"`
	Other       *jsonTestStats       `json:"other,omitempty"`
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

type jsonTestStats struct {
	Files    int `json:"files"`
	Lines    int `json:"lines"`
	Code     int `json:"code"`
	Blanks   int `json:"blanks"`
	Comments int `json:"comments"`
}

// jsonSubStats represents sub-stats for src/test/other in breakdown entries
type jsonSubStats struct {
	Files    int `json:"files"`
	Code     int `json:"code"`
	Comments int `json:"comments"`
}

type jsonLanguageStats struct {
	Language string        `json:"language"`
	Files    int           `json:"files"`
	Lines    int           `json:"lines"`
	Code     int           `json:"code"`
	Blanks   int           `json:"blanks"`
	Comments int           `json:"comments"`
	Src      *jsonSubStats `json:"src,omitempty"`
	Test     *jsonSubStats `json:"test,omitempty"`
	Other    *jsonSubStats `json:"other,omitempty"`
}

type jsonDirectoryStats struct {
	Path     string        `json:"path"`
	Files    int           `json:"files"`
	Lines    int           `json:"lines"`
	Code     int           `json:"code"`
	Blanks   int           `json:"blanks"`
	Comments int           `json:"comments"`
	Src      *jsonSubStats `json:"src,omitempty"`
	Test     *jsonSubStats `json:"test,omitempty"`
	Other    *jsonSubStats `json:"other,omitempty"`
}

type jsonPackageStats struct {
	Package  string        `json:"package"`
	Files    int           `json:"files"`
	Lines    int           `json:"lines"`
	Code     int           `json:"code"`
	Blanks   int           `json:"blanks"`
	Comments int           `json:"comments"`
	Src      *jsonSubStats `json:"src,omitempty"`
	Test     *jsonSubStats `json:"test,omitempty"`
	Other    *jsonSubStats `json:"other,omitempty"`
}

// formatJSON formats the summary as JSON
// When not combined: includes src/test/other breakdown in top-level and in each breakdown entry
// When combined: omits src/test/other fields
// When showAll: includes "other" stats
func formatJSON(summary *Summary, config *OutputConfig) string {
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

	// Add src/test/other stats when not combined
	if !config.Combined {
		output.Src = &jsonTestStats{
			Files:    summary.SrcFiles,
			Lines:    summary.SrcLines,
			Code:     summary.SrcCode,
			Blanks:   summary.SrcBlanks,
			Comments: summary.SrcComments,
		}
		output.Test = &jsonTestStats{
			Files:    summary.TestFiles,
			Lines:    summary.TestLines,
			Code:     summary.TestCode,
			Blanks:   summary.TestBlanks,
			Comments: summary.TestComments,
		}
		// Include "other" only when --all is set and there are other files
		if config.ShowAll && hasOtherFiles(summary) {
			output.Other = &jsonTestStats{
				Files:    summary.OtherFiles,
				Lines:    summary.OtherLines,
				Code:     summary.OtherCode,
				Blanks:   summary.OtherBlanks,
				Comments: summary.OtherComments,
			}
		}
	}

	// Sort and add language stats
	languages := make([]string, 0, len(summary.ByLanguage))
	for lang := range summary.ByLanguage {
		languages = append(languages, lang)
	}
	sort.Strings(languages)
	for _, lang := range languages {
		stats := summary.ByLanguage[lang]
		langStats := jsonLanguageStats{
			Language: stats.Language,
			Files:    stats.Files,
			Lines:    stats.Lines,
			Code:     stats.Code,
			Blanks:   stats.Blanks,
			Comments: stats.Comments,
		}
		// Add sub-stats when not combined
		if !config.Combined {
			langStats.Src = &jsonSubStats{
				Files:    stats.SrcFiles,
				Code:     stats.SrcCode,
				Comments: stats.SrcComments,
			}
			langStats.Test = &jsonSubStats{
				Files:    stats.TestFiles,
				Code:     stats.TestCode,
				Comments: stats.TestComments,
			}
			// Include "other" only when --all is set and there are other files for this language
			if config.ShowAll && stats.OtherFiles > 0 {
				langStats.Other = &jsonSubStats{
					Files:    stats.OtherFiles,
					Code:     stats.OtherCode,
					Comments: stats.OtherComments,
				}
			}
		}
		output.ByLanguage = append(output.ByLanguage, langStats)
	}

	// Sort and add directory stats
	dirs := make([]string, 0, len(summary.ByDirectory))
	for dir := range summary.ByDirectory {
		dirs = append(dirs, dir)
	}
	sort.Strings(dirs)
	for _, dir := range dirs {
		stats := summary.ByDirectory[dir]
		dirStats := jsonDirectoryStats{
			Path:     stats.Path,
			Files:    stats.Files,
			Lines:    stats.Lines,
			Code:     stats.Code,
			Blanks:   stats.Blanks,
			Comments: stats.Comments,
		}
		// Add sub-stats when not combined
		if !config.Combined {
			dirStats.Src = &jsonSubStats{
				Files:    stats.SrcFiles,
				Code:     stats.SrcCode,
				Comments: stats.SrcComments,
			}
			dirStats.Test = &jsonSubStats{
				Files:    stats.TestFiles,
				Code:     stats.TestCode,
				Comments: stats.TestComments,
			}
			// Include "other" only when --all is set and there are other files for this directory
			if config.ShowAll && stats.OtherFiles > 0 {
				dirStats.Other = &jsonSubStats{
					Files:    stats.OtherFiles,
					Code:     stats.OtherCode,
					Comments: stats.OtherComments,
				}
			}
		}
		output.ByDirectory = append(output.ByDirectory, dirStats)
	}

	// Sort and add package stats
	packages := make([]string, 0, len(summary.ByPackage))
	for pkg := range summary.ByPackage {
		packages = append(packages, pkg)
	}
	sort.Strings(packages)
	for _, pkg := range packages {
		stats := summary.ByPackage[pkg]
		pkgStats := jsonPackageStats{
			Package:  stats.Package,
			Files:    stats.Files,
			Lines:    stats.Lines,
			Code:     stats.Code,
			Blanks:   stats.Blanks,
			Comments: stats.Comments,
		}
		// Add sub-stats when not combined
		if !config.Combined {
			pkgStats.Src = &jsonSubStats{
				Files:    stats.SrcFiles,
				Code:     stats.SrcCode,
				Comments: stats.SrcComments,
			}
			pkgStats.Test = &jsonSubStats{
				Files:    stats.TestFiles,
				Code:     stats.TestCode,
				Comments: stats.TestComments,
			}
			// Include "other" only when --all is set and there are other files for this package
			if config.ShowAll && stats.OtherFiles > 0 {
				pkgStats.Other = &jsonSubStats{
					Files:    stats.OtherFiles,
					Code:     stats.OtherCode,
					Comments: stats.OtherComments,
				}
			}
		}
		output.ByPackage = append(output.ByPackage, pkgStats)
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": "failed to marshal output: %s"}`+"\n", err)
	}
	return string(data) + "\n"
}

// formatRaw formats the summary as labeled, one stat per line
// Default: shows src/test/other breakdown
// --combined: shows only totals
// --all: includes "other" stats when present
func formatRaw(summary *Summary, config *OutputConfig) string {
	var sb strings.Builder

	// Basic summary (always shown)
	sb.WriteString(fmt.Sprintf("Files: %d\n", summary.TotalFiles))
	sb.WriteString(fmt.Sprintf("Lines: %d\n", summary.TotalLines))
	sb.WriteString(fmt.Sprintf("Code: %d\n", summary.TotalCode))
	if !config.CodeOnly {
		sb.WriteString(fmt.Sprintf("Blanks: %d\n", summary.TotalBlanks))
		sb.WriteString(fmt.Sprintf("Comments: %d\n", summary.TotalComments))
	}

	// Add src/test/other breakdown if not combined
	if !config.Combined {
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf("Source Files: %d\n", summary.SrcFiles))
		sb.WriteString(fmt.Sprintf("Source Code: %d\n", summary.SrcCode))
		sb.WriteString(fmt.Sprintf("Test Files: %d\n", summary.TestFiles))
		sb.WriteString(fmt.Sprintf("Test Code: %d\n", summary.TestCode))
		// Include "other" only when --all is set and there are other files
		if config.ShowAll && hasOtherFiles(summary) {
			sb.WriteString(fmt.Sprintf("Other Files: %d\n", summary.OtherFiles))
			sb.WriteString(fmt.Sprintf("Other Code: %d\n", summary.OtherCode))
		}
	}

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
			// Add sub-breakdown when not combined
			if !config.Combined {
				sb.WriteString(fmt.Sprintf("  src: %d files, %d code\n",
					stats.SrcFiles, stats.SrcCode))
				sb.WriteString(fmt.Sprintf("  test: %d files, %d code\n",
					stats.TestFiles, stats.TestCode))
				if config.ShowAll && stats.OtherFiles > 0 {
					sb.WriteString(fmt.Sprintf("  other: %d files, %d code\n",
						stats.OtherFiles, stats.OtherCode))
				}
			}
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
			sb.WriteString(fmt.Sprintf("%s: %d files, %d lines, %d code\n",
				dir, stats.Files, stats.Lines, stats.Code))
			// Add sub-breakdown when not combined
			if !config.Combined {
				sb.WriteString(fmt.Sprintf("  src: %d files, %d code\n",
					stats.SrcFiles, stats.SrcCode))
				sb.WriteString(fmt.Sprintf("  test: %d files, %d code\n",
					stats.TestFiles, stats.TestCode))
				if config.ShowAll && stats.OtherFiles > 0 {
					sb.WriteString(fmt.Sprintf("  other: %d files, %d code\n",
						stats.OtherFiles, stats.OtherCode))
				}
			}
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
			// Add sub-breakdown when not combined
			if !config.Combined {
				sb.WriteString(fmt.Sprintf("  src: %d files, %d code\n",
					stats.SrcFiles, stats.SrcCode))
				sb.WriteString(fmt.Sprintf("  test: %d files, %d code\n",
					stats.TestFiles, stats.TestCode))
				if config.ShowAll && stats.OtherFiles > 0 {
					sb.WriteString(fmt.Sprintf("  other: %d files, %d code\n",
						stats.OtherFiles, stats.OtherCode))
				}
			}
		}
	}

	return sb.String()
}
