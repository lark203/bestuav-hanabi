package view

import (
	"path/filepath"

	"github.com/egoist/mygo"

	"bestuav-hanabi/internal/apppath"
	"bestuav-hanabi/internal/assets"
	"bestuav-hanabi/internal/themepack"
	"bestuav-hanabi/internal/visual"
)

// 导入外部角色主题包：新增一个主题（不是只补素材），立即注册到视觉层。
func (a *App) importThemePack() {
	go func() {
		paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{
			Title:       "导入角色主题包",
			ButtonLabel: "导入",
			Filters:     []mygo.FileFilter{{Name: "主题包", Extensions: []string{"zip"}}},
		})
		if err != nil || len(paths) == 0 {
			return
		}
		themesDir, terr := apppath.ThemesDir()
		if terr != nil {
			return
		}
		ch, ierr := themepack.Install(paths[0], themesDir, assets.UserDir(), visual.BuiltinIDs())
		if ierr == nil {
			visual.RegisterExternal([]visual.Character{ch})
			assets.Clear()
		}
		mygo.RunOnMain(func() {
			if ierr != nil {
				a.notice = "主题包导入失败：" + ierr.Error()
			} else {
				a.notice = "已导入主题「" + ch.NameCN + "」（" + filepath.Base(paths[0]) + "），可在主题库切换"
			}
			a.updateNotice()
		})
	}()
}
