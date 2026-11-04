package view

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// 页面滚动体必须走 pageScroll：它会统一给右侧滚动条留空。
// 直接写 ui.Scroll 的页面会重新出现“输入框/卡片贴着滚动条”的问题。
func TestPagesUsePageScroll(t *testing.T) {
	dir := "."
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") ||
			strings.HasSuffix(name, "_test.go") || name == "page_layout.go" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(data), "ui.Scroll(") {
			t.Errorf("%s 直接用了 ui.Scroll；请改用 pageScroll（统一留出滚动条空档）", name)
		}
	}
}
