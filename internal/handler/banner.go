package handler

import (
	"context"
	"log/slog"
	"net/http"
	"regexp"
	"strings"

	"github.com/somya/git-banner-backend/internal/banner"
	"github.com/somya/git-banner-backend/internal/github"
)

// validUsername matches GitHub's username rules:
// 1–39 chars, alphanumeric or single interior hyphens, no leading/trailing hyphen.
var validUsername = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]{0,37}[a-zA-Z0-9])?$`)

var (
	validTypes   = map[string]bool{"stats": true, "contributions": true, "pinned": true}
	validFormats = map[string]bool{"twitter": true, "linkedin": true}
	validThemes  = map[string]bool{"dark": true, "light": true}
)

// Fetcher is the interface satisfied by *github.Client.
// It is defined here so handlers can be tested with a mock.
type Fetcher interface {
	FetchStats(ctx context.Context, username string) (*github.StatsData, error)
	FetchPinned(ctx context.Context, username string) (*github.PinnedData, error)
	FetchContributions(ctx context.Context, username string) (*github.ContribData, error)
}

// BannerHandler holds the Fetcher needed to serve banner requests.
type BannerHandler struct {
	gh Fetcher
}

// NewBannerHandler creates a BannerHandler with the given Fetcher.
func NewBannerHandler(gh Fetcher) *BannerHandler {
	return &BannerHandler{gh: gh}
}

// ServeHTTP handles GET /banner/{username}.
//
// Query params:
//
//	type   — stats (default) | contributions | pinned
//	format — twitter (default) | linkedin
//	theme  — dark (default) | light
func (h *BannerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	if !validUsername.MatchString(username) {
		http.Error(w, "invalid username", http.StatusBadRequest)
		return
	}

	q := r.URL.Query()

	bannerType := q.Get("type")
	if bannerType == "" {
		bannerType = "stats"
	} else if !validTypes[bannerType] {
		http.Error(w, "invalid type: must be stats, contributions, or pinned", http.StatusBadRequest)
		return
	}

	rawFormat := q.Get("format")
	if rawFormat == "" {
		rawFormat = "twitter"
	} else if !validFormats[rawFormat] {
		http.Error(w, "invalid format: must be twitter or linkedin", http.StatusBadRequest)
		return
	}
	format := banner.Format(rawFormat)

	rawTheme := q.Get("theme")
	if rawTheme == "" {
		rawTheme = "dark"
	} else if !validThemes[rawTheme] {
		http.Error(w, "invalid theme: must be dark or light", http.StatusBadRequest)
		return
	}
	theme := banner.Theme(rawTheme)

	dims := banner.DimsFor(format)
	colors := banner.PaletteFor(theme)

	var renderer banner.Renderer

	switch bannerType {
	case "contributions":
		data, err := h.gh.FetchContributions(r.Context(), username)
		if err != nil {
			writeGitHubError(w, r, "fetch contributions failed", username, err)
			return
		}
		renderer = &banner.ContribRenderer{Data: data, Dims: dims, Colors: colors, Theme: theme}

	case "pinned":
		data, err := h.gh.FetchPinned(r.Context(), username)
		if err != nil {
			writeGitHubError(w, r, "fetch pinned failed", username, err)
			return
		}
		renderer = &banner.PinnedRenderer{Data: data, Dims: dims, Colors: colors, Theme: theme}

	case "stats":
		data, err := h.gh.FetchStats(r.Context(), username)
		if err != nil {
			writeGitHubError(w, r, "fetch stats failed", username, err)
			return
		}
		renderer = &banner.StatsRenderer{Data: data, Dims: dims, Colors: colors, Theme: theme}

	}

	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "public, max-age=300")
	if err := renderer.Render(w); err != nil {
		slog.Error("render failed", "username", username, "type", bannerType, "error", err)
	}
}

// writeGitHubError logs the error and writes an appropriate HTTP error response.
// Token-missing errors become 401; not-found becomes 404; everything else is 502.
func writeGitHubError(w http.ResponseWriter, _ *http.Request, msg, username string, err error) {
	slog.Error(msg, "username", username, "error", err)

	errStr := err.Error()
	switch {
	case strings.Contains(errStr, "GITHUB_TOKEN"):
		http.Error(w, errStr, http.StatusUnauthorized)
	case strings.Contains(errStr, "not found"):
		http.Error(w, errStr, http.StatusNotFound)
	default:
		http.Error(w, "failed to fetch GitHub data: "+errStr, http.StatusBadGateway)
	}
}
