package banner

import (
	"io"
	"strings"
	"text/template"

	"github.com/somya/git-banner-backend/internal/github"
)

// PinnedRenderer renders a pinned-repositories banner.
type PinnedRenderer struct {
	Data   *github.PinnedData
	Dims   Dimensions
	Colors Palette
	Theme  Theme
}

type repoCard struct {
	CX, CY, CW, CH        int
	IconX, IconY          int
	NameX, NameY, NameFS  int
	DescX, DescY1, DescY2 int
	DescLine1, DescLine2  string
	DotCX, DotCY          int
	LangX, LangY          int
	StarIX, StarIY        int
	StarX, StarY          int
	ForkIX, ForkIY        int
	ForkX, ForkY          int
	MetaFS                int
	Name, Lang            string
	Stars, Forks          string
	LangColor             string
	HasLang               bool
}

type pinnedRenderData struct {
	W, H                    int
	C                       Palette
	GridColor               string
	AvatarURL               string
	AvatarCX, AvatarCY      int
	AvatarR                 int
	LoginX, LoginY, LoginFS int
	Login                   string
	Cards                   []repoCard
	RepoPath                string
	StarPath                string
	ForkPath                string
}

const (
	octRepoPath = `M2 2.5A2.5 2.5 0 0 1 4.5 0h8.75a.75.75 0 0 1 .75.75v12.5a.75.75 0 0 1-.75.75h-2.5a.75.75 0 0 1 0-1.5h1.75v-2h-8a1 1 0 0 0-.714 1.7.75.75 0 1 1-1.072 1.05A2.495 2.495 0 0 1 2 11.5Zm10.5-1h-8a1 1 0 0 0-1 1v6.708A2.486 2.486 0 0 1 4.5 9h8ZM5 12.25a.25.25 0 0 1 .25-.25h3.5a.25.25 0 0 1 .25.25v3.25a.25.25 0 0 1-.4.2l-1.45-1.087a.249.249 0 0 0-.3 0L5.4 15.7a.25.25 0 0 1-.4-.2Z`
	octStarPath = `M8 .25a.75.75 0 0 1 .673.418l1.882 3.815 4.21.612a.75.75 0 0 1 .416 1.279l-3.046 2.97.719 4.192a.75.75 0 0 1-1.088.791L8 12.347l-3.766 1.98a.75.75 0 0 1-1.088-.79l.72-4.194L.818 6.374a.75.75 0 0 1 .416-1.28l4.21-.611L7.327.668A.75.75 0 0 1 8 .25Z`
	octForkPath = `M5 3.25a.75.75 0 1 1-1.5 0 .75.75 0 0 1 1.5 0zm0 2.122a2.25 2.25 0 1 0-1.5 0v.878A2.25 2.25 0 0 0 5.75 8.5h1.5v2.128a2.251 2.251 0 1 0 1.5 0V8.5h1.5a2.25 2.25 0 0 0 2.25-2.25v-.878a2.25 2.25 0 1 0-1.5 0v.878a.75.75 0 0 1-.75.75h-4.5A.75.75 0 0 1 5 6.25v-.878zm3.75 7.378a.75.75 0 1 1-1.5 0 .75.75 0 0 1 1.5 0zm3-8.75a.75.75 0 1 1-1.5 0 .75.75 0 0 1 1.5 0z`
)

var pinnedFuncs = template.FuncMap{
	"avImgX":  func(cx, r int) int { return cx - r },
	"avImgY":  func(cy, r int) int { return cy - r },
	"avImgSz": func(r int) int { return r * 2 },
}

