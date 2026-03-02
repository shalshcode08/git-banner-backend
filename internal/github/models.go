package github

// User holds public profile data from the GitHub REST API.
type User struct {
	Login       string `json:"login"`
	Name        string `json:"name"`
	AvatarURL   string `json:"avatar_url"`
	Bio         string `json:"bio"`
	Followers   int    `json:"followers"`
	Following   int    `json:"following"`
	PublicRepos int    `json:"public_repos"`
}

// PinnedRepo holds data for a single pinned repository.
type PinnedRepo struct {
	Name        string
	Description string
	Language    string
	Stars       int
	Forks       int
	URL         string
}

// ContribDay holds contribution data for a single calendar day.
type ContribDay struct {
	Date  string
	Count int
	Color string
}

// ContribWeek holds a slice of days (Sunday–Saturday).
type ContribWeek struct {
	Days []ContribDay
}

// ContribData holds the full contribution calendar for a user.
type ContribData struct {
	Login              string
	TotalContributions int
	Weeks              []ContribWeek
}

// StatsData aggregates all data needed for the stats banner.
type StatsData struct {
	User         User
	TotalStars   int
	TotalCommits int
}

// PinnedData aggregates all data needed for the pinned banner.
type PinnedData struct {
	User  User
	Repos []PinnedRepo
}
