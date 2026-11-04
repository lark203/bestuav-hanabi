package view

import (
	"errors"
	"path/filepath"

	"github.com/egoist/mygo"

	"bestuav-hanabi/internal/appmeta"
	"bestuav-hanabi/internal/profile"
	"bestuav-hanabi/internal/settings"
)

// 设置页的配置导出/导入：一个 JSON 包（设置 + 页面状态），不含素材。
// 文件对话框与读写都在 goroutine 里，结果回主线程更新提示。

func (a *App) exportProfile() {
	s := a.settings
	prefs := profile.Prefs(a.st)
	go func() {
		dir, _ := settings.DefaultPath()
		dir = filepath.Dir(dir)
		out, err := mygo.Dialog.Save(mygo.SaveDialogOptions{
			Title:       "导出配置",
			DefaultPath: filepath.Join(dir, "hanabi-profile.json"),
			ButtonLabel: "导出",
			Filters:     []mygo.FileFilter{{Name: "配置文件", Extensions: []string{"json"}}},
		})
		if err != nil || out == "" {
			return
		}
		if filepath.Ext(out) == "" {
			out += ".json"
		}
		err = profile.Export(out, appmeta.Identifier, appmeta.BuildVersion(), s, prefs)
		mygo.RunOnMain(func() {
			if err != nil {
				a.notice = "导出失败：" + err.Error()
			} else {
				a.notice = "已导出配置：" + filepath.Base(out)
			}
			a.updateNotice()
		})
	}()
}

func (a *App) importProfile() {
	go func() {
		paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{
			Title:       "导入配置",
			ButtonLabel: "导入",
			Filters:     []mygo.FileFilter{{Name: "配置文件", Extensions: []string{"json"}}},
		})
		if err != nil || len(paths) == 0 {
			return
		}
		b, err := profile.Import(paths[0])
		if err == nil && b.App != "" && b.App != appmeta.Identifier {
			err = errors.New("配置属于另一个应用（" + b.App + "）")
		}
		mygo.RunOnMain(func() {
			if err != nil {
				a.notice = "导入失败：" + err.Error()
				a.updateNotice()
				return
			}
			a.settings = b.Settings
			if a.st != nil && len(b.PagePrefs) > 0 {
				_ = a.st.ReplacePagePrefs(b.PagePrefs)
			}
			_ = settings.Save(a.settings)
			a.notice = "配置已导入，应用将重启以生效"
			a.updateNotice()
			_, _ = mygo.Dialog.Message(mygo.MessageOptions{
				Type:    mygo.MessageInfo,
				Title:   "导入完成",
				Message: "配置已导入，应用将重启以生效。",
			})
			mygo.App.Relaunch()
		})
	}()
}

// updateNotice 提示变化后请求重绘。
func (a *App) updateNotice() {
	if a.win != nil {
		a.win.Invalidate()
	}
}