var pinnedTmpl = template.Must(template.New("pinned").Funcs(pinnedFuncs).Parse(
	`<svg xmlns="http://www.w3.org/2000/svg" width="{{.W}}" height="{{.H}}" viewBox="0 0 {{.W}} {{.H}}">
  <defs>
    <pattern id="grid" x="0" y="0" width="40" height="40" patternUnits="userSpaceOnUse">
      <path d="M 40 0 L 0 0 0 40" fill="none" stroke="{{.GridColor}}" stroke-width="0.5" opacity="0.4"/>
    </pattern>
    <clipPath id="pav">
      <circle cx="{{.AvatarCX}}" cy="{{.AvatarCY}}" r="{{.AvatarR}}"/>
    </clipPath>
  </defs>
  <rect width="{{.W}}" height="{{.H}}" fill="{{.C.Background}}"/>
  <rect width="{{.W}}" height="{{.H}}" fill="url(#grid)"/>
  <!-- Avatar -->
  <circle cx="{{.AvatarCX}}" cy="{{.AvatarCY}}" r="{{.AvatarR}}" fill="{{.C.Border}}"/>
  {{if .AvatarURL}}<image href="{{.AvatarURL}}" x="{{avImgX .AvatarCX .AvatarR}}" y="{{avImgY .AvatarCY .AvatarR}}" width="{{avImgSz .AvatarR}}" height="{{avImgSz .AvatarR}}" clip-path="url(#pav)" preserveAspectRatio="xMidYMid slice"/>{{end}}
  <!-- @username -->
  <text x="{{.LoginX}}" y="{{.LoginY}}" font-family="'Segoe UI',Helvetica,Arial,sans-serif" font-size="{{.LoginFS}}" font-weight="400" fill="{{.C.Text}}">@{{.Login}}</text>
  {{range .Cards}}
  <rect x="{{.CX}}" y="{{.CY}}" width="{{.CW}}" height="{{.CH}}" rx="12" ry="12" fill="{{$.C.Surface}}"/>
  <g transform="translate({{.IconX}},{{.IconY}})"><path d="{{$.RepoPath}}" fill="{{$.C.Subtext}}"/></g>
  <text x="{{.NameX}}" y="{{.NameY}}" font-family="'Segoe UI',Helvetica,Arial,sans-serif" font-size="{{.NameFS}}" font-weight="700" fill="{{$.C.Text}}">{{.Name}}</text>
  {{if .DescLine1}}<text x="{{.DescX}}" y="{{.DescY1}}" font-family="'Segoe UI',Helvetica,Arial,sans-serif" font-size="13" fill="{{$.C.Subtext}}">{{.DescLine1}}</text>{{end}}
  {{if .DescLine2}}<text x="{{.DescX}}" y="{{.DescY2}}" font-family="'Segoe UI',Helvetica,Arial,sans-serif" font-size="13" fill="{{$.C.Subtext}}">{{.DescLine2}}</text>{{end}}
  {{if .HasLang}}<circle cx="{{.DotCX}}" cy="{{.DotCY}}" r="6" fill="{{.LangColor}}"/>
  <text x="{{.LangX}}" y="{{.LangY}}" font-family="'Segoe UI',Helvetica,Arial,sans-serif" font-size="{{.MetaFS}}" fill="{{$.C.Subtext}}">{{.Lang}}</text>{{end}}
  <g transform="translate({{.StarIX}},{{.StarIY}})"><path d="{{$.StarPath}}" fill="{{$.C.Subtext}}"/></g>
  <text x="{{.StarX}}" y="{{.StarY}}" font-family="'Segoe UI',Helvetica,Arial,sans-serif" font-size="{{.MetaFS}}" fill="{{$.C.Subtext}}">{{.Stars}}</text>
  <g transform="translate({{.ForkIX}},{{.ForkIY}})"><path d="{{$.ForkPath}}" fill="{{$.C.Subtext}}"/></g>
  <text x="{{.ForkX}}" y="{{.ForkY}}" font-family="'Segoe UI',Helvetica,Arial,sans-serif" font-size="{{.MetaFS}}" fill="{{$.C.Subtext}}">{{.Forks}}</text>
  {{end}}
</svg>`))

