package loc

import (
	"encoding/json"
	"strings"
	"testing"
)

// createTestSummary creates a sample Summary for testing
func createTestSummary() *Summary {
	return &Summary{
		TotalFiles:    90,
		TotalLines:    6673,
		TotalCode:     5511,
		TotalBlanks:   555,
		TotalComments: 607,
		// Source file stats (non-test, non-other)
		SrcFiles:    60,
		SrcLines:    4500,
		SrcCode:     3700,
		SrcBlanks:   400,
		SrcComments: 400,
		// Test file stats
		TestFiles:    30,
		TestLines:    2173,
		TestCode:     1811,
		TestBlanks:   155,
		TestComments: 207,
		// Other file stats (none in default test)
		OtherFiles:    0,
		OtherLines:    0,
		OtherCode:     0,
		OtherBlanks:   0,
		OtherComments: 0,
		ByLanguage: map[string]*LanguageStats{
			"Go": {
				Language: "Go",
				BaseStats: BaseStats{
					Files: 42, Lines: 3521, Code: 2890, Blanks: 319, Comments: 312,
					SrcFiles: 30, SrcLines: 2500, SrcCode: 2000, SrcBlanks: 250, SrcComments: 250,
					TestFiles: 12, TestLines: 1021, TestCode: 890, TestBlanks: 69, TestComments: 62,
				},
			},
			"TypeScript": {
				Language: "TypeScript",
				BaseStats: BaseStats{
					Files: 28, Lines: 2104, Code: 1756, Blanks: 150, Comments: 198,
					SrcFiles: 18, SrcLines: 1400, SrcCode: 1200, SrcBlanks: 100, SrcComments: 100,
					TestFiles: 10, TestLines: 704, TestCode: 556, TestBlanks: 50, TestComments: 98,
				},
			},
			"Python": {
				Language: "Python",
				BaseStats: BaseStats{
					Files: 15, Lines: 892, Code: 723, Blanks: 80, Comments: 89,
					SrcFiles: 10, SrcLines: 600, SrcCode: 500, SrcBlanks: 50, SrcComments: 50,
					TestFiles: 5, TestLines: 292, TestCode: 223, TestBlanks: 30, TestComments: 39,
				},
			},
		},
		ByDirectory: map[string]*DirectoryStats{
			"pkg/api": {
				Path: "pkg/api",
				BaseStats: BaseStats{
					Files: 18, Lines: 1892, Code: 1580, Blanks: 180, Comments: 132,
					SrcFiles: 12, SrcLines: 1200, SrcCode: 1000, SrcBlanks: 100, SrcComments: 100,
					TestFiles: 6, TestLines: 692, TestCode: 580, TestBlanks: 80, TestComments: 32,
				},
			},
			"pkg/core": {
				Path: "pkg/core",
				BaseStats: BaseStats{
					Files: 24, Lines: 1629, Code: 1350, Blanks: 150, Comments: 129,
					SrcFiles: 16, SrcLines: 1100, SrcCode: 900, SrcBlanks: 100, SrcComments: 100,
					TestFiles: 8, TestLines: 529, TestCode: 450, TestBlanks: 50, TestComments: 29,
				},
			},
			"cmd": {
				Path: "cmd",
				BaseStats: BaseStats{
					Files: 12, Lines: 945, Code: 800, Blanks: 90, Comments: 55,
					SrcFiles: 8, SrcLines: 600, SrcCode: 500, SrcBlanks: 50, SrcComments: 50,
					TestFiles: 4, TestLines: 345, TestCode: 300, TestBlanks: 40, TestComments: 5,
				},
			},
		},
		ByPackage: map[string]*PackageStats{
			"api": {
				Package: "api",
				BaseStats: BaseStats{
					Files: 18, Lines: 1892, Code: 1580, Blanks: 180, Comments: 132,
					SrcFiles: 12, SrcLines: 1200, SrcCode: 1000, SrcBlanks: 100, SrcComments: 100,
					TestFiles: 6, TestLines: 692, TestCode: 580, TestBlanks: 80, TestComments: 32,
				},
			},
			"core": {
				Package: "core",
				BaseStats: BaseStats{
					Files: 24, Lines: 1629, Code: 1350, Blanks: 150, Comments: 129,
					SrcFiles: 16, SrcLines: 1100, SrcCode: 900, SrcBlanks: 100, SrcComments: 100,
					TestFiles: 8, TestLines: 529, TestCode: 450, TestBlanks: 50, TestComments: 29,
				},
			},
		},
	}
}

// createTestSummaryWithOther creates a sample Summary that includes "other" files
func createTestSummaryWithOther() *Summary {
	summary := createTestSummary()
	// Add "other" files (config, docs)
	summary.OtherFiles = 5
	summary.OtherLines = 200
	summary.OtherCode = 150
	summary.OtherBlanks = 30
	summary.OtherComments = 20
	// Update totals
	summary.TotalFiles += 5
	summary.TotalLines += 200
	summary.TotalCode += 150
	summary.TotalBlanks += 30
	summary.TotalComments += 20
	return summary
}

