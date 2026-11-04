package view

import (
	"path/filepath"
	"testing"

	"bestuav-hanabi/internal/store"
)

func TestRegistryConsistent(t *testing.T) {
	items := navRegistry()
	if len(items) == 0 {
		t.Fatal("注册表为空")
	}
	seen := map[string]bool{}
	for _, it := range items {
		if it.id == "" || it.label == "" || it.icon == nil || it.build == nil {
			t.Errorf("条目不完整: %+v", it)
		}
		if seen[it.id] {
			t.Errorf("重复 id: %s", it.id)
		}
		seen[it.id] = true
	}
	if firstPageID() != libraryPageID {
		t.Errorf("默认页 = %q", firstPageID())
	}
	if _, ok := pageByID(settingsPageID); !ok {
		t.Error("设置页必须存在")
	}
	if _, ok := pageByID("nope"); ok {
		t.Error("未知页面不应命中")
	}
}

func TestSetPage(t *testing.T) {
	a := NewApp()
	a.SetPage(settingsPageID)
	if a.page != settingsPageID {
		t.Errorf("SetPage 失败: %q", a.page)
	}
	a.SetPage("nope")
	if a.page != settingsPageID {
		t.Error("未知页面不应改变当前页")
	}
}

func TestPagePrefsPersist(t *testing.T) {
	st, err := store.Open(filepath.Join(t.TempDir(), "data.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	a := NewApp()
	a.AttachStore(st)
	a.prefSet(libraryPageID, "filter", 3)
	a.prefSet(libraryPageID, "view", 1)

	b := NewApp()
	if b.filter != -1 || b.themeView != 0 {
		t.Fatalf("默认值不对: filter=%d view=%d", b.filter, b.themeView)
	}
	b.AttachStore(st)
	if b.filter != 3 || b.themeView != 1 {
		t.Errorf("页面状态未恢复: filter=%d view=%d", b.filter, b.themeView)
	}
}

func TestNoStoreIsMemoryOnly(t *testing.T) {
	a := NewApp()
	a.AttachStore(nil)
	a.prefSet(libraryPageID, "filter", 2) // 不应 panic
}
