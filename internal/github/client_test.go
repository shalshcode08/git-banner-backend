package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// newTestClient creates a Client pointing at srv for both REST and GraphQL.
func newTestClient(srv *httptest.Server, token string) *Client {
	return &Client{
		http:        &http.Client{Timeout: 5 * time.Second},
		token:       token,
		cache:       NewCache(5 * time.Minute),
		restBase:    srv.URL,
		graphqlBase: srv.URL + "/graphql",
	}
}

// ---- FetchUser ----

func TestFetchUser_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/users/octocat" {
			json.NewEncoder(w).Encode(User{
				Login:       "octocat",
				Name:        "The Octocat",
				Followers:   100,
				Following:   10,
				PublicRepos: 8,
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	c := newTestClient(srv, "")
	u, err := c.FetchUser(context.Background(), "octocat")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.Login != "octocat" {
		t.Errorf("Login: expected %q, got %q", "octocat", u.Login)
	}
	if u.Followers != 100 {
		t.Errorf("Followers: expected 100, got %d", u.Followers)
	}
}

func TestFetchUser_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	c := newTestClient(srv, "")
	_, err := c.FetchUser(context.Background(), "ghost")
	if err == nil {
		t.Fatal("expected error for 404, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got: %v", err)
	}
}

func TestFetchUser_NonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "server error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := newTestClient(srv, "")
	_, err := c.FetchUser(context.Background(), "octocat")
	if err == nil {
		t.Fatal("expected error for 500, got nil")
	}
}

func TestFetchUser_Cache(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		json.NewEncoder(w).Encode(User{Login: "octocat"})
	}))
	defer srv.Close()

	c := newTestClient(srv, "")
	c.FetchUser(context.Background(), "octocat") //nolint:errcheck
	c.FetchUser(context.Background(), "octocat") //nolint:errcheck

	if callCount != 1 {
		t.Errorf("expected 1 HTTP call (cache hit on 2nd), got %d", callCount)
	}
}

// ---- FetchStats ----

func TestFetchStats_AggregatesStars(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/users/octocat":
			json.NewEncoder(w).Encode(User{Login: "octocat", PublicRepos: 3})
		case "/users/octocat/repos":
			repos := []struct {
				Stars int  `json:"stargazers_count"`
				Fork  bool `json:"fork"`
			}{
				{Stars: 10, Fork: false},
				{Stars: 5, Fork: true}, // fork — should be excluded
				{Stars: 7, Fork: false},
			}
			json.NewEncoder(w).Encode(repos)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(srv, "")
	data, err := c.FetchStats(context.Background(), "octocat")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data.TotalStars != 17 {
		t.Errorf("TotalStars: expected 17 (excludes fork), got %d", data.TotalStars)
	}
	if data.User.Login != "octocat" {
		t.Errorf("User.Login: expected %q, got %q", "octocat", data.User.Login)
	}
}

func TestFetchStats_UserNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	c := newTestClient(srv, "")
	_, err := c.FetchStats(context.Background(), "ghost")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ---- doGraphQL / FetchPinned ----

func TestDoGraphQL_NoToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not make HTTP call when token is empty")
	}))
	defer srv.Close()

	c := newTestClient(srv, "") // no token
	err := c.doGraphQL(context.Background(), "query {}", nil, nil)
	if err == nil {
		t.Fatal("expected error for missing token")
	}
	if !strings.Contains(err.Error(), "GITHUB_TOKEN") {
		t.Errorf("expected GITHUB_TOKEN in error, got: %v", err)
	}
}

func TestDoGraphQL_401(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}))
	defer srv.Close()

	c := newTestClient(srv, "bad-token")
	err := c.doGraphQL(context.Background(), "query {}", nil, nil)
	if err == nil {
		t.Fatal("expected error for 401")
	}
}

func TestDoGraphQL_GraphQLErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"errors": []map[string]any{{"message": "something went wrong"}},
		})
	}))
	defer srv.Close()

	c := newTestClient(srv, "token")
	err := c.doGraphQL(context.Background(), "query {}", nil, nil)
	if err == nil {
		t.Fatal("expected error for graphql errors array")
	}
	if !strings.Contains(err.Error(), "something went wrong") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestFetchPinned_Success(t *testing.T) {
	type gqlBody struct {
		Query     string            `json:"query"`
		Variables map[string]string `json:"variables"`
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/users/octocat":
			json.NewEncoder(w).Encode(User{Login: "octocat"})
		case "/graphql":
			resp := map[string]any{
				"data": map[string]any{
					"user": map[string]any{
						"pinnedItems": map[string]any{
							"nodes": []map[string]any{
								{
									"name":           "hello-world",
									"description":    "My first repo",
									"url":            "https://github.com/octocat/hello-world",
									"stargazerCount": 42,
									"forkCount":      5,
									"primaryLanguage": map[string]any{"name": "Go"},
								},
							},
						},
					},
				},
			}
			json.NewEncoder(w).Encode(resp)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := newTestClient(srv, "test-token")
	data, err := c.FetchPinned(context.Background(), "octocat")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data.Repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(data.Repos))
	}
	if data.Repos[0].Name != "hello-world" {
		t.Errorf("repo name: expected %q, got %q", "hello-world", data.Repos[0].Name)
	}
	if data.Repos[0].Stars != 42 {
		t.Errorf("stars: expected 42, got %d", data.Repos[0].Stars)
	}
	if data.Repos[0].Language != "Go" {
		t.Errorf("language: expected %q, got %q", "Go", data.Repos[0].Language)
	}
}

func TestFetchPinned_RequiresToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/users/octocat" {
			json.NewEncoder(w).Encode(User{Login: "octocat"})
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	c := newTestClient(srv, "") // no token
	_, err := c.FetchPinned(context.Background(), "octocat")
	if err == nil {
		t.Fatal("expected error for missing token")
	}
	if !strings.Contains(err.Error(), "GITHUB_TOKEN") {
		t.Errorf("expected GITHUB_TOKEN in error, got: %v", err)
	}
}

// ---- FetchContributions ----

func TestFetchContributions_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/graphql" {
			http.NotFound(w, r)
			return
		}
		resp := map[string]any{
			"data": map[string]any{
				"user": map[string]any{
					"contributionsCollection": map[string]any{
						"contributionCalendar": map[string]any{
							"totalContributions": 365,
							"weeks": []map[string]any{
								{
									"contributionDays": []map[string]any{
										{"date": "2024-01-01", "contributionCount": 3, "color": "#216e39"},
									},
								},
							},
						},
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := newTestClient(srv, "test-token")
	data, err := c.FetchContributions(context.Background(), "octocat")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data.TotalContributions != 365 {
		t.Errorf("TotalContributions: expected 365, got %d", data.TotalContributions)
	}
	if data.Login != "octocat" {
		t.Errorf("Login: expected %q, got %q", "octocat", data.Login)
	}
	if len(data.Weeks) != 1 || len(data.Weeks[0].Days) != 1 {
		t.Errorf("unexpected week/day structure: %+v", data.Weeks)
	}
	if data.Weeks[0].Days[0].Count != 3 {
		t.Errorf("day count: expected 3, got %d", data.Weeks[0].Days[0].Count)
	}
}

func TestFetchContributions_RequiresToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer srv.Close()

	c := newTestClient(srv, "") // no token
	_, err := c.FetchContributions(context.Background(), "octocat")
	if err == nil {
		t.Fatal("expected error for missing token")
	}
}
