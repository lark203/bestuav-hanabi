package view

import (
	"github.com/egoist/mygo/ui"

	"bestuav-hanabi/internal/ui/icons"
)

// 页面注册表：侧栏、页面分发、默认页都从这里来。
//
// 新增页面：写页面函数（func (a *App) xxxPage(c *ui.Context)），再在 navRegistry
// 的 items 里加一条。生成器（scripts/newproject）也按这个文件裁剪。
type pageFunc func(a *App, c *ui.Context)

type navItem struct {
	id    string
	label string
	icon  *ui.SVG
	build pageFunc
}

const (
	libraryPageID   = "library"
	wallpaperPageID = "wallpaper"
	settingsPageID  = "settings"
)

// navRegistry 返回导航项（一个字面量，生成器按行裁剪：行尾 //gen:id=<id> 是裁剪标记）。
func navRegistry() []navItem {
	return []navItem{
		{id: libraryPageID, label: "主题库", icon: icons.Palette, build: (*App).libraryPage},      //gen:id=library
		{id: wallpaperPageID, label: "壁纸", icon: icons.Picture, build: (*App).placeholderPage}, //gen:id=wallpaper
		{id: newprojectPageID, label: "新建项目", icon: icons.Plus, build: (*App).newProjectPage},  //gen:id=newproject
		{id: settingsPageID, label: "设置", icon: icons.Sliders, build: (*App).settingsPage},     //gen:id=settings
	}
}

// pageByID 按 id 查页面。
func pageByID(id string) (navItem, bool) {
	for _, it := range navRegistry() {
		if it.id == id {
			return it, true
		}
	}
	return navItem{}, false
}

// firstPageID 默认页：注册表第一项（生成器删掉主题库后自动变成壁纸/设置）。
func firstPageID() string {
	items := navRegistry()
	if len(items) == 0 {
		return settingsPageID
	}
	return items[0].id
}
