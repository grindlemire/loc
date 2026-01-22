package loc

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// Color palette constants for the LOC output styling.
// These colors are inspired by the Tokyo Night theme for a modern, readable appearance.
const (
	// Primary accent - subtle purple
	AccentColor = lipgloss.Color("#9D7CD8")

	// Text colors
	HeaderText = lipgloss.Color("#C0CAF5") // Light for headers
	DimText    = lipgloss.Color("#565F89") // Muted for secondary info
	NormalText = lipgloss.Color("#A9B1D6") // Standard text

	// Table colors
	BorderColor = lipgloss.Color("#414868") // Subtle border
	RowAltBg    = lipgloss.Color("#1A1B26") // Alternating row background (optional)
)

// Styles holds all the lipgloss styles used for rendering output.
// It supports both colored and no-color modes via the renderer.
type Styles struct {
	// Header style for section titles (e.g., "LOC - Lines of Code Counter")
	Header lipgloss.Style

	// SectionTitle style for table section headers (e.g., "By Language")
	SectionTitle lipgloss.Style

	// Label style for column headers and labels
	Label lipgloss.Style

	// Number style for numeric values
	Number lipgloss.Style

	// DimLabel style for secondary/muted labels
	DimLabel lipgloss.Style

	// TotalLabel style for the "Total:" label in summary
	TotalLabel lipgloss.Style

	// TotalNumber style for total numbers (uses accent color)
	TotalNumber lipgloss.Style

	// Border style configuration for tables
	Border lipgloss.Style

	// RowEven style for even rows (optional alternating background)
	RowEven lipgloss.Style

	// RowOdd style for odd rows
	RowOdd lipgloss.Style
}

// NewStyles creates a new Styles instance with the given renderer.
// Pass a renderer with lipgloss.WithColorProfile(termenv.Ascii) for no-color mode.
// Pass nil to use the default renderer with color support.
func NewStyles(r *lipgloss.Renderer) *Styles {
	if r == nil {
		r = lipgloss.DefaultRenderer()
	}

	return &Styles{
		Header: r.NewStyle().
			Bold(true).
			Foreground(HeaderText).
			MarginBottom(1),

		SectionTitle: r.NewStyle().
			Bold(true).
			Foreground(AccentColor),

		Label: r.NewStyle().
			Foreground(HeaderText),

		Number: r.NewStyle().
			Foreground(NormalText),

		DimLabel: r.NewStyle().
			Foreground(DimText),

		TotalLabel: r.NewStyle().
			Bold(true).
			Foreground(HeaderText),

		TotalNumber: r.NewStyle().
			Bold(true).
			Foreground(AccentColor),

		Border: r.NewStyle().
			Foreground(BorderColor),

		RowEven: r.NewStyle().
			Background(RowAltBg),

		RowOdd: r.NewStyle(),
	}
}

// DefaultStyles returns a Styles instance with default color settings.
func DefaultStyles() *Styles {
	return NewStyles(nil)
}

// NoColorStyles returns a Styles instance without any color formatting.
// This is useful for piping output or when NO_COLOR environment variable is set.
func NoColorStyles() *Styles {
	// Create a renderer with no color profile (Ascii = no colors)
	r := lipgloss.NewRenderer(nil)
	r.SetColorProfile(termenv.Ascii)
	return NewStyles(r)
}

// TableStyleConfig holds configuration for rendering tables with lipgloss.
type TableStyleConfig struct {
	// HeaderStyle is applied to the header row
	HeaderStyle lipgloss.Style

	// CellStyle is the default style for table cells
	CellStyle lipgloss.Style

	// BorderStyle is applied to table borders
	BorderStyle lipgloss.Style

	// UseAlternatingRows enables alternating row backgrounds
	UseAlternatingRows bool

	// EvenRowStyle is applied to even rows when UseAlternatingRows is true
	EvenRowStyle lipgloss.Style

	// OddRowStyle is applied to odd rows when UseAlternatingRows is true
	OddRowStyle lipgloss.Style
}

// NewTableStyleConfig creates a TableStyleConfig from the given Styles.
func NewTableStyleConfig(s *Styles) *TableStyleConfig {
	return &TableStyleConfig{
		HeaderStyle:        s.Label,
		CellStyle:          s.Number,
		BorderStyle:        s.Border,
		UseAlternatingRows: false, // Disabled by default for cleaner look
		EvenRowStyle:       s.RowEven,
		OddRowStyle:        s.RowOdd,
	}
}

// WithAlternatingRows returns a copy of the config with alternating rows enabled.
func (c *TableStyleConfig) WithAlternatingRows(enabled bool) *TableStyleConfig {
	c.UseAlternatingRows = enabled
	return c
}
