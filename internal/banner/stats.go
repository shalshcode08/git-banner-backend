package banner

import (
	"io"
	"text/template"

	"github.com/somya/git-banner-backend/internal/github"
)

// StatsRenderer renders a GitHub stats banner.
type StatsRenderer struct {
	Data   *github.StatsData
	Dims   Dimensions
	Colors Palette
	Theme  Theme
}

type statTile struct {
	IconX, IconY   int // top-left of 16×16 icon
	ValueX, ValueY int // text-anchor=middle
	LabelX, LabelY int // text-anchor=middle
	ValueFS        int
	Value, Label   string
	IconPath       string
}

type statsDivider struct {
	X, Y1, Y2 int
}

type statsRenderData struct {
	W, H                        int
	C                           Palette
	DotColor                    string
	AvatarURL                   string
	AvatarCX, AvatarCY, AvatarR int
	LoginX, LoginY, LoginFS     int
	Login                       string
	BoxX, BoxY, BoxW, BoxH      int
	Dividers                    []statsDivider
	Tiles                       []statTile
}

// octPersonPath is the GitHub "person" Octicon (16×16).
const octPersonPath = `M10.561 8.073a6.005 6.005 0 0 1 3.432 5.142.75.75 0 1 1-1.498.07 4.5 4.5 0 0 0-8.99 0 .75.75 0 0 1-1.498-.07 6.004 6.004 0 0 1 3.431-5.142 3.5 3.5 0 1 1 5.123 0ZM10.5 5a2 2 0 1 0-4 0 2 2 0 0 0 4 0Z`

var statsAvFuncs = template.FuncMap{
	"savImgX":  func(cx, r int) int { return cx - r },
	"savImgY":  func(cy, r int) int { return cy - r },
	"savImgSz": func(r int) int { return r * 2 },
}

var statsTmpl = template.Must(template.New("stats").Funcs(statsAvFuncs).Parse(
	`<svg xmlns="http://www.w3.org/2000/svg" width="{{.W}}" height="{{.H}}" viewBox="0 0 {{.W}} {{.H}}">
  <defs>
    <pattern id="dotgrid" x="0" y="0" width="30" height="30" patternUnits="userSpaceOnUse">
      <circle cx="1" cy="1" r="1.2" fill="{{.DotColor}}" fill-opacity="0.45"/>
    </pattern>
    <clipPath id="sav">
      <circle cx="{{.AvatarCX}}" cy="{{.AvatarCY}}" r="{{.AvatarR}}"/>
    </clipPath>
  </defs>

  <!-- Background -->
  <rect width="{{.W}}" height="{{.H}}" fill="{{.C.Background}}"/>
  <rect width="{{.W}}" height="{{.H}}" fill="url(#dotgrid)"/>

  <!-- Avatar -->
  <circle cx="{{.AvatarCX}}" cy="{{.AvatarCY}}" r="{{.AvatarR}}" fill="{{.C.Border}}"/>
  {{if .AvatarURL}}<image href="{{.AvatarURL}}" x="{{savImgX .AvatarCX .AvatarR}}" y="{{savImgY .AvatarCY .AvatarR}}" width="{{savImgSz .AvatarR}}" height="{{savImgSz .AvatarR}}" clip-path="url(#sav)" preserveAspectRatio="xMidYMid slice"/>{{end}}

  <!-- @username -->
  <text x="{{.LoginX}}" y="{{.LoginY}}" font-family="'Segoe UI',Helvetica,Arial,sans-serif" font-size="{{.LoginFS}}" font-weight="600" fill="{{.C.Text}}">@{{.Login}}</text>

  <!-- Stats container -->
  <rect x="{{.BoxX}}" y="{{.BoxY}}" width="{{.BoxW}}" height="{{.BoxH}}" rx="14" ry="14" fill="{{.C.Surface}}"/>

  <!-- Column dividers -->
  {{range .Dividers}}<line x1="{{.X}}" y1="{{.Y1}}" x2="{{.X}}" y2="{{.Y2}}" stroke="{{$.C.Border}}" stroke-width="1"/>
  {{end}}

  <!-- Stat tiles -->
  {{range .Tiles}}
  <g transform="translate({{.IconX}},{{.IconY}})"><path d="{{.IconPath}}" fill="{{$.C.Subtext}}"/></g>
  <text x="{{.ValueX}}" y="{{.ValueY}}" text-anchor="middle" font-family="'Segoe UI',Helvetica,Arial,sans-serif" font-size="{{.ValueFS}}" font-weight="700" fill="{{$.C.Text}}">{{.Value}}</text>
  <text x="{{.LabelX}}" y="{{.LabelY}}" text-anchor="middle" font-family="'Segoe UI',Helvetica,Arial,sans-serif" font-size="14" fill="{{$.C.Subtext}}">{{.Label}}</text>
  {{end}}
</svg>`))

