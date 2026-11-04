package view

import (
	"fmt"
	"path/filepath"

	"github.com/egoist/mygo"

	"bestuav-hanabi/internal/apppath"
	"bestuav-hanabi/internal/assets"
	"bestuav-hanabi/internal/themepack"
	"bestuav-hanabi/internal/visual"
)

// 重新导入指定 id 的主题包：必须是同一个角色，否则拒绝（避免误覆盖别的主题）。
func (a *App) reimportThemePack(id string) {
	go func() {
		paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{
			Title:       "重新导入主题「" + id + "」",
			ButtonLabel: "导入",
			Filters:     []mygo.FileFilter{{Name: "主题包", Extensions: []string{"zip"}}},
		})
		if err != nil || len(paths) == 0 {
			return
		}
		pack, perr := themepack.Inspect(paths[0])
		if perr == nil && pack.ID != id {
			perr = fmt.Errorf("这个包是「%s」，不是要更新的「%s」", pack.ID, id)
		}
		var ch visual.Character
		if perr == nil {
			themesDir, terr := apppath.ThemesDir()
			if terr != nil {
				perr = terr
			} else {
				ch, perr = themepack.Install(paths[0], themesDir, assets.UserDir(), visual.BuiltinIDs())
			}
		}
		if perr == nil {
			visual.RegisterExternal([]visual.Character{ch})
			assets.Clear()
		}
		mygo.RunOnMain(func() {
			if perr != nil {
				a.notice = "重新导入失败：" + perr.Error()
			} else {
				a.notice = "已更新主题「" + ch.NameCN + "」"
			}
			a.updateNotice()
		})
	}()
}

// 删除外部主题包：确认后移除 theme.json 与 assets/<id>，并从视觉层注销。
func (a *App) removeThemePack(id string) {
	go func() {
		res, err := mygo.Dialog.Message(mygo.MessageOptions{
			Title:              "删除主题包",
			Message:            "确定删除主题「" + id + "」？",
			Detail:             "会同时删除用户目录里的主题配置与该角色的导入素材；内建主题不受影响。",
			Buttons:            []string{"删除", "取消"},
			DestructiveButtons: []int{0},
		})
		if err != nil || res.Button != 0 {
			return
		}
		themesDir, terr := apppath.ThemesDir()
		if terr != nil {
			return
		}
		rerr := themepack.Remove(themesDir, assets.UserDir(), id)
		if rerr == nil {
			visual.RemoveExternal(id)
			assets.Clear()
		}
		mygo.RunOnMain(func() {
			if rerr != nil {
				a.notice = "删除主题包失败：" + rerr.Error()
			} else {
				if a.char.ID == id {
					a.setCharacter(visual.Default().ID)
					a.finishSwitch()
				}
				a.notice = "已删除主题「" + id + "」"
			}
			a.updateNotice()
		})
	}()
}

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
