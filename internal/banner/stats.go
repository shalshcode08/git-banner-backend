package banner

import (
	"html/template"
	"io"

	"github.com/somya/git-banner-backend/internal/github"
)

// StatsRenderer renders a GitHub stats card banner.
type StatsRenderer struct {
	Data   *github.StatsData
	Dims   Dimensions
	Colors Palette
}

var statsTmpl = template.Must(template.New("stats").Parse(`<svg xmlns="http://www.w3.org/2000/svg" width="{{.W}}" height="{{.H}}" viewBox="0 0 {{.W}} {{.H}}">
  <defs>
    <linearGradient id="bg" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" style="stop-color:{{.C.Background}};stop-opacity:1"/>
      <stop offset="100%" style="stop-color:{{.C.Surface}};stop-opacity:1"/>
    </linearGradient>
    <clipPath id="round"><rect width="{{.W}}" height="{{.H}}" rx="16" ry="16"/></clipPath>
  </defs>

  <!-- background -->
  <rect width="{{.W}}" height="{{.H}}" fill="url(#bg)" rx="16" ry="16"/>
  <rect width="{{.W}}" height="{{.H}}" fill="none" stroke="{{.C.Border}}" stroke-width="2" rx="16" ry="16"/>

  <!-- left accent bar -->
  <rect x="0" y="0" width="6" height="{{.H}}" fill="{{.C.Accent}}" rx="3" ry="3"/>

  <!-- avatar placeholder circle -->
  <circle cx="{{.AvatarX}}" cy="{{.AvatarY}}" r="{{.AvatarR}}" fill="{{.C.Surface}}" stroke="{{.C.Accent}}" stroke-width="3"/>
  <text x="{{.AvatarX}}" y="{{.AvatarYText}}" text-anchor="middle" font-size="{{.AvatarEmoji}}" font-family="serif">👤</text>

  <!-- name -->
  <text x="{{.NameX}}" y="{{.NameY}}" font-family="'Segoe UI',Helvetica,Arial,sans-serif" font-size="{{.NameSize}}" font-weight="700" fill="{{.C.Text}}">{{.Name}}</text>

  <!-- login -->
  <text x="{{.NameX}}" y="{{.LoginY}}" font-family="'Segoe UI',Helvetica,Arial,sans-serif" font-size="{{.LoginSize}}" fill="{{.C.Accent}}">@{{.Login}}</text>

  <!-- bio -->
  <text x="{{.NameX}}" y="{{.BioY}}" font-family="'Segoe UI',Helvetica,Arial,sans-serif" font-size="{{.BioSize}}" fill="{{.C.Subtext}}">{{.Bio}}</text>

  <!-- stat cards -->
  {{range .Stats}}
  <rect x="{{.X}}" y="{{.Y}}" width="{{.W}}" height="{{.H}}" rx="10" ry="10" fill="{{$.C.Surface}}" stroke="{{$.C.Border}}" stroke-width="1.5"/>
  <text x="{{.LabelX}}" y="{{.LabelY}}" text-anchor="middle" font-family="'Segoe UI',Helvetica,Arial,sans-serif" font-size="{{.LabelSize}}" fill="{{$.C.Subtext}}">{{.Label}}</text>
  <text x="{{.ValX}}" y="{{.ValY}}" text-anchor="middle" font-family="'Segoe UI',Helvetica,Arial,sans-serif" font-size="{{.ValSize}}" font-weight="700" fill="{{$.C.Accent}}">{{.Value}}</text>
  {{end}}
</svg>`))

type statCard struct {
	X, Y, W, H                 int
	LabelX, LabelY, ValX, ValY int
	LabelSize, ValSize         int
	Label, Value               string
}

type statsData struct {
	W, H                      int
	C                         Palette
	AvatarX, AvatarY, AvatarR int
	AvatarYText, AvatarEmoji  int
	NameX, NameY, NameSize    int
	LoginY, LoginSize         int
	BioY, BioSize             int
	Name, Login, Bio          string
	Stats                     []statCard
}

func (r *StatsRenderer) Render(w io.Writer) error {
	W, H := r.Dims.Width, r.Dims.Height
	C := r.Colors
	u := r.Data.User

	avatarR := H / 5
	avatarX := 80 + avatarR
	avatarY := H / 2

	nameX := avatarX*2 + 20
	nameY := H/2 - 60
	loginY := nameY + 36
	bioY := loginY + 28

	cardW := (W - nameX - 40) / 4
	cardH := 90
	cardY := H - cardH - 40
	cardGap := 20

	cards := []statCard{
		makeCard(nameX, cardY, cardW, cardH, cardGap, 0, "Repos", formatInt(u.PublicRepos)),
		makeCard(nameX, cardY, cardW, cardH, cardGap, 1, "Followers", formatInt(u.Followers)),
		makeCard(nameX, cardY, cardW, cardH, cardGap, 2, "Following", formatInt(u.Following)),
		makeCard(nameX, cardY, cardW, cardH, cardGap, 3, "Stars", formatInt(r.Data.TotalStars)),
	}

	name := u.Name
	if name == "" {
		name = u.Login
	}
	bio := u.Bio
	if len(bio) > 80 {
		bio = bio[:77] + "..."
	}

	fontSize := scaledFontSize(W, 36)

	return statsTmpl.Execute(w, statsData{
		W: W, H: H, C: C,
		AvatarX: avatarX, AvatarY: avatarY, AvatarR: avatarR,
		AvatarYText: avatarY + avatarR/3, AvatarEmoji: avatarR,
		NameX: nameX, NameY: nameY, NameSize: fontSize,
		LoginY: loginY, LoginSize: fontSize * 22 / 36,
		BioY: bioY, BioSize: fontSize * 18 / 36,
		Name: name, Login: u.Login, Bio: bio,
		Stats: cards,
	})
}

func makeCard(baseX, baseY, cW, cH, gap, idx int, label, value string) statCard {
	x := baseX + idx*(cW+gap)
	cx := x + cW/2
	return statCard{
		X: x, Y: baseY, W: cW, H: cH,
		LabelX: cx, LabelY: baseY + 28, ValX: cx, ValY: baseY + 66,
		LabelSize: 18, ValSize: 26,
		Label: label, Value: value,
	}
}