func (r *StatsRenderer) Render(w io.Writer) error {
	W, H := r.Dims.Width, r.Dims.Height
	C := r.Colors
	u := r.Data.User

	const (
		avatarR    = 12
		avatarCX   = 55
		minTopPad  = 18
		boxSidePad = 40
		headerGap  = 14
		maxBoxH    = 290
	)

	boxW := W - 2*boxSidePad
	headerH := avatarR * 2
	loginFS := scaledFontSize(W, 22)

	// Box height: fill available space up to max
	availForBox := H - 2*minTopPad - headerH - headerGap
	boxH := availForBox
	if boxH > maxBoxH {
		boxH = maxBoxH
	}

	// Vertical centering
	contentH := headerH + headerGap + boxH
	topPad := (H - contentH) / 2
	if topPad < minTopPad {
		topPad = minTopPad
		boxH = H - 2*minTopPad - headerH - headerGap
	}

	avatarCY := topPad + avatarR
	loginX := avatarCX + avatarR + 12
	loginY := avatarCY + loginFS/2 - 2
	boxX := boxSidePad
	boxY := topPad + headerH + headerGap

	// 4 equal columns inside the box
	colW := boxW / 4
	valFS := scaledFontSize(W, 50)

	// Content block inside each column, vertically centered
	innerH := 16 + 12 + valFS + 10 + 14 // icon + gap + number + gap + label
	innerTop := boxY + (boxH-innerH)/2
	if innerTop < boxY+12 {
		innerTop = boxY + 12
	}
	numY := innerTop + 16 + 12 + valFS
	labelY := numY + 10 + 14

	type entry struct {
		label string
		value string
		icon  string
	}
	entries := []entry{
		{"Repositories", formatInt(u.PublicRepos), octRepoPath},
		{"Total Stars", formatInt(r.Data.TotalStars), octStarPath},
		{"Followers", formatInt(u.Followers), octPersonPath},
		{"Following", formatInt(u.Following), octPersonPath},
	}

	tiles := make([]statTile, len(entries))
	for i, e := range entries {
		colCX := boxX + i*colW + colW/2
		tiles[i] = statTile{
			IconX:    colCX - 8, // center 16px icon
			IconY:    innerTop,
			ValueX:   colCX,
			ValueY:   numY,
			LabelX:   colCX,
			LabelY:   labelY,
			ValueFS:  valFS,
			Value:    e.value,
			Label:    e.label,
			IconPath: e.icon,
		}
	}

	dividers := make([]statsDivider, 3)
	for i := 1; i <= 3; i++ {
		dividers[i-1] = statsDivider{
			X:  boxX + i*colW,
			Y1: boxY + 20,
			Y2: boxY + boxH - 20,
		}
	}

	dotColor := "#21262d"
	if r.Theme == ThemeLight {
		dotColor = C.Border
	}

	return statsTmpl.Execute(w, statsRenderData{
		W: W, H: H,
		C:         C,
		DotColor:  dotColor,
		AvatarURL: r.Data.AvatarDataURI,
		AvatarCX:  avatarCX,
		AvatarCY:  avatarCY,
		AvatarR:   avatarR,
		LoginX:    loginX,
		LoginY:    loginY,
		LoginFS:   loginFS,
		Login:     u.Login,
		BoxX:      boxX,
		BoxY:      boxY,
		BoxW:      boxW,
		BoxH:      boxH,
		Dividers:  dividers,
		Tiles:     tiles,
	})
}
