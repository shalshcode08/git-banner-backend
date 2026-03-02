package banner

import "testing"

func TestDimsFor_Twitter(t *testing.T) {
	d := DimsFor(FormatTwitter)
	if d.Width != 1500 || d.Height != 500 {
		t.Errorf("Twitter dims: expected 1500x500, got %dx%d", d.Width, d.Height)
	}
}

func TestDimsFor_LinkedIn(t *testing.T) {
	d := DimsFor(FormatLinkedIn)
	if d.Width != 1584 || d.Height != 396 {
		t.Errorf("LinkedIn dims: expected 1584x396, got %dx%d", d.Width, d.Height)
	}
}

func TestDimsFor_Default(t *testing.T) {
	// unknown format falls back to Twitter
	d := DimsFor("unknown")
	if d.Width != 1500 || d.Height != 500 {
		t.Errorf("default dims: expected 1500x500, got %dx%d", d.Width, d.Height)
	}
}

func TestPaletteFor_Dark(t *testing.T) {
	p := PaletteFor(ThemeDark)
	if p.Background != "#0d1117" {
		t.Errorf("dark background: expected #0d1117, got %s", p.Background)
	}
	if p.Text != "#e6edf3" {
		t.Errorf("dark text: expected #e6edf3, got %s", p.Text)
	}
}

func TestPaletteFor_Light(t *testing.T) {
	p := PaletteFor(ThemeLight)
	if p.Background != "#ffffff" {
		t.Errorf("light background: expected #ffffff, got %s", p.Background)
	}
	if p.Text != "#1f2328" {
		t.Errorf("light text: expected #1f2328, got %s", p.Text)
	}
}

func TestPaletteFor_Default(t *testing.T) {
	// unknown theme falls back to dark
	p := PaletteFor("unknown")
	if p.Background != "#0d1117" {
		t.Errorf("default palette should be dark, got background %s", p.Background)
	}
}

func TestScaledFontSize_Normal(t *testing.T) {
	// base 36 at 1500px → 36
	got := scaledFontSize(1500, 36)
	if got != 36 {
		t.Errorf("expected 36, got %d", got)
	}
}

func TestScaledFontSize_LinkedIn(t *testing.T) {
	// 1584/1500 * 36 ≈ 38
	got := scaledFontSize(1584, 36)
	expected := 36 * 1584 / 1500
	if got != expected {
		t.Errorf("expected %d, got %d", expected, got)
	}
}

func TestScaledFontSize_MinClamp(t *testing.T) {
	// very small width should clamp to 10
	got := scaledFontSize(100, 10)
	if got < 10 {
		t.Errorf("expected at least 10, got %d", got)
	}
}

func TestFormatInt_Small(t *testing.T) {
	if s := formatInt(42); s != "42" {
		t.Errorf("expected %q, got %q", "42", s)
	}
	if s := formatInt(999); s != "999" {
		t.Errorf("expected %q, got %q", "999", s)
	}
}

func TestFormatInt_Thousands(t *testing.T) {
	if s := formatInt(1000); s != "1.0k" {
		t.Errorf("expected %q, got %q", "1.0k", s)
	}
	if s := formatInt(2500); s != "2.5k" {
		t.Errorf("expected %q, got %q", "2.5k", s)
	}
}
