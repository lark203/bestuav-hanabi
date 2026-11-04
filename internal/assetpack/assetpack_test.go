package assetpack

import (
	"archive/zip"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func makeZip(t *testing.T, entries map[string]string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "pack.zip")
	f, err := os.Create(p)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	for name, body := range entries {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write([]byte(body))
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	_ = f.Close()
	return p
}

func manifest(id string) string {
	data, _ := json.Marshal(map[string]any{"format": 1, "id": id, "name": id})
	return string(data)
}

func TestImport(t *testing.T) {
	zipPath := makeZip(t, map[string]string{
		"manifest.json":  manifest("ganyu"),
		"ganyu/bg.jpg":   "JPGDATA",
		"ganyu/icon.png": "PNGDATA",
	})
	dir := t.TempDir()
	res, err := Import(zipPath, dir, map[string]bool{"ganyu": true})
	if err != nil {
		t.Fatal(err)
	}
	if res.ID != "ganyu" || res.Files != 2 {
		t.Errorf("结果不对: %+v", res)
	}
	got, err := os.ReadFile(filepath.Join(dir, "ganyu", "bg.jpg"))
	if err != nil || string(got) != "JPGDATA" {
		t.Errorf("文件未写入: %v %q", err, got)
	}
	if _, err := os.Stat(filepath.Join(dir, "ganyu", "icon.png")); err != nil {
		t.Errorf("icon 缺失: %v", err)
	}
}

func TestImportRejects(t *testing.T) {
	dir := t.TempDir()
	known := map[string]bool{"ganyu": true}

	// 未知角色。
	if _, err := Import(makeZip(t, map[string]string{
		"manifest.json": manifest("nope"),
		"nope/bg.jpg":   "x",
	}), dir, known); err == nil {
		t.Error("未知 id 应拒绝")
	}
	// 路径穿越。
	if _, err := Import(makeZip(t, map[string]string{
		"manifest.json":  manifest("ganyu"),
		"ganyu/../x.png": "x",
	}), dir, known); err == nil {
		t.Error("路径穿越应拒绝")
	}
	// 非图片。
	if _, err := Import(makeZip(t, map[string]string{
		"manifest.json":    manifest("ganyu"),
		"ganyu/readme.txt": "x",
	}), dir, known); err == nil {
		t.Error("非图片应拒绝")
	}
	// 缺 manifest。
	if _, err := Import(makeZip(t, map[string]string{
		"ganyu/bg.jpg": "x",
	}), dir, known); err == nil {
		t.Error("缺 manifest 应拒绝")
	}
}
