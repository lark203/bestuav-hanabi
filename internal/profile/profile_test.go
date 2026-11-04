package profile

import (
	"os"
	"path/filepath"
	"testing"

	"bestuav-hanabi/internal/settings"
)

func TestRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "profile.json")
	s := settings.Default()
	s.Theme = "nahida"
	s.FontPercent = 108
	s.IntroPlay = false
	prefs := map[string]map[string]string{
		"library": {"filter": "3", "view": "1"},
	}
	if err := Export(path, "com.bestuav.hanabi", "1.2.3", s, prefs); err != nil {
		t.Fatal(err)
	}
	b, err := Import(path)
	if err != nil {
		t.Fatal(err)
	}
	if b.App != "com.bestuav.hanabi" || b.Version != "1.2.3" {
		t.Errorf("元信息不对: %+v", b)
	}
	if b.Settings.Theme != "nahida" || b.Settings.FontPercent != 108 || b.Settings.IntroPlay {
		t.Errorf("设置未保留: %+v", b.Settings)
	}
	if b.PagePrefs["library"]["filter"] != "3" || b.PagePrefs["library"]["view"] != "1" {
		t.Errorf("页面状态未保留: %+v", b.PagePrefs)
	}
}

func TestImportRejectsBadFormat(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(path, []byte(`{"format":99}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Import(path); err == nil {
		t.Error("未来格式应拒绝")
	}
	if err := os.WriteFile(path, []byte("not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Import(path); err == nil {
		t.Error("坏 JSON 应拒绝")
	}
}
