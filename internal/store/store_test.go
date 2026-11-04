package store

import (
	"path/filepath"
	"testing"
)

func TestPagePrefRoundTrip(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "data.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	if _, ok, err := s.PagePref("library", "filter"); err != nil || ok {
		t.Fatalf("不存在的键应 ok=false: %v %v", ok, err)
	}
	if err := s.SetPagePref("library", "filter", `3`); err != nil {
		t.Fatal(err)
	}
	if v, ok, err := s.PagePref("library", "filter"); err != nil || !ok || v != "3" {
		t.Fatalf("读回失败: %q %v %v", v, ok, err)
	}
	// 覆盖写。
	if err := s.SetPagePref("library", "filter", `-1`); err != nil {
		t.Fatal(err)
	}
	if v, _, _ := s.PagePref("library", "filter"); v != "-1" {
		t.Errorf("覆盖后 = %q", v)
	}
	if n, err := s.PagePrefCount(); err != nil || n != 1 {
		t.Errorf("count = %d, %v", n, err)
	}
	if err := s.ClearPagePrefs(); err != nil {
		t.Fatal(err)
	}
	if n, _ := s.PagePrefCount(); n != 0 {
		t.Errorf("清理后 count = %d", n)
	}
}

func TestOpenEmptyPath(t *testing.T) {
	if _, err := Open(""); err == nil {
		t.Error("空路径应报错")
	}
}
