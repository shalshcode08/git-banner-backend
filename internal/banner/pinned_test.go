package banner

import (
	"bytes"
	"strings"
	"testing"

	"github.com/somya/git-banner-backend/internal/github"
)

func makeTestPinnedData(repos []github.PinnedRepo) *github.PinnedData {
	return &github.PinnedData{
		User:  github.User{Login: "octocat"},
		Repos: repos,
	}
}

func TestPinnedRenderer_Render_WithRepos(t *testing.T) {
	repos := []github.PinnedRepo{
		{Name: "hello-world", Description: "My first repo", Language: "Go", Stars: 42, Forks: 3},
		{Name: "cli-tool", Description: "A CLI", Language: "Python", Stars: 5, Forks: 1},
	}
	r := &PinnedRenderer{
		Data:   makeTestPinnedData(repos),
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
	if !strings.Contains(out, "octocat") {
		t.Error("output missing username")
	}
	if !strings.Contains(out, "hello-world") {
		t.Error("output missing repo name")
	}
	if !strings.Contains(out, "cli-tool") {
		t.Error("output missing second repo name")
	}
}

func TestPinnedRenderer_Render_EmptyRepos(t *testing.T) {
	r := &PinnedRenderer{
		Data:   makeTestPinnedData(nil),
		Dims:   DimsFor(FormatTwitter),
		Colors: PaletteFor(ThemeDark),
	}

	var buf bytes.Buffer
	if err := r.Render(&buf); err != nil {
		t.Fatalf("Render error on empty repos: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "<svg") {
		t.Error("output missing <svg tag even with empty repos")
	}
}

func TestPinnedRenderer_Render_MaxSixRepos(t *testing.T) {
	repos := make([]github.PinnedRepo, 8)
	for i := range repos {
		repos[i] = github.PinnedRepo{Name: strings.Repeat("x", i+1)}
	}
	r := &PinnedRenderer{
		Data:   makeTestPinnedData(repos),
		Dims:   DimsFor(FormatTwitter),
		Colors: PaletteFor(ThemeDark),
	}

	var buf bytes.Buffer
	if err := r.Render(&buf); err != nil {
		t.Fatalf("Render returned error: %v", err)
	}
	// repos[6] and repos[7] names should not appear
	out := buf.String()
	if strings.Contains(out, "xxxxxxxx") { // repos[7] = 8 chars
		t.Error("output should cap at 6 repos")
	}
}

func TestPinnedRenderer_Render_LinkedIn(t *testing.T) {
	r := &PinnedRenderer{
		Data:   makeTestPinnedData([]github.PinnedRepo{{Name: "repo1"}}),
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