func (r *PinnedRenderer) Render(w io.Writer) error {
	W, H := r.Dims.Width, r.Dims.Height
	C := r.Colors

	repos := r.Data.Repos
	if len(repos) > 6 {
		repos = repos[:6]
	}
	n := len(repos)
	if n == 0 {
		n = 1
	}

	const (
		cols      = 2
		sidePad   = 45
		gapX      = 20
		gapY      = 16
		ipadH     = 22 // horizontal padding inside card
		ipadVT    = 22 // top padding inside card
		iconSz    = 16 // octicon native size (16×16)
		metaFS    = 13
		minTopPad = 18
		avatarR   = 12
		avatarCX  = 55
	)

	rows := (n + cols - 1) / cols
	cardW := (W - 2*sidePad - gapX) / cols

	// Card height: fill available space, capped at a comfortable max
	loginFS := scaledFontSize(W, 22)
	loginToCards := 16
	headerH := avatarR * 2
	availForCards := H - 2*minTopPad - headerH - loginToCards - (rows-1)*gapY
	cardH := availForCards / rows
	if cardH > 155 {
		cardH = 155
	}
	if cardH < 100 {
		cardH = 100
	}

	// Vertical centering with computed cardH
	contentH := headerH + loginToCards + rows*cardH + (rows-1)*gapY
	topPad := (H - contentH) / 2
	if topPad < minTopPad {
		topPad = minTopPad
	}

	avatarCY := topPad + avatarR
	loginX := avatarCX + avatarR + 12
	loginY := avatarCY + loginFS/2 - 2 // vertically centered with avatar
	firstCardY := topPad + headerH + loginToCards

	maxDescChars := (cardW - 2*ipadH) / 8

	cards := make([]repoCard, 0, n)
	for i, repo := range repos {
		col := i % cols
		row := i / cols
		cx := sidePad + col*(cardW+gapX)
		cy := firstCardY + row*(cardH+gapY)

		// Icon and name row
		iconX := cx + ipadH
		iconY := cy + ipadVT
		nameX := cx + ipadH + iconSz + 8
		nameY := cy + ipadVT + 13 // baseline centered with 16px icon

		// Description (below icon row)
		descX := cx + ipadH
		descY1 := cy + ipadVT + iconSz + 14 // 14px gap after icon
		descY2 := descY1 + 20

		// Meta row anchored to card bottom
		metaY := cy + cardH - 20
		dotCX := cx + ipadH + 6
		dotCY := metaY - 4
		langX := cx + ipadH + 16
		langY := metaY

		// Star and fork icons inline after language (fixed x offset)
		starIX := cx + ipadH + 16 + 105 // lang text area = 105px
		starIY := metaY - 13
		starX := starIX + iconSz + 4
		starY := metaY
		forkIX := starX + 34 // star count width ≈ 34px
		forkIY := starIY
		forkX := forkIX + iconSz + 4
		forkY := metaY

		// If no language, shift stars to the left
		if repo.Language == "" {
			starIX = cx + ipadH
			starIY = metaY - 13
			starX = starIX + iconSz + 4
			starY = metaY
			forkIX = starX + 34
			forkIY = starIY
			forkX = forkIX + iconSz + 4
			forkY = metaY
		}

		desc1, desc2 := wrapText(repo.Description, maxDescChars)

		cards = append(cards, repoCard{
			CX: cx, CY: cy, CW: cardW, CH: cardH,
			IconX: iconX, IconY: iconY,
			NameX: nameX, NameY: nameY, NameFS: scaledFontSize(W, 17),
			DescX: descX, DescY1: descY1, DescY2: descY2,
			DescLine1: desc1, DescLine2: desc2,
			DotCX: dotCX, DotCY: dotCY,
			LangX: langX, LangY: langY,
			StarIX: starIX, StarIY: starIY,
			StarX: starX, StarY: starY,
			ForkIX: forkIX, ForkIY: forkIY,
			ForkX: forkX, ForkY: forkY,
			MetaFS:    metaFS,
			Name:      repo.Name,
			Lang:      repo.Language,
			Stars:     formatInt(repo.Stars),
			Forks:     formatInt(repo.Forks),
			LangColor: pinnedLangColor(repo.Language),
			HasLang:   repo.Language != "",
		})
	}

	gridColor := "#21262d"
	if r.Theme == ThemeLight {
		gridColor = C.Border
	}

	return pinnedTmpl.Execute(w, pinnedRenderData{
		W: W, H: H,
		C:         C,
		GridColor: gridColor,
		AvatarURL: r.Data.AvatarDataURI,
		AvatarCX:  avatarCX,
		AvatarCY:  avatarCY,
		AvatarR:   avatarR,
		LoginX:    loginX,
		LoginY:    loginY,
		LoginFS:   loginFS,
		Login:     r.Data.User.Login,
		Cards:     cards,
		RepoPath:  octRepoPath,
		StarPath:  octStarPath,
		ForkPath:  octForkPath,
	})
}

// wrapText splits s into at most two lines at word boundaries.
func wrapText(s string, maxChars int) (line1, line2 string) {
	if maxChars < 10 {
		maxChars = 10
	}
	s = strings.TrimSpace(s)
	if len(s) <= maxChars {
		return s, ""
	}
	cut := maxChars
	for cut > 0 && s[cut-1] != ' ' {
		cut--
	}
	if cut == 0 {
		cut = maxChars
	}
	line1 = strings.TrimSpace(s[:cut])
	rest := strings.TrimSpace(s[cut:])
	if len(rest) > maxChars {
		rest = rest[:maxChars-3] + "..."
	}
	return line1, rest
}

// pinnedLangColor returns the GitHub-standard hex colour for a language.
func pinnedLangColor(lang string) string {
	colors := map[string]string{
		"Go":         "#00ADD8",
		"TypeScript": "#3178c6",
		"JavaScript": "#f1e05a",
		"Python":     "#3572A5",
		"Rust":       "#dea584",
		"Java":       "#b07219",
		"C++":        "#f34b7d",
		"C":          "#555555",
		"C#":         "#178600",
		"Ruby":       "#701516",
		"PHP":        "#4F5D95",
		"Swift":      "#F05138",
		"Kotlin":     "#A97BFF",
		"Dart":       "#00B4AB",
		"HTML":       "#e34c26",
		"CSS":        "#563d7c",
		"Shell":      "#89e051",
		"Vue":        "#41b883",
		"Svelte":     "#ff3e00",
	}
	if c, ok := colors[lang]; ok {
		return c
	}
	return "#8b949e"
}
