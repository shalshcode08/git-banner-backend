package banner

import (
	"io"
	"text/template"
	"time"

	"github.com/somya/git-banner-backend/internal/github"
)

// ContribRenderer renders a GitHub-style contribution graph banner.
type ContribRenderer struct {
	Data   *github.ContribData
	Dims   Dimensions
	Colors Palette
	Theme  Theme
}

type monthLabel struct {
	X, Y int
	Name string
}

type contribRenderData struct {
	W, H        int
	BoxX, BoxY  int
	BoxW, BoxH  int
	AvatarURL   string
	AvatarCX    int
	AvatarCY    int
	AvatarR     int
	Login       string
	Total       int
	TotalX      int
	TotalY      int
	MonthLabels []monthLabel
	Cells       []cell
	DayLabelX   int
	MonLabelY   int
	WedLabelY   int
	FriLabelY   int
	DotColor    string
	C           Palette
}

type cell struct {
	X, Y, S int
	Color   string
	Date    string
	Count   int
}

const contribTmplStr = `<svg xmlns="http://www.w3.org/2000/svg" width="{{.W}}" height="{{.H}}" viewBox="0 0 {{.W}} {{.H}}">
  <defs>
    <pattern id="dotgrid" x="0" y="0" width="30" height="30" patternUnits="userSpaceOnUse">
      <circle cx="1" cy="1" r="1.2" fill="{{.DotColor}}" fill-opacity="0.45"/>
    </pattern>
    <clipPath id="av">
      <circle cx="{{.AvatarCX}}" cy="{{.AvatarCY}}" r="{{.AvatarR}}"/>
    </clipPath>
  </defs>

  <!-- Background with subtle dot grid -->
  <rect width="{{.W}}" height="{{.H}}" fill="{{.C.Background}}"/>
  <rect width="{{.W}}" height="{{.H}}" fill="url(#dotgrid)"/>

  <!-- Avatar -->
  <circle cx="{{.AvatarCX}}" cy="{{.AvatarCY}}" r="{{.AvatarR}}" fill="{{.C.Border}}"/>
  {{if .AvatarURL}}<image href="{{.AvatarURL}}" x="{{imgX .AvatarCX .AvatarR}}" y="{{imgY .AvatarCY .AvatarR}}" width="{{imgSize .AvatarR}}" height="{{imgSize .AvatarR}}" clip-path="url(#av)" preserveAspectRatio="xMidYMid slice"/>{{end}}

  <!-- @username -->
  <text x="{{unameX .AvatarCX .AvatarR}}" y="{{unameY .AvatarCY}}" font-family="'Segoe UI',Helvetica,Arial,sans-serif" font-size="{{unameFS .W}}" font-weight="400" fill="{{.C.Text}}">@{{.Login}}</text>

  <!-- Heatmap container -->
  <rect x="{{.BoxX}}" y="{{.BoxY}}" width="{{.BoxW}}" height="{{.BoxH}}" rx="14" ry="14" fill="{{.C.Surface}}"/>

  <!-- Month labels -->
  {{range .MonthLabels}}<text x="{{.X}}" y="{{.Y}}" font-family="'Segoe UI',Helvetica,Arial,sans-serif" font-size="13" fill="{{$.C.Subtext}}">{{.Name}}</text>
  {{end}}
  <!-- Day labels -->
  <text x="{{.DayLabelX}}" y="{{.MonLabelY}}" font-family="'Segoe UI',Helvetica,Arial,sans-serif" font-size="13" fill="{{.C.Subtext}}" text-anchor="end">Mon</text>
  <text x="{{.DayLabelX}}" y="{{.WedLabelY}}" font-family="'Segoe UI',Helvetica,Arial,sans-serif" font-size="13" fill="{{.C.Subtext}}" text-anchor="end">Wed</text>
  <text x="{{.DayLabelX}}" y="{{.FriLabelY}}" font-family="'Segoe UI',Helvetica,Arial,sans-serif" font-size="13" fill="{{.C.Subtext}}" text-anchor="end">Fri</text>

  <!-- Contribution cells -->
  {{range .Cells}}<rect x="{{.X}}" y="{{.Y}}" width="{{.S}}" height="{{.S}}" rx="3" ry="3" fill="{{.Color}}"><title>{{.Date}}: {{.Count}} contributions</title></rect>
  {{end}}
  <!-- Total contributions -->
  <text x="{{.TotalX}}" y="{{.TotalY}}" font-family="'Segoe UI',Helvetica,Arial,sans-serif" font-size="14" fill="{{.C.Subtext}}" text-anchor="end">{{.Total}} contributions in the last year</text>
</svg>`

