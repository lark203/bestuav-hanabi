package view

import (
	"archive/zip"
	"path/filepath"
	"testing"

	"bestuav-hanabi/internal/store"
)

func TestDiagnosticsZip(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "data.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if err := st.SetPagePref("library", "filter", "1"); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "diag.zip")
	if err := writeDiagnostics(out, st); err != nil {
		t.Fatal(err)
	}
	zr, err := zip.OpenReader(out)
	if err != nil {
		t.Fatal(err)
	}
	defer zr.Close()
	want := map[string]bool{"env.txt": false, "data.db": false}
	for _, f := range zr.File {
		if _, ok := want[f.Name]; ok {
			want[f.Name] = true
		}
	}
	for name, ok := range want {
		if !ok {
			t.Errorf("诊断包缺少 %s", name)
		}
	}
}
