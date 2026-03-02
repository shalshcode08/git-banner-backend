package banner

import (
	"fmt"
	"io"
)

// Dimensions defines banner canvas size in pixels.
type Dimensions struct {
	Width  int
	Height int
}

// Format represents the target social platform.
type Format string

const (
	FormatTwitter  Format = "twitter"
	FormatLinkedIn Format = "linkedin"
)

// Theme controls the colour palette.
type Theme string

const (
	ThemeDark  Theme = "dark"
	ThemeLight Theme = "light"
)

// DimsFor returns the pixel dimensions for a given format.
func DimsFor(f Format) Dimensions {
	if f == FormatLinkedIn {
		return Dimensions{Width: 1584, Height: 396}
	}
	return Dimensions{Width: 1500, Height: 500}
}

// Palette holds the colour tokens for a theme.
type Palette struct {
	Background string
	Surface    string
	Border     string
	Text       string
	Subtext    string
	Accent     string
	Green      string
}

// PaletteFor returns the colour palette for a given theme.
func PaletteFor(t Theme) Palette {
	if t == ThemeLight {
		return Palette{
			Background: "#ffffff",
			Surface:    "#f6f8fa",
			Border:     "#d0d7de",
			Text:       "#1f2328",
			Subtext:    "#656d76",
			Accent:     "#0969da",
			Green:      "#2da44e",
		}
	}
	// dark (default)
	return Palette{
		Background: "#0d1117",
		Surface:    "#161b22",
		Border:     "#30363d",
		Text:       "#e6edf3",
		Subtext:    "#8b949e",
		Accent:     "#58a6ff",
		Green:      "#3fb950",
	}
}

// Renderer writes an SVG banner to w.
type Renderer interface {
	Render(w io.Writer) error
}

// scaledFontSize scales a base font size proportionally to banner width.
func scaledFontSize(bannerWidth, base int) int {
	scaled := base * bannerWidth / 1500
	if scaled < 10 {
		return 10
	}
	return scaled
}

// formatInt formats an integer with K suffix for values >= 1000.
func formatInt(n int) string {
	if n >= 1000 {
		return fmt.Sprintf("%.1fk", float64(n)/1000)
	}
	return fmt.Sprintf("%d", n)
}
