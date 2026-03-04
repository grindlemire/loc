package loc

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// Color palette constants for the LOC output styling.
// These colors are vibrant and modern for clear visibility on terminal.
const (
	// Primary accent - vibrant cyan
	AccentColor = lipgloss.Color("#00D9FF")

	// Text colors
	HeaderText = lipgloss.Color("#F8FAFC") // Bright white for headers
	NormalText = lipgloss.Color("#E2E8F0") // Light gray for standard text

	// Table colors
	BorderColor = lipgloss.Color("#6366F1") // Vibrant indigo border

	// Sub-row colors
	SubRowDim = lipgloss.Color("#64748B") // Dimmer color for sub-row elements
)

// Tree-style elbow separator constants for sub-row display.
// These are used to show src/test/other breakdowns under language rows.
const (
	// ElbowMiddle is used for non-last items in the sub-row list
	ElbowMiddle = "\u251C\u2500 " // "├─ "

	// ElbowLast is used for the last item in the sub-row list
	ElbowLast = "\u2514\u2500 " // "└─ "

	// ASCII fallback versions for when colors/unicode are disabled
	ElbowMiddleASCII = "+- "
	ElbowLastASCII   = "`- "
)

// Styles holds all the lipgloss styles used for rendering output.
// It supports both colored and no-color modes via the renderer.
type Styles struct {
	// SectionTitle style for table section headers (e.g., "By Language")
	SectionTitle lipgloss.Style

	// Label style for column headers and labels
	Label lipgloss.Style

	// Number style for numeric values
	Number lipgloss.Style

	// TotalLabel style for the "Total:" label in summary
	TotalLabel lipgloss.Style

	// TotalNumber style for total numbers (uses accent color)
	TotalNumber lipgloss.Style

	// Border style configuration for tables
	Border lipgloss.Style

	// SubRowLabel style for sub-row labels ("src", "test", "other")
	// Slightly indented appearance with dim styling
	SubRowLabel lipgloss.Style
}

// NewStyles creates a new Styles instance with the given renderer.
// Pass a renderer with lipgloss.WithColorProfile(termenv.Ascii) for no-color mode.
// Pass nil to use the default renderer with color support.
func NewStyles(r *lipgloss.Renderer) *Styles {
	if r == nil {
		r = lipgloss.DefaultRenderer()
	}

	return &Styles{
		SectionTitle: r.NewStyle().
			Bold(true).
			Foreground(AccentColor),

		Label: r.NewStyle().
			Foreground(HeaderText),

		Number: r.NewStyle().
			Foreground(NormalText),

		TotalLabel: r.NewStyle().
			Bold(true).
			Foreground(HeaderText),

		TotalNumber: r.NewStyle().
			Bold(true).
			Foreground(AccentColor),

		Border: r.NewStyle().
			Foreground(BorderColor),

		SubRowLabel: r.NewStyle().
			Foreground(SubRowDim).
			PaddingLeft(1),
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
