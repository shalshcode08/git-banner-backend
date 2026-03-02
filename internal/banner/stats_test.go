package banner

import (
	"bytes"
	"strings"
	"testing"

	"github.com/somya/git-banner-backend/internal/github"
)

func TestStatsRenderer_Render_Twitter(t *testing.T) {
	data := &github.StatsData{
		User: github.User{
			Login:       "octocat",
			Name:        "The Octocat",
			Followers:   1000,
			Following:   50,
			PublicRepos: 8,
		},
		TotalStars: 42,
	}
	r := &StatsRenderer{
		Data:   data,
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
	if !strings.Contains(out, "Followers") {
		t.Error("output missing Followers label")
	}
	if !strings.Contains(out, "1.0k") {
		t.Error("output missing formatted follower count (1000 → 1.0k)")
	}
}

func TestStatsRenderer_Render_LinkedIn(t *testing.T) {
	data := &github.StatsData{
		User:       github.User{Login: "dev", Name: "Dev User"},
		TotalStars: 5,
	}
	r := &StatsRenderer{
		Data:   data,
		Dims:   DimsFor(FormatLinkedIn),
		Colors: PaletteFor(ThemeLight),
	}

	var buf bytes.Buffer
	if err := r.Render(&buf); err != nil {
		t.Fatalf("Render returned error: %v", err)
	}
	out := buf.String()

	if !strings.Contains(out, `width="1584"`) {
		t.Error("output missing LinkedIn width")
	}
	if !strings.Contains(out, `height="396"`) {
		t.Error("output missing LinkedIn height")
	}
}

func TestStatsRenderer_Render_FallbackName(t *testing.T) {
	// When Name is empty, Login is used as display name
	data := &github.StatsData{
		User: github.User{Login: "nameless"},
	}
	r := &StatsRenderer{
		Data:   data,
		Dims:   DimsFor(FormatTwitter),
		Colors: PaletteFor(ThemeDark),
	}

	var buf bytes.Buffer
	if err := r.Render(&buf); err != nil {
		t.Fatalf("Render returned error: %v", err)
	}
	if !strings.Contains(buf.String(), "nameless") {
		t.Error("output missing login as fallback name")
	}
}

func TestStatsRenderer_Render_LongBioTruncated(t *testing.T) {
	longBio := strings.Repeat("a", 100)
	data := &github.StatsData{
		User: github.User{Login: "x", Bio: longBio},
	}
	r := &StatsRenderer{
		Data:   data,
		Dims:   DimsFor(FormatTwitter),
		Colors: PaletteFor(ThemeDark),
	}

	var buf bytes.Buffer
	if err := r.Render(&buf); err != nil {
		t.Fatalf("Render returned error: %v", err)
	}
	if strings.Contains(buf.String(), longBio) {
		t.Error("long bio should be truncated in output")
	}
	if !strings.Contains(buf.String(), "...") {
		t.Error("truncated bio should end with ...")
	}
}