var contribFuncs = template.FuncMap{
	"imgX":    func(cx, r int) int { return cx - r },
	"imgY":    func(cy, r int) int { return cy - r },
	"imgSize": func(r int) int { return r * 2 },
	"unameX":  func(cx, r int) int { return cx + r + 12 },
	"unameY":  func(cy int) int { return cy + 5 },
	"unameFS": func(w int) int { return scaledFontSize(w, 16) },
}

var contribTmpl = template.Must(template.New("contrib").Funcs(contribFuncs).Parse(contribTmplStr))

func (r *ContribRenderer) Render(w io.Writer) error {
	W, H := r.Dims.Width, r.Dims.Height
	C := r.Colors

	// Avatar
	avatarR := 12
	avatarCX := 55
	boxX := 40
	boxW := W - boxX*2

	// Grid layout constants
	dayLabelColW := 82
	gridPadRight := 15
	monthFromTop := 26
	gridFromTop := 50

	weeks := r.Data.Weeks
	numWeeks := len(weeks)
	if numWeeks == 0 {
		numWeeks = 1
	}

	// Step is width-driven
	availW := boxW - dayLabelColW - gridPadRight
	step := availW / numWeeks
	if step > 28 {
		step = 28
	}
	if step < 5 {
		step = 5
	}
	cellSize := step - 3
	if cellSize < 3 {
		cellSize = 3
	}

	// Box height wraps grid snugly
	boxH := gridFromTop + 7*step + 32

	// Vertical centering: equal gap above avatar and below box
	headerGap := 10 // gap between avatar bottom and box top
	contentSpan := 2*avatarR + headerGap + boxH
	topPad := (H - contentSpan) / 2
	if topPad < 12 {
		topPad = 12
	}

	avatarCY := topPad + avatarR
	boxY := topPad + 2*avatarR + headerGap

	gridX := boxX + dayLabelColW
	gridY := boxY + gridFromTop

	cells := make([]cell, 0, numWeeks*7)
	var monthLabels []monthLabel
	prevMonth := -1

	for wi, week := range weeks {
		if len(week.Days) == 0 {
			continue
		}
		t, err := time.Parse("2006-01-02", week.Days[0].Date)
		if err == nil {
			m := int(t.Month())
			if m != prevMonth {
				monthLabels = append(monthLabels, monthLabel{
					X:    gridX + wi*step,
					Y:    boxY + monthFromTop,
					Name: t.Format("Jan"),
				})
				prevMonth = m
			}
		}
		for di, day := range week.Days {
			cells = append(cells, cell{
				X:     gridX + wi*step,
				Y:     gridY + di*step,
				S:     cellSize,
				Color: contribCellColor(day.Count, r.Theme),
				Date:  day.Date,
				Count: day.Count,
			})
		}
	}

	// Day label Y: centered on each row
	textBaseline := cellSize/2 + 5
	monLabelY := gridY + 1*step + textBaseline
	wedLabelY := gridY + 3*step + textBaseline
	friLabelY := gridY + 5*step + textBaseline

	dotColor := "#21262d"
	if r.Theme == ThemeLight {
		dotColor = C.Border
	}

	return contribTmpl.Execute(w, contribRenderData{
		W: W, H: H,
		BoxX: boxX, BoxY: boxY, BoxW: boxW, BoxH: boxH,
		AvatarURL:   r.Data.AvatarURL,
		AvatarCX:    avatarCX,
		AvatarCY:    avatarCY,
		AvatarR:     avatarR,
		Login:       r.Data.Login,
		Total:       r.Data.TotalContributions,
		TotalX:      boxX + boxW - 15,
		TotalY:      boxY + boxH - 15,
		MonthLabels: monthLabels,
		Cells:       cells,
		DayLabelX:   gridX - 8,
		MonLabelY:   monLabelY,
		WedLabelY:   wedLabelY,
		FriLabelY:   friLabelY,
		DotColor:    dotColor,
		C:           C,
	})
}

// contribCellColor maps a contribution count to a theme-appropriate colour.
func contribCellColor(count int, t Theme) string {
	if t == ThemeLight {
		switch {
		case count == 0:
			return "#ebedf0"
		case count < 5:
			return "#9be9a8"
		case count < 15:
			return "#40c463"
		case count < 30:
			return "#30a14e"
		default:
			return "#216e39"
		}
	}
	// dark theme (GitHub dark green palette)
	switch {
	case count == 0:
		return "#21262d"
	case count < 5:
		return "#0e4429"
	case count < 15:
		return "#006d32"
	case count < 30:
		return "#26a641"
	default:
		return "#39d353"
	}
}
