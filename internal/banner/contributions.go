package banner

import (
	"html/template"
	"io"

	"github.com/somya/git-banner-backend/internal/github"
)

// ContribRenderer renders a contribution graph banner.
type ContribRenderer struct {
	Data   *github.ContribData
	Dims   Dimensions
	Colors Palette
}

var contribTmpl = template.Must(template.New("contrib").Parse(`<svg xmlns="http://www.w3.org/2000/svg" width="{{.W}}" height="{{.H}}" viewBox="0 0 {{.W}} {{.H}}">
  <defs>
    <linearGradient id="bg" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" style="stop-color:{{.C.Background}};stop-opacity:1"/>
      <stop offset="100%" style="stop-color:{{.C.Surface}};stop-opacity:1"/>
    </linearGradient>
  </defs>

  <rect width="{{.W}}" height="{{.H}}" fill="url(#bg)" rx="16" ry="16"/>
  <rect width="{{.W}}" height="{{.H}}" fill="none" stroke="{{.C.Border}}" stroke-width="2" rx="16" ry="16"/>
  <rect x="0" y="0" width="6" height="{{.H}}" fill="{{.C.Accent}}" rx="3" ry="3"/>

  <!-- title -->
  <text x="40" y="{{.TitleY}}" font-family="'Segoe UI',Helvetica,Arial,sans-serif" font-size="{{.TitleSize}}" font-weight="700" fill="{{.C.Text}}">{{.Login}}'s Contributions</text>
  <text x="40" y="{{.SubtitleY}}" font-family="'Segoe UI',Helvetica,Arial,sans-serif" font-size="{{.SubSize}}" fill="{{.C.Subtext}}">{{.Total}} contributions in the last year</text>

  <!-- contribution grid -->
  {{range .Cells}}
  <rect x="{{.X}}" y="{{.Y}}" width="{{.S}}" height="{{.S}}" rx="2" ry="2" fill="{{.Color}}">
    <title>{{.Date}}: {{.Count}} contributions</title>
  </rect>
  {{end}}
</svg>`))

type cell struct {
	X, Y, S int
	Color   string
	Date    string
	Count   int
}

type contribRenderData struct {
	W, H               int
	C                  Palette
	Login              string
	Total              int
	TitleY, TitleSize  int
	SubtitleY, SubSize int
	Cells              []cell
}

func (r *ContribRenderer) Render(w io.Writer) error {
	W, H := r.Dims.Width, r.Dims.Height
	C := r.Colors

	// Layout
	headerH := 90
	padL := 40
	padR := 40
	padB := 30

	weeks := r.Data.Weeks
	if len(weeks) == 0 {
		weeks = []github.ContribWeek{}
	}

	numWeeks := len(weeks)
	if numWeeks == 0 {
		numWeeks = 1
	}

	gridW := W - padL - padR
	gridH := H - headerH - padB

	cellSize := gridW / numWeeks
	if cellSize > gridH/7 {
		cellSize = gridH / 7
	}
	if cellSize < 4 {
		cellSize = 4
	}
	gap := 2
	step := cellSize + gap

	cells := make([]cell, 0, numWeeks*7)
	for wi, week := range weeks {
		for di, day := range week.Days {
			color := day.Color
			if color == "" {
				color = C.Surface
			}
			x := padL + wi*step
			y := headerH + di*step
			cells = append(cells, cell{
				X: x, Y: y, S: cellSize,
				Color: color,
				Date:  day.Date,
				Count: day.Count,
			})
		}
	}

	return contribTmpl.Execute(w, contribRenderData{
		W: W, H: H, C: C,
		Login:  r.Data.Login,
		Total:  r.Data.TotalContributions,
		TitleY: 46, TitleSize: scaledFontSize(W, 28),
		SubtitleY: 76, SubSize: scaledFontSize(W, 18),
		Cells: cells,
	})
}
