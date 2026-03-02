package handler

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/somya/git-banner-backend/internal/banner"
	"github.com/somya/git-banner-backend/internal/github"
)

// BannerHandler holds the GitHub client needed to serve banner requests.
type BannerHandler struct {
	gh *github.Client
}

// NewBannerHandler creates a BannerHandler with the given GitHub client.
func NewBannerHandler(gh *github.Client) *BannerHandler {
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
	if username == "" {
		http.Error(w, "username is required", http.StatusBadRequest)
		return
	}

	q := r.URL.Query()
	bannerType := q.Get("type")
	if bannerType == "" {
		bannerType = "stats"
	}
	format := banner.Format(q.Get("format"))
	if format == "" {
		format = banner.FormatTwitter
	}
	theme := banner.Theme(q.Get("theme"))
	if theme == "" {
		theme = banner.ThemeDark
	}

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
		renderer = &banner.ContribRenderer{Data: data, Dims: dims, Colors: colors}

	case "pinned":
		data, err := h.gh.FetchPinned(r.Context(), username)
		if err != nil {
			writeGitHubError(w, r, "fetch pinned failed", username, err)
			return
		}
		renderer = &banner.PinnedRenderer{Data: data, Dims: dims, Colors: colors}

	case "stats":
		data, err := h.gh.FetchStats(r.Context(), username)
		if err != nil {
			writeGitHubError(w, r, "fetch stats failed", username, err)
			return
		}
		renderer = &banner.StatsRenderer{Data: data, Dims: dims, Colors: colors}

	default:
		http.Error(w, "invalid type: must be stats, contributions, or pinned", http.StatusBadRequest)
		return
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
