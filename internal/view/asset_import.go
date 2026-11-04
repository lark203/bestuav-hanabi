package view

import (
	"path/filepath"

	"github.com/egoist/mygo"

	"bestuav-hanabi/internal/assetpack"
	"bestuav-hanabi/internal/assets"
	"bestuav-hanabi/internal/visual"
)

// 导入角色素材包：只补已有角色的图片（bg/icon/遮罩），不新增角色。
// 写入用户素材目录（安装目录通常只读），清空解码缓存后下一帧生效。
func (a *App) importAssetPack() {
	go func() {
		paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{
			Title:       "导入角色素材包",
			ButtonLabel: "导入",
			Filters:     []mygo.FileFilter{{Name: "素材包", Extensions: []string{"zip"}}},
		})
		if err != nil || len(paths) == 0 {
			return
		}
		res, ierr := assetpack.Import(paths[0], assets.UserDir(), knownThemeIDs())
		if ierr == nil {
			assets.Clear() // 下一帧重新解码（替换同角色素材时尤其需要）
		}
		mygo.RunOnMain(func() {
			if ierr != nil {
				a.notice = "素材包导入失败：" + ierr.Error()
			} else {
				a.notice = "已导入 " + res.ID + " 的素材（" + filepath.Base(paths[0]) + "）"
			}
			a.updateNotice()
		})
	}()
}

// knownThemeIDs 返回主题表里的角色 id（素材包只能补这些角色）。
func knownThemeIDs() map[string]bool {
	out := map[string]bool{}
	for _, c := range visual.All() {
		out[c.ID] = true
	}
	return out
}