func TestFormatNumber(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{12, "12"},
		{123, "123"},
		{1234, "1,234"},
		{12345, "12,345"},
		{123456, "123,456"},
		{1234567, "1,234,567"},
		{6673, "6,673"},
		{5511, "5,511"},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			result := formatNumber(tc.input)
			if result != tc.expected {
				t.Errorf("formatNumber(%d) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

func TestPrettyOutputSummaryOnly(t *testing.T) {
	summary := createTestSummary()
	config := &OutputConfig{
		Format:     FormatPretty,
		ByLanguage: false,
		ByDir:      false,
		ByPackage:  false,
	}

	output := FormatOutput(summary, config)

	// Check summary table headers are present (Files, Code, Comments)
	if !strings.Contains(output, "Files") {
		t.Error("expected 'Files' header in summary table")
	}
	if !strings.Contains(output, "Code") {
		t.Error("expected 'Code' header in summary table")
	}
	if !strings.Contains(output, "Comments") {
		t.Error("expected 'Comments' header in summary table")
	}

	// Check formatted numbers are present
	if !strings.Contains(output, "90") {
		t.Error("expected '90' (files count) in output")
	}
	if !strings.Contains(output, "5,511") {
		t.Error("expected '5,511' (code count) in output")
	}
	if !strings.Contains(output, "607") {
		t.Error("expected '607' (comments count) in output")
	}

	// Check that rounded border characters are present
	if !strings.Contains(output, "\u256d") && !strings.Contains(output, "\u2502") {
		t.Error("expected rounded border characters in output")
	}

	// Check that breakdown tables are NOT present
	if strings.Contains(output, "By Language") {
		t.Error("did not expect 'By Language' table when ByLanguage is false")
	}
	if strings.Contains(output, "By Directory") {
		t.Error("did not expect 'By Directory' table when ByDir is false")
	}
	if strings.Contains(output, "By Package") {
		t.Error("did not expect 'By Package' table when ByPackage is false")
	}
}

func TestPrettyOutputWithLanguageBreakdown(t *testing.T) {
	summary := createTestSummary()
	config := &OutputConfig{
		Format:     FormatPretty,
		ByLanguage: true,
		ByDir:      false,
		ByPackage:  false,
	}

	output := FormatOutput(summary, config)

	// Check that language table is present
	if !strings.Contains(output, "By Language") {
		t.Error("expected 'By Language' table")
	}
	if !strings.Contains(output, "Go") {
		t.Error("expected 'Go' in language table")
	}
	if !strings.Contains(output, "TypeScript") {
		t.Error("expected 'TypeScript' in language table")
	}
	if !strings.Contains(output, "Python") {
		t.Error("expected 'Python' in language table")
	}

	// Check that Unicode rounded border characters are present (converted to lipgloss/table)
	if !strings.Contains(output, "\u256d") && !strings.Contains(output, "\u2502") {
		t.Error("expected rounded border characters in output")
	}

	// Check that percentage column is present
	if !strings.Contains(output, "%") {
		t.Error("expected '%' column header in output")
	}

	// Check that totals row is present
	if !strings.Contains(output, "Total") {
		t.Error("expected 'Total' row in output")
	}
}

func TestPrettyOutputWithDirectoryBreakdown(t *testing.T) {
	summary := createTestSummary()
	config := &OutputConfig{
		Format:     FormatPretty,
		ByLanguage: false,
		ByDir:      true,
		ByPackage:  false,
	}

	output := FormatOutput(summary, config)

	// Check that directory table is present
	if !strings.Contains(output, "By Directory") {
		t.Error("expected 'By Directory' table")
	}
	if !strings.Contains(output, "pkg/api") {
		t.Error("expected 'pkg/api' in directory table")
	}
	if !strings.Contains(output, "pkg/core") {
		t.Error("expected 'pkg/core' in directory table")
	}
	if !strings.Contains(output, "cmd") {
		t.Error("expected 'cmd' in directory table")
	}
}

func TestPrettyOutputWithPackageBreakdown(t *testing.T) {
	summary := createTestSummary()
	config := &OutputConfig{
		Format:     FormatPretty,
		ByLanguage: false,
		ByDir:      false,
		ByPackage:  true,
	}

	output := FormatOutput(summary, config)

	// Check that package table is present
	if !strings.Contains(output, "By Package") {
		t.Error("expected 'By Package' table")
	}
	if !strings.Contains(output, "api") {
		t.Error("expected 'api' in package table")
	}
	if !strings.Contains(output, "core") {
		t.Error("expected 'core' in package table")
	}
}

func TestPrettyOutputWithAllBreakdowns(t *testing.T) {
	summary := createTestSummary()
	config := &OutputConfig{
		Format:     FormatPretty,
		ByLanguage: true,
		ByDir:      true,
		ByPackage:  true,
	}

	output := FormatOutput(summary, config)

	// Check all tables are present
	if !strings.Contains(output, "By Language") {
		t.Error("expected 'By Language' table")
	}
	if !strings.Contains(output, "By Directory") {
		t.Error("expected 'By Directory' table")
	}
	if !strings.Contains(output, "By Package") {
		t.Error("expected 'By Package' table")
	}

	// Note: The "Total:" summary line is removed in the new format.
	// When breakdown flags are set, the breakdown tables are shown.
	// The summary table is only shown when NO breakdown flags are set.
}

func TestJSONOutputStructure(t *testing.T) {
	summary := createTestSummary()
	config := &OutputConfig{
		Format:     FormatJSON,
		ByLanguage: false, // flags should be ignored for JSON
		ByDir:      false,
		ByPackage:  false,
	}

	output := FormatOutput(summary, config)

	// Verify it's valid JSON
	var result jsonOutput
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	// Verify total values
	if result.Total.Files != 90 {
		t.Errorf("expected total.files = 90, got %d", result.Total.Files)
	}
	if result.Total.Lines != 6673 {
		t.Errorf("expected total.lines = 6673, got %d", result.Total.Lines)
	}
	if result.Total.Code != 5511 {
		t.Errorf("expected total.code = 5511, got %d", result.Total.Code)
	}
	if result.Total.Blanks != 555 {
		t.Errorf("expected total.blanks = 555, got %d", result.Total.Blanks)
	}
	if result.Total.Comments != 607 {
		t.Errorf("expected total.comments = 607, got %d", result.Total.Comments)
	}

	// Verify byLanguage is present (should always be included regardless of flags)
	if len(result.ByLanguage) != 3 {
		t.Errorf("expected 3 languages, got %d", len(result.ByLanguage))
	}

	// Verify byDirectory is present
	if len(result.ByDirectory) != 3 {
		t.Errorf("expected 3 directories, got %d", len(result.ByDirectory))
	}

	// Verify byPackage is present
	if len(result.ByPackage) != 2 {
		t.Errorf("expected 2 packages, got %d", len(result.ByPackage))
	}
}


func TestJSONOutputSorted(t *testing.T) {
	summary := createTestSummary()
	config := &OutputConfig{Format: FormatJSON}

	output := FormatOutput(summary, config)

	var result jsonOutput
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// Languages should be sorted alphabetically
	if len(result.ByLanguage) >= 3 {
		if result.ByLanguage[0].Language != "Go" {
			t.Errorf("expected first language to be Go, got %s", result.ByLanguage[0].Language)
		}
		if result.ByLanguage[1].Language != "Python" {
			t.Errorf("expected second language to be Python, got %s", result.ByLanguage[1].Language)
		}
		if result.ByLanguage[2].Language != "TypeScript" {
			t.Errorf("expected third language to be TypeScript, got %s", result.ByLanguage[2].Language)
		}
	}

	// Directories should be sorted alphabetically
	if len(result.ByDirectory) >= 3 {
		if result.ByDirectory[0].Path != "cmd" {
			t.Errorf("expected first directory to be cmd, got %s", result.ByDirectory[0].Path)
		}
	}

	// Packages should be sorted alphabetically
	if len(result.ByPackage) >= 2 {
		if result.ByPackage[0].Package != "api" {
			t.Errorf("expected first package to be api, got %s", result.ByPackage[0].Package)
		}
		if result.ByPackage[1].Package != "core" {
			t.Errorf("expected second package to be core, got %s", result.ByPackage[1].Package)
		}
	}
}

func TestRawOutputFormat(t *testing.T) {
	summary := createTestSummary()
	config := &OutputConfig{
		Format:     FormatRaw,
		ByLanguage: false,
		ByDir:      false,
		ByPackage:  false,
	}

	output := FormatOutput(summary, config)

	// Check each line exists (raw format uses unformatted numbers)
	expectedLines := []string{
		"Files: 90",
		"Lines: 6673",
		"Code: 5511",
		"Blanks: 555",
		"Comments: 607",
	}

	for _, expected := range expectedLines {
		if !strings.Contains(output, expected) {
			t.Errorf("expected %q in raw output", expected)
		}
	}

	// Check that breakdowns are NOT present
	if strings.Contains(output, "Go:") {
		t.Error("did not expect language breakdown when ByLanguage is false")
	}
}

func TestRawOutputWithLanguageBreakdown(t *testing.T) {
	summary := createTestSummary()
	config := &OutputConfig{
		Format:     FormatRaw,
		ByLanguage: true,
		ByDir:      false,
		ByPackage:  false,
	}

	output := FormatOutput(summary, config)

	// Check basic summary is present
	if !strings.Contains(output, "Files: 90") {
		t.Error("expected 'Files: 90' in output")
	}

	// Check language breakdown is present
	if !strings.Contains(output, "Go: 42 files, 3521 lines, 2890 code") {
		t.Error("expected Go breakdown in output")
	}
	if !strings.Contains(output, "TypeScript: 28 files, 2104 lines, 1756 code") {
		t.Error("expected TypeScript breakdown in output")
	}
	if !strings.Contains(output, "Python: 15 files, 892 lines, 723 code") {
		t.Error("expected Python breakdown in output")
	}
}

func TestRawOutputWithDirectoryBreakdown(t *testing.T) {
	summary := createTestSummary()
	config := &OutputConfig{
		Format:     FormatRaw,
		ByLanguage: false,
		ByDir:      true,
		ByPackage:  false,
	}

	output := FormatOutput(summary, config)

	// Check directory breakdown is present
	if !strings.Contains(output, "pkg/api: 18 files, 1892 lines") {
		t.Error("expected pkg/api breakdown in output")
	}
	if !strings.Contains(output, "pkg/core: 24 files, 1629 lines") {
		t.Error("expected pkg/core breakdown in output")
	}
	if !strings.Contains(output, "cmd: 12 files, 945 lines") {
		t.Error("expected cmd breakdown in output")
	}
}

func TestRawOutputWithPackageBreakdown(t *testing.T) {
	summary := createTestSummary()
	config := &OutputConfig{
		Format:     FormatRaw,
		ByLanguage: false,
		ByDir:      false,
		ByPackage:  true,
	}

	output := FormatOutput(summary, config)

	// Check package breakdown is present
	if !strings.Contains(output, "api: 18 files, 1892 lines, 1580 code") {
		t.Error("expected api package breakdown in output")
	}
	if !strings.Contains(output, "core: 24 files, 1629 lines, 1350 code") {
		t.Error("expected core package breakdown in output")
	}
}

func TestBreakdownFlagsAffectPrettyAndRaw(t *testing.T) {
	summary := createTestSummary()

	// With no flags set
	configNoFlags := &OutputConfig{
		Format:     FormatPretty,
		ByLanguage: false,
		ByDir:      false,
		ByPackage:  false,
	}

	// With all flags set
	configAllFlags := &OutputConfig{
		Format:     FormatPretty,
		ByLanguage: true,
		ByDir:      true,
		ByPackage:  true,
	}

	prettyNoFlags := FormatOutput(summary, configNoFlags)
	prettyAllFlags := FormatOutput(summary, configAllFlags)

	// Pretty output should differ
	if prettyNoFlags == prettyAllFlags {
		t.Error("expected pretty output to differ based on breakdown flags")
	}

	// Raw output should differ
	configNoFlags.Format = FormatRaw
	configAllFlags.Format = FormatRaw

	rawNoFlags := FormatOutput(summary, configNoFlags)
	rawAllFlags := FormatOutput(summary, configAllFlags)

	if rawNoFlags == rawAllFlags {
		t.Error("expected raw output to differ based on breakdown flags")
	}
}

// TestJSONOutputIncludesSrcTestRegardlessOfBreakdownFlags tests that JSON output
// always includes src/test breakdown (unless --combined is set), regardless of
// --by-language, --by-dir, --by-package flags
func TestJSONOutputIncludesSrcTestRegardlessOfBreakdownFlags(t *testing.T) {
	summary := createTestSummary()

	testCases := []struct {
		name       string
		byLanguage bool
		byDir      bool
		byPackage  bool
	}{
		{"no flags", false, false, false},
		{"only language", true, false, false},
		{"only dir", false, true, false},
		{"only package", false, false, true},
		{"all flags", true, true, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &OutputConfig{
				Format:     FormatJSON,
				ByLanguage: tc.byLanguage,
				ByDir:      tc.byDir,
				ByPackage:  tc.byPackage,
				Combined:   false, // Default: show src/test
			}

			output := FormatOutput(summary, config)

			var result jsonOutput
			if err := json.Unmarshal([]byte(output), &result); err != nil {
				t.Fatalf("invalid JSON: %v", err)
			}

			// src/test should always be present when not combined
			if result.Src == nil {
				t.Error("expected 'src' in JSON output regardless of breakdown flags")
			}
			if result.Test == nil {
				t.Error("expected 'test' in JSON output regardless of breakdown flags")
			}

			// All data should be present regardless of flags
			if len(result.ByLanguage) != 3 {
				t.Errorf("expected 3 languages regardless of flags, got %d", len(result.ByLanguage))
			}
			if len(result.ByDirectory) != 3 {
				t.Errorf("expected 3 directories regardless of flags, got %d", len(result.ByDirectory))
			}
			if len(result.ByPackage) != 2 {
				t.Errorf("expected 2 packages regardless of flags, got %d", len(result.ByPackage))
			}
		})
	}
}

func TestEmptySummary(t *testing.T) {
	summary := &Summary{
		TotalFiles:    0,
		TotalLines:    0,
		TotalCode:     0,
		TotalBlanks:   0,
		TotalComments: 0,
		ByLanguage:    make(map[string]*LanguageStats),
		ByDirectory:   make(map[string]*DirectoryStats),
		ByPackage:     make(map[string]*PackageStats),
	}

	t.Run("pretty", func(t *testing.T) {
		// With no breakdown flags, show summary table
		config := &OutputConfig{Format: FormatPretty}
		output := FormatOutput(summary, config)
		// Should have the table headers
		if !strings.Contains(output, "Files") {
			t.Error("expected 'Files' header in output")
		}
		// Should have zeros in the table
		if !strings.Contains(output, "0") {
			t.Error("expected '0' in output for empty summary")
		}
		// Should not crash with empty maps
	})

	t.Run("json", func(t *testing.T) {
		config := &OutputConfig{Format: FormatJSON}
		output := FormatOutput(summary, config)
		var result jsonOutput
		if err := json.Unmarshal([]byte(output), &result); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}
		if result.Total.Files != 0 {
			t.Errorf("expected 0 files, got %d", result.Total.Files)
		}
		if len(result.ByLanguage) != 0 {
			t.Errorf("expected empty byLanguage, got %d items", len(result.ByLanguage))
		}
	})

	t.Run("raw", func(t *testing.T) {
		config := &OutputConfig{Format: FormatRaw}
		output := FormatOutput(summary, config)
		if !strings.Contains(output, "Files: 0") {
			t.Error("expected 'Files: 0' in output")
		}
	})
}

func TestOutputFormatDefault(t *testing.T) {
	summary := createTestSummary()
	config := &OutputConfig{
		Format: "", // empty/default should use pretty
	}

	output := FormatOutput(summary, config)

	// Should produce pretty output (with summary table)
	// Check for table headers
	if !strings.Contains(output, "Files") || !strings.Contains(output, "Code") || !strings.Contains(output, "Comments") {
		t.Error("expected pretty output with summary table for empty/default format")
	}
}

func TestPrettyOutputHorizontalLine(t *testing.T) {
	summary := createTestSummary()
	config := &OutputConfig{Format: FormatPretty}

	output := FormatOutput(summary, config)

	// Check that Unicode table border characters are present (rounded borders)
	// The rounded border uses characters like: horizontal line
	if !strings.Contains(output, "\u2500") && !strings.Contains(output, "\u2502") {
		t.Error("expected Unicode border characters in output")
	}
}

func TestLargeNumbers(t *testing.T) {
	summary := &Summary{
		TotalFiles:    12345,
		TotalLines:    1234567,
		TotalCode:     987654,
		TotalBlanks:   123456,
		TotalComments: 123457,
		ByLanguage:    make(map[string]*LanguageStats),
		ByDirectory:   make(map[string]*DirectoryStats),
		ByPackage:     make(map[string]*PackageStats),
	}

	config := &OutputConfig{Format: FormatPretty}
	output := FormatOutput(summary, config)

	// Check formatted numbers (now in table format)
	if !strings.Contains(output, "12,345") {
		t.Error("expected '12,345' (files) in output")
	}
	if !strings.Contains(output, "987,654") {
		t.Error("expected '987,654' (code) in output")
	}
	if !strings.Contains(output, "123,457") {
		t.Error("expected '123,457' (comments) in output")
	}
}

func TestNoColorOutput(t *testing.T) {
	summary := createTestSummary()
	config := &OutputConfig{
		Format:  FormatPretty,
		NoColor: true,
	}

	output := FormatOutput(summary, config)

	// Should still have table structure with rounded borders
	if !strings.Contains(output, "Files") {
		t.Error("expected 'Files' header in no-color output")
	}
	if !strings.Contains(output, "Code") {
		t.Error("expected 'Code' header in no-color output")
	}
	if !strings.Contains(output, "Comments") {
		t.Error("expected 'Comments' header in no-color output")
	}

	// Check that formatted numbers are present
	if !strings.Contains(output, "90") {
		t.Error("expected '90' (files count) in no-color output")
	}
	if !strings.Contains(output, "5,511") {
		t.Error("expected '5,511' (code count) in no-color output")
	}

	// Rounded Unicode borders should still be present
	// The rounded border uses characters like: corners and lines
	if !strings.Contains(output, "\u2502") && !strings.Contains(output, "\u2500") {
		t.Error("expected Unicode border characters in no-color output")
	}
}

func TestNoColorOutputWithBreakdowns(t *testing.T) {
	summary := createTestSummary()
	config := &OutputConfig{
		Format:     FormatPretty,
		NoColor:    true,
		ByLanguage: true,
		ByDir:      true,
		ByPackage:  true,
	}

	output := FormatOutput(summary, config)

	// Should have all breakdown sections
	if !strings.Contains(output, "By Language") {
		t.Error("expected 'By Language' section in no-color output")
	}
	if !strings.Contains(output, "By Directory") {
		t.Error("expected 'By Directory' section in no-color output")
	}
	if !strings.Contains(output, "By Package") {
		t.Error("expected 'By Package' section in no-color output")
	}

	// Should have percentage columns
	if !strings.Contains(output, "%") {
		t.Error("expected '%' column in no-color output")
	}

	// Should have Total rows
	if !strings.Contains(output, "Total") {
		t.Error("expected 'Total' row in no-color output")
	}
}

func TestPercentageCalculations(t *testing.T) {
	// Create a summary with known values for percentage calculation
	summary := &Summary{
		TotalFiles:    3,
		TotalLines:    300,
		TotalCode:     200, // Total code for percentage base
		TotalBlanks:   50,
		TotalComments: 50,
		ByLanguage: map[string]*LanguageStats{
			"Go": {
				Language:  "Go",
				BaseStats: BaseStats{Files: 2, Lines: 200, Code: 150, Blanks: 25, Comments: 25}, // 75% of 200
			},
			"Python": {
				Language:  "Python",
				BaseStats: BaseStats{Files: 1, Lines: 100, Code: 50, Blanks: 25, Comments: 25}, // 25% of 200
			},
		},
		ByDirectory: make(map[string]*DirectoryStats),
		ByPackage:   make(map[string]*PackageStats),
	}

	config := &OutputConfig{
		Format:     FormatPretty,
		ByLanguage: true,
		NoColor:    true, // Disable colors to avoid ANSI escape sequences in test
	}

	output := FormatOutput(summary, config)

	// Check that percentages are calculated correctly
	// Go has 150/200 = 75%
	if !strings.Contains(output, "75.0%") {
		t.Errorf("expected '75.0%%' for Go language, got: %s", output)
	}

	// Python has 50/200 = 25%
	if !strings.Contains(output, "25.0%") {
		t.Errorf("expected '25.0%%' for Python language, got: %s", output)
	}

	// Total should be 100%
	if !strings.Contains(output, "100.0%") {
		t.Errorf("expected '100.0%%' for Total row, got: %s", output)
	}
}

func TestTotalsRowInBreakdownTables(t *testing.T) {
	summary := createTestSummary()

	t.Run("language totals", func(t *testing.T) {
		config := &OutputConfig{
			Format:     FormatPretty,
			ByLanguage: true,
			NoColor:    true,
		}

		output := FormatOutput(summary, config)

		// Check for Total row
		if !strings.Contains(output, "Total") {
			t.Error("expected 'Total' row in language breakdown")
		}

		// Check that the totals are summed correctly
		// Go: 42 + TypeScript: 28 + Python: 15 = 85 files
		// But we're using the total from createTestSummary which sums to 85
		if !strings.Contains(output, "85") {
			t.Errorf("expected total files count '85' in output, got: %s", output)
		}
	})

	t.Run("directory totals", func(t *testing.T) {
		config := &OutputConfig{
			Format:  FormatPretty,
			ByDir:   true,
			NoColor: true,
		}

		output := FormatOutput(summary, config)

		// Check for Total row
		if !strings.Contains(output, "Total") {
			t.Error("expected 'Total' row in directory breakdown")
		}
	})

	t.Run("package totals", func(t *testing.T) {
		config := &OutputConfig{
			Format:    FormatPretty,
			ByPackage: true,
			NoColor:   true,
		}

		output := FormatOutput(summary, config)

		// Check for Total row
		if !strings.Contains(output, "Total") {
			t.Error("expected 'Total' row in package breakdown")
		}
	})
}

func TestZeroTotalCodePercentage(t *testing.T) {
	// Test edge case where TotalCode is 0 (avoid division by zero)
	summary := &Summary{
		TotalFiles:    0,
		TotalLines:    0,
		TotalCode:     0, // Zero total code
		TotalBlanks:   0,
		TotalComments: 0,
		ByLanguage: map[string]*LanguageStats{
			"Go": {
				Language:  "Go",
				BaseStats: BaseStats{Files: 0, Lines: 0, Code: 0, Blanks: 0, Comments: 0},
			},
		},
		ByDirectory: make(map[string]*DirectoryStats),
		ByPackage:   make(map[string]*PackageStats),
	}

	config := &OutputConfig{
		Format:     FormatPretty,
		ByLanguage: true,
		NoColor:    true,
	}

	// This should not panic due to division by zero
	output := FormatOutput(summary, config)

	// Should handle 0/0 case gracefully (showing 0.0%)
	if !strings.Contains(output, "0.0%") {
		t.Errorf("expected '0.0%%' when total code is 0, got: %s", output)
	}
}

// TestElbowSeparatorsInSummaryTable tests that elbow separators appear in default output
func TestElbowSeparatorsInSummaryTable(t *testing.T) {
	summary := createTestSummary()
	config := &OutputConfig{
		Format:  FormatPretty,
		NoColor: false, // Use Unicode elbows
	}

	output := FormatOutput(summary, config)

	// Should have Unicode elbow characters for src/test sub-rows
	if !strings.Contains(output, ElbowMiddle) && !strings.Contains(output, ElbowLast) {
		t.Errorf("expected elbow separators in output, got: %s", output)
	}

	// Should have "src" and "test" sub-rows
	if !strings.Contains(output, "src") {
		t.Error("expected 'src' sub-row in output")
	}
	if !strings.Contains(output, "test") {
		t.Error("expected 'test' sub-row in output")
	}
}

// TestElbowSeparatorsASCII tests ASCII elbow separators in no-color mode
func TestElbowSeparatorsASCII(t *testing.T) {
	summary := createTestSummary()
	config := &OutputConfig{
		Format:  FormatPretty,
		NoColor: true, // Use ASCII elbows
	}

	output := FormatOutput(summary, config)

	// Should have ASCII elbow characters for src/test sub-rows
	if !strings.Contains(output, ElbowMiddleASCII) && !strings.Contains(output, ElbowLastASCII) {
		t.Errorf("expected ASCII elbow separators in no-color output, got: %s", output)
	}
}

// TestCombinedVsSeparatedOutput tests --combined flag behavior
func TestCombinedVsSeparatedOutput(t *testing.T) {
	summary := createTestSummary()

	t.Run("separated output (default)", func(t *testing.T) {
		config := &OutputConfig{
			Format:   FormatPretty,
			Combined: false, // Default: show src/test breakdown
			NoColor:  true,
		}

		output := FormatOutput(summary, config)

		// Should show Total row and sub-rows
		if !strings.Contains(output, "Total") {
			t.Error("expected 'Total' row in separated output")
		}
		if !strings.Contains(output, "src") {
			t.Error("expected 'src' sub-row in separated output")
		}
		if !strings.Contains(output, "test") {
			t.Error("expected 'test' sub-row in separated output")
		}
	})

	t.Run("combined output", func(t *testing.T) {
		config := &OutputConfig{
			Format:   FormatPretty,
			Combined: true, // Hide src/test breakdown
			NoColor:  true,
		}

		output := FormatOutput(summary, config)

		// Should NOT have elbow separators or breakdown labels
		if strings.Contains(output, ElbowMiddleASCII) || strings.Contains(output, ElbowLastASCII) {
			t.Error("did not expect elbow separators in combined output")
		}
		if strings.Contains(output, "+- src") || strings.Contains(output, "`- test") {
			t.Error("did not expect src/test breakdown labels in combined output")
		}
	})
}

// TestShowAllWithOtherFiles tests --all flag includes "other" category
func TestShowAllWithOtherFiles(t *testing.T) {
	summary := createTestSummaryWithOther()

	t.Run("pretty output with other files", func(t *testing.T) {
		config := &OutputConfig{
			Format:  FormatPretty,
			ShowAll: true,
			NoColor: true,
		}

		output := FormatOutput(summary, config)

		// Should show "other" sub-row
		if !strings.Contains(output, "other") {
			t.Errorf("expected 'other' sub-row with --all flag, got: %s", output)
		}
	})

	t.Run("JSON output with other files", func(t *testing.T) {
		config := &OutputConfig{
			Format:  FormatJSON,
			ShowAll: true,
		}

		output := FormatOutput(summary, config)

		// Should include "other" in JSON
		if !strings.Contains(output, `"other"`) {
			t.Errorf("expected 'other' field in JSON with --all flag, got: %s", output)
		}
	})

	t.Run("raw output with other files", func(t *testing.T) {
		config := &OutputConfig{
			Format:  FormatRaw,
			ShowAll: true,
		}

		output := FormatOutput(summary, config)

		// Should include "Other Files:" in raw output
		if !strings.Contains(output, "Other Files:") {
			t.Errorf("expected 'Other Files:' in raw output with --all flag, got: %s", output)
		}
	})
}

// TestTotalsMathSrcTestOther verifies that src + test + other = total
func TestTotalsMathSrcTestOther(t *testing.T) {
	summary := createTestSummaryWithOther()

	// Verify the math: src + test + other should equal total
	totalFiles := summary.SrcFiles + summary.TestFiles + summary.OtherFiles
	if totalFiles != summary.TotalFiles {
		t.Errorf("SrcFiles(%d) + TestFiles(%d) + OtherFiles(%d) = %d, want TotalFiles=%d",
			summary.SrcFiles, summary.TestFiles, summary.OtherFiles, totalFiles, summary.TotalFiles)
	}

	totalCode := summary.SrcCode + summary.TestCode + summary.OtherCode
	if totalCode != summary.TotalCode {
		t.Errorf("SrcCode(%d) + TestCode(%d) + OtherCode(%d) = %d, want TotalCode=%d",
			summary.SrcCode, summary.TestCode, summary.OtherCode, totalCode, summary.TotalCode)
	}

	totalLines := summary.SrcLines + summary.TestLines + summary.OtherLines
	if totalLines != summary.TotalLines {
		t.Errorf("SrcLines(%d) + TestLines(%d) + OtherLines(%d) = %d, want TotalLines=%d",
			summary.SrcLines, summary.TestLines, summary.OtherLines, totalLines, summary.TotalLines)
	}

	totalBlanks := summary.SrcBlanks + summary.TestBlanks + summary.OtherBlanks
	if totalBlanks != summary.TotalBlanks {
		t.Errorf("SrcBlanks(%d) + TestBlanks(%d) + OtherBlanks(%d) = %d, want TotalBlanks=%d",
			summary.SrcBlanks, summary.TestBlanks, summary.OtherBlanks, totalBlanks, summary.TotalBlanks)
	}

	totalComments := summary.SrcComments + summary.TestComments + summary.OtherComments
	if totalComments != summary.TotalComments {
		t.Errorf("SrcComments(%d) + TestComments(%d) + OtherComments(%d) = %d, want TotalComments=%d",
			summary.SrcComments, summary.TestComments, summary.OtherComments, totalComments, summary.TotalComments)
	}
}

// TestJSONOutputStructureWithSrcTestBreakdown tests JSON includes src/test by default
func TestJSONOutputStructureWithSrcTestBreakdown(t *testing.T) {
	summary := createTestSummary()
	config := &OutputConfig{
		Format:   FormatJSON,
		Combined: false, // Default
	}

	output := FormatOutput(summary, config)

	// Verify it's valid JSON
	var result jsonOutput
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	// Should have src and test fields at top level
	if result.Src == nil {
		t.Error("expected 'src' field in JSON output by default")
	}
	if result.Test == nil {
		t.Error("expected 'test' field in JSON output by default")
	}

	// Verify src stats
	if result.Src != nil {
		if result.Src.Files != summary.SrcFiles {
			t.Errorf("expected src.files = %d, got %d", summary.SrcFiles, result.Src.Files)
		}
		if result.Src.Code != summary.SrcCode {
			t.Errorf("expected src.code = %d, got %d", summary.SrcCode, result.Src.Code)
		}
	}

	// Verify test stats
	if result.Test != nil {
		if result.Test.Files != summary.TestFiles {
			t.Errorf("expected test.files = %d, got %d", summary.TestFiles, result.Test.Files)
		}
		if result.Test.Code != summary.TestCode {
			t.Errorf("expected test.code = %d, got %d", summary.TestCode, result.Test.Code)
		}
	}

	// "other" should be nil when no other files
	if result.Other != nil {
		t.Error("did not expect 'other' field when no other files")
	}
}

// TestJSONOutputCombinedOmitsSrcTest tests JSON with --combined omits src/test
func TestJSONOutputCombinedOmitsSrcTest(t *testing.T) {
	summary := createTestSummary()
	config := &OutputConfig{
		Format:   FormatJSON,
		Combined: true,
	}

	output := FormatOutput(summary, config)

	// Verify it's valid JSON
	var result jsonOutput
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	// Should NOT have src and test fields
	if result.Src != nil {
		t.Error("did not expect 'src' field in JSON output with --combined")
	}
	if result.Test != nil {
		t.Error("did not expect 'test' field in JSON output with --combined")
	}

	// Total should still be present
	if result.Total.Files != summary.TotalFiles {
		t.Errorf("expected total.files = %d, got %d", summary.TotalFiles, result.Total.Files)
	}
}

// TestJSONByLanguageIncludesSrcTestBreakdown tests language breakdown includes sub-stats
func TestJSONByLanguageIncludesSrcTestBreakdown(t *testing.T) {
	summary := createTestSummary()
	config := &OutputConfig{
		Format:   FormatJSON,
		Combined: false,
	}

	output := FormatOutput(summary, config)

	var result jsonOutput
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	// Find Go language stats
	var goLang *jsonLanguageStats
	for i := range result.ByLanguage {
		if result.ByLanguage[i].Language == "Go" {
			goLang = &result.ByLanguage[i]
			break
		}
	}

	if goLang == nil {
		t.Fatal("expected Go language in byLanguage")
	}

	// Go should have src and test sub-stats
	if goLang.Src == nil {
		t.Error("expected 'src' sub-stats in Go language entry")
	}
	if goLang.Test == nil {
		t.Error("expected 'test' sub-stats in Go language entry")
	}

	// Verify Go src stats match
	goStats := summary.ByLanguage["Go"]
	if goLang.Src != nil && goLang.Src.Files != goStats.SrcFiles {
		t.Errorf("expected Go src.files = %d, got %d", goStats.SrcFiles, goLang.Src.Files)
	}
}

// TestRawOutputWithSrcTestBreakdown tests raw output shows src/test by default
func TestRawOutputWithSrcTestBreakdown(t *testing.T) {
	summary := createTestSummary()
	config := &OutputConfig{
		Format:   FormatRaw,
		Combined: false, // Default
	}

	output := FormatOutput(summary, config)

	// Should show Source Files and Test Files
	if !strings.Contains(output, "Source Files:") {
		t.Errorf("expected 'Source Files:' in raw output, got: %s", output)
	}
	if !strings.Contains(output, "Source Code:") {
		t.Errorf("expected 'Source Code:' in raw output, got: %s", output)
	}
	if !strings.Contains(output, "Test Files:") {
		t.Errorf("expected 'Test Files:' in raw output, got: %s", output)
	}
	if !strings.Contains(output, "Test Code:") {
		t.Errorf("expected 'Test Code:' in raw output, got: %s", output)
	}
}

// TestRawOutputCombinedOmitsSrcTest tests raw output with --combined
func TestRawOutputCombinedOmitsSrcTest(t *testing.T) {
	summary := createTestSummary()
	config := &OutputConfig{
		Format:   FormatRaw,
		Combined: true,
	}

	output := FormatOutput(summary, config)

	// Should NOT show Source Files or Test Files breakdown
	if strings.Contains(output, "Source Files:") {
		t.Errorf("did not expect 'Source Files:' in combined raw output, got: %s", output)
	}
	if strings.Contains(output, "Test Files:") {
		t.Errorf("did not expect 'Test Files:' in combined raw output, got: %s", output)
	}

	// But should still show total Files
	if !strings.Contains(output, "Files:") {
		t.Errorf("expected 'Files:' in combined raw output, got: %s", output)
	}
}

// TestLanguageBreakdownTableShowsSubRows tests language table has src/test sub-rows
func TestLanguageBreakdownTableShowsSubRows(t *testing.T) {
	summary := createTestSummary()
	config := &OutputConfig{
		Format:     FormatPretty,
		ByLanguage: true,
		Combined:   false,
		NoColor:    true,
	}

	output := FormatOutput(summary, config)

	// Should show "By Language" section
	if !strings.Contains(output, "By Language") {
		t.Error("expected 'By Language' section")
	}

	// Should have elbow separators for sub-rows
	if !strings.Contains(output, ElbowMiddleASCII) || !strings.Contains(output, ElbowLastASCII) {
		t.Errorf("expected elbow separators in language table, got: %s", output)
	}

	// Should have "src" and "test" sub-rows
	if !strings.Contains(output, "src") {
		t.Error("expected 'src' sub-row in language table")
	}
	if !strings.Contains(output, "test") {
		t.Error("expected 'test' sub-row in language table")
	}
}

// TestLanguageBreakdownTableCombined tests language table with --combined
func TestLanguageBreakdownTableCombined(t *testing.T) {
	summary := createTestSummary()
	config := &OutputConfig{
		Format:     FormatPretty,
		ByLanguage: true,
		Combined:   true,
		NoColor:    true,
	}

	output := FormatOutput(summary, config)

	// Should show "By Language" section
	if !strings.Contains(output, "By Language") {
		t.Error("expected 'By Language' section")
	}

	// Should NOT have elbow separators
	if strings.Contains(output, ElbowMiddleASCII) || strings.Contains(output, ElbowLastASCII) {
		t.Errorf("did not expect elbow separators in combined language table, got: %s", output)
	}
}

// TestRawOutputByLanguageSubBreakdown tests raw output language breakdown shows sub-stats
func TestRawOutputByLanguageSubBreakdown(t *testing.T) {
	summary := createTestSummary()
	config := &OutputConfig{
		Format:     FormatRaw,
		ByLanguage: true,
		Combined:   false,
	}

	output := FormatOutput(summary, config)

	// Should have language with src/test sub-breakdown
	if !strings.Contains(output, "Go:") {
		t.Error("expected 'Go:' in language breakdown")
	}
	if !strings.Contains(output, "  src:") {
		t.Errorf("expected '  src:' sub-breakdown in language, got: %s", output)
	}
	if !strings.Contains(output, "  test:") {
		t.Errorf("expected '  test:' sub-breakdown in language, got: %s", output)
	}
}

// TestOutputConfigFields tests that OutputConfig has expected fields
func TestOutputConfigFields(t *testing.T) {
	config := &OutputConfig{
		Format:     FormatPretty,
		ByLanguage: true,
		ByDir:      true,
		ByPackage:  true,
		NoColor:    false,
		Combined:   false,
		ShowAll:    false,
	}

	// Verify fields can be set
	if config.Combined {
		t.Error("Combined should be false")
	}
	if config.ShowAll {
		t.Error("ShowAll should be false")
	}

	// Set new fields
	config.Combined = true
	config.ShowAll = true

	if !config.Combined {
		t.Error("Combined should be true after setting")
	}
	if !config.ShowAll {
		t.Error("ShowAll should be true after setting")
	}
}

func TestCodeOnlyRawOutput(t *testing.T) {
	summary := createTestSummary()

	config := &OutputConfig{
		Format:   FormatRaw,
		CodeOnly: true,
	}

	output := FormatOutput(summary, config)

	// CodeOnly should suppress Blanks and Comments lines
	if strings.Contains(output, "Blanks:") {
		t.Errorf("code-only output should not contain 'Blanks:', got: %s", output)
	}
	if strings.Contains(output, "Comments:") {
		t.Errorf("code-only output should not contain 'Comments:', got: %s", output)
	}
	// But should still show Files and Code
	if !strings.Contains(output, "Files:") {
		t.Errorf("code-only output should contain 'Files:', got: %s", output)
	}
	if !strings.Contains(output, "Code:") {
		t.Errorf("code-only output should contain 'Code:', got: %s", output)
	}
}
