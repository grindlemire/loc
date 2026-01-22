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
		ByLanguage: map[string]*LanguageStats{
			"Go": {
				Language: "Go",
				Files:    42,
				Lines:    3521,
				Code:     2890,
				Blanks:   319,
				Comments: 312,
			},
			"TypeScript": {
				Language: "TypeScript",
				Files:    28,
				Lines:    2104,
				Code:     1756,
				Blanks:   150,
				Comments: 198,
			},
			"Python": {
				Language: "Python",
				Files:    15,
				Lines:    892,
				Code:     723,
				Blanks:   80,
				Comments: 89,
			},
		},
		ByDirectory: map[string]*DirectoryStats{
			"pkg/api": {
				Path:     "pkg/api",
				Files:    18,
				Lines:    1892,
				Code:     1580,
				Blanks:   180,
				Comments: 132,
			},
			"pkg/core": {
				Path:     "pkg/core",
				Files:    24,
				Lines:    1629,
				Code:     1350,
				Blanks:   150,
				Comments: 129,
			},
			"cmd": {
				Path:     "cmd",
				Files:    12,
				Lines:    945,
				Code:     800,
				Blanks:   90,
				Comments: 55,
			},
		},
		ByPackage: map[string]*PackageStats{
			"api": {
				Package:  "api",
				Files:    18,
				Lines:    1892,
				Code:     1580,
				Blanks:   180,
				Comments: 132,
			},
			"core": {
				Package:  "core",
				Files:    24,
				Lines:    1629,
				Code:     1350,
				Blanks:   150,
				Comments: 129,
			},
		},
	}
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

	// Check header is present
	if !strings.Contains(output, "LOC - Lines of Code Counter") {
		t.Error("expected header in pretty output")
	}

	// Check summary line with formatted numbers
	if !strings.Contains(output, "90 files") {
		t.Error("expected '90 files' in output")
	}
	if !strings.Contains(output, "6,673 lines") {
		t.Error("expected '6,673 lines' in output")
	}
	if !strings.Contains(output, "5,511 code") {
		t.Error("expected '5,511 code' in output")
	}
	if !strings.Contains(output, "607 comments") {
		t.Error("expected '607 comments' in output")
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

	// Check box-drawing characters are present
	if !strings.Contains(output, "\u250c") { // top-left corner
		t.Error("expected box-drawing characters in output")
	}
	if !strings.Contains(output, "\u2502") { // vertical line
		t.Error("expected vertical line characters in output")
	}
	if !strings.Contains(output, "\u2518") { // bottom-right corner
		t.Error("expected box-drawing characters in output")
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

	// Check summary line is still present
	if !strings.Contains(output, "Total:") {
		t.Error("expected 'Total:' summary line")
	}
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

func TestJSONOutputAlwaysIncludesFullData(t *testing.T) {
	summary := createTestSummary()

	// Test with different flag combinations - JSON should always include all data
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
			}

			output := FormatOutput(summary, config)

			var result jsonOutput
			if err := json.Unmarshal([]byte(output), &result); err != nil {
				t.Fatalf("invalid JSON: %v", err)
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

func TestBreakdownFlagsAffectPrettyAndRawNotJSON(t *testing.T) {
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

	// JSON output should NOT differ
	configNoFlags.Format = FormatJSON
	configAllFlags.Format = FormatJSON

	jsonNoFlags := FormatOutput(summary, configNoFlags)
	jsonAllFlags := FormatOutput(summary, configAllFlags)

	if jsonNoFlags != jsonAllFlags {
		t.Error("expected JSON output to be the same regardless of breakdown flags")
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
		config := &OutputConfig{Format: FormatPretty, ByLanguage: true}
		output := FormatOutput(summary, config)
		if !strings.Contains(output, "0 files") {
			t.Error("expected '0 files' in output")
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

	// Should produce pretty output (with header)
	if !strings.Contains(output, "LOC - Lines of Code Counter") {
		t.Error("expected pretty output for empty/default format")
	}
}

func TestPrettyOutputHorizontalLine(t *testing.T) {
	summary := createTestSummary()
	config := &OutputConfig{Format: FormatPretty}

	output := FormatOutput(summary, config)

	// Check horizontal line character is present
	if !strings.Contains(output, "\u2500") {
		t.Error("expected horizontal line character in output")
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

	// Check formatted numbers
	if !strings.Contains(output, "12,345 files") {
		t.Error("expected '12,345 files' in output")
	}
	if !strings.Contains(output, "1,234,567 lines") {
		t.Error("expected '1,234,567 lines' in output")
	}
	if !strings.Contains(output, "987,654 code") {
		t.Error("expected '987,654 code' in output")
	}
}
