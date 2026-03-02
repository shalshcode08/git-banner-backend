package banner

import (
	"html/template"
	"io"

	"github.com/somya/git-banner-backend/internal/github"
)

// PinnedRenderer renders a pinned repos banner.
type PinnedRenderer struct {
	Data   *github.PinnedData
	Dims   Dimensions
	Colors Palette
}

var pinnedTmpl = template.Must(template.New("pinned").Parse(`<svg xmlns="http://www.w3.org/2000/svg" width="{{.W}}" height="{{.H}}" viewBox="0 0 {{.W}} {{.H}}">
  <defs>
    <linearGradient id="bg" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" style="stop-color:{{.C.Background}};stop-opacity:1"/>
      <stop offset="100%" style="stop-color:{{.C.Surface}};stop-opacity:1"/>
    </linearGradient>
  </defs>

  <rect width="{{.W}}" height="{{.H}}" fill="url(#bg)" rx="16" ry="16"/>
  <rect width="{{.W}}" height="{{.H}}" fill="none" stroke="{{.C.Border}}" stroke-width="2" rx="16" ry="16"/>
  <rect x="0" y="0" width="6" height="{{.H}}" fill="{{.C.Accent}}" rx="3" ry="3"/>

  <!-- header -->
  <text x="40" y="{{.HeaderY}}" font-family="'Segoe UI',Helvetica,Arial,sans-serif" font-size="{{.HeaderSize}}" font-weight="700" fill="{{.C.Text}}">{{.Login}}'s Pinned Repos</text>

  <!-- repo cards -->
  {{range .Cards}}
  <rect x="{{.X}}" y="{{.Y}}" width="{{.W}}" height="{{.H}}" rx="10" ry="10" fill="{{$.C.Surface}}" stroke="{{$.C.Border}}" stroke-width="1.5"/>
  <!-- repo name -->
  <text x="{{.TX}}" y="{{.NameY}}" font-family="'Segoe UI',Helvetica,Arial,sans-serif" font-size="{{.NameSize}}" font-weight="600" fill="{{$.C.Accent}}">{{.Name}}</text>
  <!-- description -->
  <text x="{{.TX}}" y="{{.DescY}}" font-family="'Segoe UI',Helvetica,Arial,sans-serif" font-size="{{.DescSize}}" fill="{{$.C.Subtext}}">{{.Desc}}</text>
  <!-- language dot -->
  {{if .Lang}}<circle cx="{{.LangDotX}}" cy="{{.LangDotY}}" r="6" fill="{{$.C.Green}}"/>
  <text x="{{.LangX}}" y="{{.LangTextY}}" font-family="'Segoe UI',Helvetica,Arial,sans-serif" font-size="{{.MetaSize}}" fill="{{$.C.Subtext}}">{{.Lang}}</text>{{end}}
  <!-- stars -->
  <text x="{{.StarsX}}" y="{{.LangTextY}}" font-family="'Segoe UI',Helvetica,Arial,sans-serif" font-size="{{.MetaSize}}" fill="{{$.C.Subtext}}">★ {{.Stars}}</text>
  <!-- forks -->
  <text x="{{.ForksX}}" y="{{.LangTextY}}" font-family="'Segoe UI',Helvetica,Arial,sans-serif" font-size="{{.MetaSize}}" fill="{{$.C.Subtext}}">⑂ {{.Forks}}</text>
  {{end}}
</svg>`))

type repoCard struct {
	X, Y, W, H                        int
	TX, NameY, NameSize               int
	DescY, DescSize                   int
	LangDotX, LangDotY                int
	LangX, LangTextY, LangTextYOffset int
	StarsX, ForksX, MetaSize          int
	Name, Desc, Lang                  string
	Stars, Forks                      string
}

type pinnedData struct {
	W, H       int
	C          Palette
	Login      string
	HeaderY    int
	HeaderSize int
	Cards      []repoCard
}

func (r *PinnedRenderer) Render(w io.Writer) error {
	W, H := r.Dims.Width, r.Dims.Height
	C := r.Colors

	repos := r.Data.Repos
	if len(repos) > 6 {
		repos = repos[:6]
	}

	cols := 3
	if len(repos) <= 3 {
		cols = len(repos)
	}
	if cols == 0 {
		cols = 1
	}
	rows := (len(repos) + cols - 1) / cols

	headerH := 70
	pad := 30
	gap := 16
	cardW := (W - pad*2 - gap*(cols-1)) / cols
	cardH := (H - headerH - pad*2 - gap*(rows-1)) / rows

	cards := make([]repoCard, 0, len(repos))
	for i, repo := range repos {
		col := i % cols
		row := i / cols
		x := pad + col*(cardW+gap)
		y := headerH + pad + row*(cardH+gap)
		tx := x + 14
		nameY := y + 28
		descY := nameY + 22
		langDotY := y + cardH - 18
		langDotX := tx + 6
		langX := langDotX + 16
		langTextY := langDotY + 5
		starsX := langX + 100
		forksX := starsX + 80

		desc := repo.Description
		maxDescLen := cardW / 8
		if len(desc) > maxDescLen {
			desc = desc[:maxDescLen-3] + "..."
		}

		cards = append(cards, repoCard{
			X: x, Y: y, W: cardW, H: cardH,
			TX:    tx,
			NameY: nameY, NameSize: scaledFontSize(W, 18),
			DescY: descY, DescSize: scaledFontSize(W, 14),
			LangDotX: langDotX, LangDotY: langDotY,
			LangX: langX, LangTextY: langTextY,
			StarsX: starsX, ForksX: forksX,
			MetaSize: scaledFontSize(W, 13),
			Name:     repo.Name,
			Desc:     desc,
			Lang:     repo.Language,
			Stars:    formatInt(repo.Stars),
			Forks:    formatInt(repo.Forks),
		})
	}

	return pinnedTmpl.Execute(w, pinnedData{
		W: W, H: H, C: C,
		Login:      r.Data.User.Login,
		HeaderY:    46,
		HeaderSize: scaledFontSize(W, 28),
		Cards:      cards,
	})
}
