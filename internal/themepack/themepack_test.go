package themepack

import (
	"archive/zip"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func samplePack() Pack {
	return Pack{
		Format: 1, ID: "demo", NameCN: "演示", NameEN: "Demo",
		Epithet: "试作", Motto: "测试一下", Constellation: "D E M O",
		Scene: "测试 · 场景", SceneEN: "A TEST",
		Element: "cryo", Motif: "crystal", Opening: "ganyu",
		Chrome: "plain", Ornament: "none", Art: "blossom",
		CardBg: "#203040", Bg: "demo/bg.jpg", Icon: "demo/icon.png",
		Tokens: TokensPack{
			WinFill: "#101820", WinStroke: "#203040", Sidebar: "#0c1218",
			Card: "#182430", CardStroke: "#304050", Text: "#e8f0f8",
			TextMuted: "#a0b0c0", TextFaint: "#708090", Accent: "#80c0ff",
			AccentSoft: "#20304080", Good: "#80ff80", Danger: "#ff6080",
			Radius: 12,
		},
		Motion: MotionPack{Person: "demo/person.png", Breath: 0.002, Hair: "demo/hair.png", Sway: 0.001},
	}
}

func makeZip(t *testing.T, p Pack, extra map[string]string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "theme.zip")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	data, _ := json.Marshal(p)
	w, _ := zw.Create("theme.json")
	_, _ = w.Write(data)
	for name, body := range extra {
		w, _ := zw.Create(name)
		_, _ = w.Write([]byte(body))
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()
	return path
}

func TestInstallAndLoad(t *testing.T) {
	zipPath := makeZip(t, samplePack(), map[string]string{
		"demo/bg.jpg":     "JPG",
		"demo/icon.png":   "PNG",
		"demo/person.png": "MASK",
	})
	themes, assets := t.TempDir(), t.TempDir()
	ch, err := Install(zipPath, themes, assets, map[string]bool{"ganyu": true})
	if err != nil {
		t.Fatal(err)
	}
	if ch.ID != "demo" || ch.NameCN != "演示" || ch.Tok.Radius != 12 || ch.WallTint == 0 || ch.Scrim == 0 {
		t.Errorf("转换结果不对: %+v", ch)
	}
	if _, err := os.Stat(filepath.Join(themes, "demo.json")); err != nil {
		t.Error("theme.json 未保存:", err)
	}
	if _, err := os.Stat(filepath.Join(assets, "demo", "bg.jpg")); err != nil {
		t.Error("bg 未复制:", err)
	}
	loaded, errs := LoadDir(themes)
	if len(errs) != 0 || len(loaded) != 1 {
		t.Fatalf("LoadDir: %d 个, errs=%v", len(loaded), errs)
	}
	if loaded[0].Motto != "测试一下" || loaded[0].Bg != "demo/bg.jpg" {
		t.Errorf("加载结果不对: %+v", loaded[0])
	}
}

func TestInstallRejects(t *testing.T) {
	themes, assets := t.TempDir(), t.TempDir()
	known := map[string]bool{"ganyu": true}

	cases := []struct {
		name  string
		pack  Pack
		extra map[string]string
	}{
		{"未知元素", func() Pack { p := samplePack(); p.Element = "light"; return p }(), map[string]string{"demo/bg.jpg": "x"}},
		{"缺少 bg", samplePack(), map[string]string{"demo/icon.png": "x"}},
		{"路径穿越", samplePack(), map[string]string{"demo/bg.jpg": "x", "demo/../evil.png": "x"}},
		{"内建冲突", func() Pack { p := samplePack(); p.ID = "ganyu"; p.Bg = "ganyu/bg.jpg"; return p }(), map[string]string{"ganyu/bg.jpg": "x"}},
	}
	for _, tc := range cases {
		zipPath := makeZip(t, tc.pack, tc.extra)
		if _, err := Install(zipPath, themes, assets, known); err == nil {
			t.Errorf("%s: 应拒绝", tc.name)
		}
	}
}

func TestLoadDirSkipsBad(t *testing.T) {
	dir := t.TempDir()
	bad := samplePack()
	bad.Element = "nope"
	data, _ := json.Marshal(bad)
	if err := os.WriteFile(filepath.Join(dir, "bad.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "broken.json"), []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	loaded, errs := LoadDir(dir)
	if len(loaded) != 0 || len(errs) != 2 {
		t.Errorf("坏包应跳过: loaded=%d errs=%d", len(loaded), len(errs))
	}
}
