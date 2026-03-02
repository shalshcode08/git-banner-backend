package handler

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/somya/git-banner-backend/internal/github"
)

// mockFetcher satisfies the Fetcher interface with controllable responses.
type mockFetcher struct {
	statsData   *github.StatsData
	pinnedData  *github.PinnedData
	contribData *github.ContribData
	err         error
}

func (m *mockFetcher) FetchStats(_ context.Context, _ string) (*github.StatsData, error) {
	return m.statsData, m.err
}

func (m *mockFetcher) FetchPinned(_ context.Context, _ string) (*github.PinnedData, error) {
	return m.pinnedData, m.err
}

func (m *mockFetcher) FetchContributions(_ context.Context, _ string) (*github.ContribData, error) {
	return m.contribData, m.err
}

// defaultMock returns a mock with valid data for all types.
func defaultMock() *mockFetcher {
	return &mockFetcher{
		statsData: &github.StatsData{
			User:       github.User{Login: "octocat", Name: "The Octocat"},
			TotalStars: 42,
		},
		pinnedData: &github.PinnedData{
			User:  github.User{Login: "octocat"},
			Repos: []github.PinnedRepo{{Name: "hello-world", Stars: 10}},
		},
		contribData: &github.ContribData{
			Login:              "octocat",
			TotalContributions: 365,
		},
	}
}

func doRequest(h *BannerHandler, method, path string, pathValues map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	for k, v := range pathValues {
		req.SetPathValue(k, v)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	return w
}

func TestBannerHandler_Stats_Success(t *testing.T) {
	h := NewBannerHandler(defaultMock())
	w := doRequest(h, http.MethodGet, "/banner/octocat?type=stats", map[string]string{"username": "octocat"})

	if w.Code != http.StatusOK {
		t.Errorf("status: expected %d, got %d", http.StatusOK, w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "image/svg+xml" {
		t.Errorf("Content-Type: expected image/svg+xml, got %s", ct)
	}
	if !strings.Contains(w.Body.String(), "<svg") {
		t.Error("body missing <svg")
	}
}

func TestBannerHandler_Pinned_Success(t *testing.T) {
	h := NewBannerHandler(defaultMock())
	w := doRequest(h, http.MethodGet, "/banner/octocat?type=pinned", map[string]string{"username": "octocat"})

	if w.Code != http.StatusOK {
		t.Errorf("status: expected %d, got %d", http.StatusOK, w.Code)
	}
	if !strings.Contains(w.Body.String(), "<svg") {
		t.Error("body missing <svg")
	}
}

func TestBannerHandler_Contributions_Success(t *testing.T) {
	h := NewBannerHandler(defaultMock())
	w := doRequest(h, http.MethodGet, "/banner/octocat?type=contributions", map[string]string{"username": "octocat"})

	if w.Code != http.StatusOK {
		t.Errorf("status: expected %d, got %d", http.StatusOK, w.Code)
	}
	if !strings.Contains(w.Body.String(), "<svg") {
		t.Error("body missing <svg")
	}
}

func TestBannerHandler_DefaultType_IsStats(t *testing.T) {
	// no type param → defaults to stats
	h := NewBannerHandler(defaultMock())
	w := doRequest(h, http.MethodGet, "/banner/octocat", map[string]string{"username": "octocat"})

	if w.Code != http.StatusOK {
		t.Errorf("status: expected %d, got %d; body: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestBannerHandler_BadType_400(t *testing.T) {
	h := NewBannerHandler(defaultMock())
	w := doRequest(h, http.MethodGet, "/banner/octocat?type=badtype", map[string]string{"username": "octocat"})

	if w.Code != http.StatusBadRequest {
		t.Errorf("status: expected %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestBannerHandler_MissingUsername_400(t *testing.T) {
	h := NewBannerHandler(defaultMock())
	// no SetPathValue → PathValue("username") returns ""
	req := httptest.NewRequest(http.MethodGet, "/banner/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status: expected %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestBannerHandler_NoToken_401(t *testing.T) {
	mock := &mockFetcher{err: fmt.Errorf("GITHUB_TOKEN is required for GraphQL queries")}
	h := NewBannerHandler(mock)
	w := doRequest(h, http.MethodGet, "/banner/octocat?type=stats", map[string]string{"username": "octocat"})

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status: expected %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestBannerHandler_UserNotFound_404(t *testing.T) {
	mock := &mockFetcher{err: fmt.Errorf("user %q not found", "ghost")}
	h := NewBannerHandler(mock)
	w := doRequest(h, http.MethodGet, "/banner/ghost?type=stats", map[string]string{"username": "ghost"})

	if w.Code != http.StatusNotFound {
		t.Errorf("status: expected %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestBannerHandler_GitHubError_502(t *testing.T) {
	mock := &mockFetcher{err: fmt.Errorf("github REST error: 503 Service Unavailable")}
	h := NewBannerHandler(mock)
	w := doRequest(h, http.MethodGet, "/banner/octocat?type=stats", map[string]string{"username": "octocat"})

	if w.Code != http.StatusBadGateway {
		t.Errorf("status: expected %d, got %d", http.StatusBadGateway, w.Code)
	}
}

func TestBannerHandler_PinnedError_401(t *testing.T) {
	mock := &mockFetcher{err: fmt.Errorf("GITHUB_TOKEN required")}
	h := NewBannerHandler(mock)
	w := doRequest(h, http.MethodGet, "/banner/octocat?type=pinned", map[string]string{"username": "octocat"})

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status: expected %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestBannerHandler_ContribError_404(t *testing.T) {
	mock := &mockFetcher{err: fmt.Errorf("user not found")}
	h := NewBannerHandler(mock)
	w := doRequest(h, http.MethodGet, "/banner/ghost?type=contributions", map[string]string{"username": "ghost"})

	if w.Code != http.StatusNotFound {
		t.Errorf("status: expected %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestBannerHandler_CacheControlHeader(t *testing.T) {
	h := NewBannerHandler(defaultMock())
	w := doRequest(h, http.MethodGet, "/banner/octocat?type=stats", map[string]string{"username": "octocat"})

	cc := w.Header().Get("Cache-Control")
	if !strings.Contains(cc, "max-age") {
		t.Errorf("Cache-Control: expected max-age, got %q", cc)
	}
}
