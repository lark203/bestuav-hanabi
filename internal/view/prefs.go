package view

import (
	"encoding/json"
	"time"

	"bestuav-hanabi/internal/store"
)

// 页面私有状态（page_pref）：筛选条件、上次选的视图这类「记住就行、丢了也无所谓」
// 的东西。应用级偏好仍在 settings.json。
//
// 没有数据库时（测试、--safe-mode）退化为内存：功能照常，只是不跨启动保留。

// AttachStore 注入数据库并恢复页面状态（main 在创建窗口前调用）。
func (a *App) AttachStore(st *store.Store) {
	a.st = st
	if st == nil {
		return
	}
	if v, ok, err := st.PagePref(libraryPageID, "filter"); err == nil && ok {
		var f int
		if json.Unmarshal([]byte(v), &f) == nil {
			a.filter = f
		}
	}
	if v, ok, err := st.PagePref(libraryPageID, "view"); err == nil && ok {
		var view int
		if json.Unmarshal([]byte(v), &view) == nil && (view == 0 || view == 1) {
			a.themeView = view
		}
	}
}

// prefSet 写一条页面状态；失败只忽略（页面状态不影响主流程）。
func (a *App) prefSet(page, key string, v any) {
	if a.st == nil {
		return
	}
	data, err := json.Marshal(v)
	if err != nil {
		return
	}
	_ = a.st.SetPagePref(page, key, string(data))
}

// SetSafeMode 安全模式：跳过自动更新检查（数据库由 main 决定是否打开）。
func (a *App) SetSafeMode(on bool) { a.safeMode = on }

// SetPage 按 id 切页（--page / 单实例转发用）；未知 id 忽略。
func (a *App) SetPage(id string) {
	if _, ok := pageByID(id); !ok {
		return
	}
	a.page = id
	a.pageAt = time.Now().Add(-time.Hour) // 启动/转发不做入场动画
	a.viewAt = time.Now().Add(-time.Hour)
}

// Page 当前页面 id（测试/诊断用）。
func (a *App) Page() string { return a.page }
