package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// Client wraps an HTTP client and handles GitHub REST + GraphQL requests.
type Client struct {
	http        *http.Client
	token       string
	cache       *Cache
	restBase    string
	graphqlBase string
}

// NewClient creates a Client with the given token and cache TTL.
func NewClient(token string, cacheTTL time.Duration) *Client {
	return &Client{
		http:        &http.Client{Timeout: 10 * time.Second},
		token:       token,
		cache:       NewCache(cacheTTL),
		restBase:    "https://api.github.com",
		graphqlBase: "https://api.github.com/graphql",
	}
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
}

// FetchUser fetches public user profile data via REST.
func (c *Client) FetchUser(ctx context.Context, username string) (*User, error) {
	key := username + ":user"
	if v, ok := c.cache.Get(key); ok {
		u := v.(User)
		return &u, nil
	}

	url := fmt.Sprintf("%s/users/%s", c.restBase, username)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("user %q not found", username)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github REST error: %s", resp.Status)
	}

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	c.cache.Set(key, user)
	return &user, nil
}

// FetchStats fetches user profile + star count via REST.
func (c *Client) FetchStats(ctx context.Context, username string) (*StatsData, error) {
	key := username + ":stats"
	if v, ok := c.cache.Get(key); ok {
		d := v.(StatsData)
		return &d, nil
	}

	user, err := c.FetchUser(ctx, username)
	if err != nil {
		return nil, err
	}

	stars, err := c.fetchTotalStars(ctx, username)
	if err != nil {
		return nil, err
	}

	data := StatsData{
		User:       *user,
		TotalStars: stars,
	}
	c.cache.Set(key, data)
	return &data, nil
}

// fetchTotalStars iterates paginated repos and sums stargazer counts.
func (c *Client) fetchTotalStars(ctx context.Context, username string) (int, error) {
	total := 0
	page := 1
	for {
		url := fmt.Sprintf("%s/users/%s/repos?per_page=100&page=%d", c.restBase, username, page)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return 0, err
		}
		c.setHeaders(req)

		resp, err := c.http.Do(req)
		if err != nil {
			return 0, err
		}

		var repos []struct {
			Stars int  `json:"stargazers_count"`
			Fork  bool `json:"fork"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
			resp.Body.Close()
			return 0, err
		}
		resp.Body.Close()

		for _, r := range repos {
			if !r.Fork {
				total += r.Stars
			}
		}
		if len(repos) < 100 {
			break
		}
		page++
	}
	return total, nil
}

// doGraphQL executes a GraphQL query against the GitHub API.
// It requires a token; callers should check c.token before calling.
func (c *Client) doGraphQL(ctx context.Context, query string, variables map[string]any, out any) error {
	if c.token == "" {
		return fmt.Errorf("GITHUB_TOKEN is required for GraphQL queries (pinned repos and contributions)")
	}

	body, err := json.Marshal(map[string]any{
		"query":     query,
		"variables": variables,
	})
	if err != nil {
		return fmt.Errorf("marshal graphql request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.graphqlBase, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	c.setHeaders(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		slog.Error("github graphql non-200", "status", resp.Status, "body", string(raw))
		if resp.StatusCode == http.StatusUnauthorized {
			return fmt.Errorf("github GraphQL: 401 Unauthorized — set GITHUB_TOKEN env var")
		}
		return fmt.Errorf("github GraphQL error: %s", resp.Status)
	}

	var envelope struct {
		Data   json.RawMessage `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return fmt.Errorf("decode graphql envelope: %w", err)
	}
	if len(envelope.Errors) > 0 {
		return fmt.Errorf("graphql error: %s", envelope.Errors[0].Message)
	}
	if out != nil {
		if err := json.Unmarshal(envelope.Data, out); err != nil {
			return fmt.Errorf("decode graphql data: %w", err)
		}
	}
	return nil
}

// FetchPinned fetches pinned repositories via GraphQL.
func (c *Client) FetchPinned(ctx context.Context, username string) (*PinnedData, error) {
	key := username + ":pinned"
	if v, ok := c.cache.Get(key); ok {
		d := v.(PinnedData)
		return &d, nil
	}

	user, err := c.FetchUser(ctx, username)
	if err != nil {
		return nil, err
	}

	const q = `query($login: String!) {
		user(login: $login) {
			pinnedItems(first: 6, types: REPOSITORY) {
				nodes {
					... on Repository {
						name
						description
						url
						stargazerCount
						forkCount
						primaryLanguage { name }
					}
				}
			}
		}
	}`

	var result struct {
		User struct {
			PinnedItems struct {
				Nodes []struct {
					Name        string `json:"name"`
					Description string `json:"description"`
					URL         string `json:"url"`
					Stars       int    `json:"stargazerCount"`
					Forks       int    `json:"forkCount"`
					Language    struct {
						Name string `json:"name"`
					} `json:"primaryLanguage"`
				} `json:"nodes"`
			} `json:"pinnedItems"`
		} `json:"user"`
	}

	if err := c.doGraphQL(ctx, q, map[string]any{"login": username}, &result); err != nil {
		return nil, err
	}

	repos := make([]PinnedRepo, 0, len(result.User.PinnedItems.Nodes))
	for _, n := range result.User.PinnedItems.Nodes {
		repos = append(repos, PinnedRepo{
			Name:        n.Name,
			Description: n.Description,
			Language:    n.Language.Name,
			Stars:       n.Stars,
			Forks:       n.Forks,
			URL:         n.URL,
		})
	}

	data := PinnedData{User: *user, Repos: repos}
	c.cache.Set(key, data)
	return &data, nil
}

// FetchContributions fetches the contribution calendar via GraphQL.
func (c *Client) FetchContributions(ctx context.Context, username string) (*ContribData, error) {
	key := username + ":contributions"
	if v, ok := c.cache.Get(key); ok {
		d := v.(ContribData)
		return &d, nil
	}

	const q = `query($login: String!) {
		user(login: $login) {
			contributionsCollection {
				contributionCalendar {
					totalContributions
					weeks {
						contributionDays {
							date
							contributionCount
							color
						}
					}
				}
			}
		}
	}`

	var result struct {
		User struct {
			ContributionsCollection struct {
				ContributionCalendar struct {
					TotalContributions int `json:"totalContributions"`
					Weeks              []struct {
						ContributionDays []struct {
							Date  string `json:"date"`
							Count int    `json:"contributionCount"`
							Color string `json:"color"`
						} `json:"contributionDays"`
					} `json:"weeks"`
				} `json:"contributionCalendar"`
			} `json:"contributionsCollection"`
		} `json:"user"`
	}

	if err := c.doGraphQL(ctx, q, map[string]any{"login": username}, &result); err != nil {
		return nil, err
	}

	cal := result.User.ContributionsCollection.ContributionCalendar
	weeks := make([]ContribWeek, 0, len(cal.Weeks))
	for _, w := range cal.Weeks {
		days := make([]ContribDay, 0, len(w.ContributionDays))
		for _, d := range w.ContributionDays {
			days = append(days, ContribDay{
				Date:  d.Date,
				Count: d.Count,
				Color: d.Color,
			})
		}
		weeks = append(weeks, ContribWeek{Days: days})
	}

	data := ContribData{
		Login:              username,
		TotalContributions: cal.TotalContributions,
		Weeks:              weeks,
	}
	c.cache.Set(key, data)
	return &data, nil
}
