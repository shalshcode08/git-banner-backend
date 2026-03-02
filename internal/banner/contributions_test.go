package banner

import (
	"bytes"
	"strings"
	"testing"

	"github.com/somya/git-banner-backend/internal/github"
)

func makeTestContribData(total int, weeks []github.ContribWeek) *github.ContribData {
	return &github.ContribData{
		Login:              "octocat",
		TotalContributions: total,
		Weeks:              weeks,
	}
}

func TestContribRenderer_Render_Basic(t *testing.T) {
	weeks := []github.ContribWeek{
		{Days: []github.ContribDay{
			{Date: "2024-01-01", Count: 3, Color: "#216e39"},
			{Date: "2024-01-02", Count: 0, Color: "#ebedf0"},
		}},
	}
	r := &ContribRenderer{
		Data:   makeTestContribData(365, weeks),
		Dims:   DimsFor(FormatTwitter),
		Colors: PaletteFor(ThemeDark),
	}

	var buf bytes.Buffer
	if err := r.Render(&buf); err != nil {
		t.Fatalf("Render returned error: %v", err)
	}
	out := buf.String()

	if !strings.Contains(out, "<svg") {
		t.Error("output missing <svg tag")
	}
	if !strings.Contains(out, `width="1500"`) {
		t.Error("output missing correct width")
	}
	if !strings.Contains(out, `height="500"`) {
		t.Error("output missing correct height")
	}
	if !strings.Contains(out, "octocat") {
		t.Error("output missing username")
	}
	if !strings.Contains(out, "365") {
		t.Error("output missing total contributions")
	}
}

func TestContribRenderer_Render_EmptyWeeks(t *testing.T) {
	r := &ContribRenderer{
		Data:   makeTestContribData(0, nil),
		Dims:   DimsFor(FormatTwitter),
		Colors: PaletteFor(ThemeDark),
	}

	var buf bytes.Buffer
	if err := r.Render(&buf); err != nil {
		t.Fatalf("Render error on empty weeks: %v", err)
	}
	if !strings.Contains(buf.String(), "<svg") {
		t.Error("output missing <svg tag with empty weeks")
	}
}

func TestContribRenderer_Render_CellColors(t *testing.T) {
	weeks := []github.ContribWeek{
		{Days: []github.ContribDay{
			{Date: "2024-01-01", Count: 5, Color: "#39d353"},
		}},
	}
	r := &ContribRenderer{
		Data:   makeTestContribData(5, weeks),
		Dims:   DimsFor(FormatTwitter),
		Colors: PaletteFor(ThemeDark),
	}

	var buf bytes.Buffer
	if err := r.Render(&buf); err != nil {
		t.Fatalf("Render returned error: %v", err)
	}
	if !strings.Contains(buf.String(), "#39d353") {
		t.Error("output missing contribution cell color")
	}
}

func TestContribRenderer_Render_EmptyDayColor(t *testing.T) {
	// When day.Color is empty, should fall back to surface color
	weeks := []github.ContribWeek{
		{Days: []github.ContribDay{
			{Date: "2024-01-01", Count: 0, Color: ""},
		}},
	}
	r := &ContribRenderer{
		Data:   makeTestContribData(0, weeks),
		Dims:   DimsFor(FormatTwitter),
		Colors: PaletteFor(ThemeDark),
	}

	var buf bytes.Buffer
	if err := r.Render(&buf); err != nil {
		t.Fatalf("Render returned error: %v", err)
	}
	// The surface color should appear as fallback
	dark := PaletteFor(ThemeDark)
	if !strings.Contains(buf.String(), dark.Surface) {
		t.Errorf("output missing fallback surface color %s", dark.Surface)
	}
}

func TestContribRenderer_Render_LinkedIn(t *testing.T) {
	r := &ContribRenderer{
		Data:   makeTestContribData(100, nil),
		Dims:   DimsFor(FormatLinkedIn),
		Colors: PaletteFor(ThemeLight),
	}

	var buf bytes.Buffer
	if err := r.Render(&buf); err != nil {
		t.Fatalf("Render returned error: %v", err)
	}
	if !strings.Contains(buf.String(), `width="1584"`) {
		t.Error("output missing LinkedIn width")
	}
}
