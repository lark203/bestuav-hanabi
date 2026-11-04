package view

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"bestuav-hanabi/internal/scaffold"
)

const newprojectPageID = "newproject"

// newProjectPage 用图形界面生成一个独立项目（与命令行脚本同一套 scaffold）。
//
// 只有模板仓库里有这页（生成器会把本文件与导航条目裁掉），生成的 App 不需要
// scaffold 包。
func (a *App) newProjectPage(c *ui.Context) {
	ui.Scroll(c).Grow(1).Children(func() {
		ui.Column(c).Padding(4, 6, 10, 0).Gap(10).Children(func() {
			ui.Text(c, "新建项目").FontSize(a.fs(22)).Bold().TextColor(a.tok.Text)
			ui.Text(c, "从当前模板生成一个可编译的独立项目；等价于 go run ./scripts/newproject。").
				FontSize(a.fs(11)).TextColor(a.tok.TextFaint)

			field := func(label string, val *string) {
				ui.Field(c, label, func() {
					ui.TextInput(c, val)
				})
			}
			field("目标目录", &a.newProj.dir)
			ui.Row(c).Gap(8).Children(func() {
				if ui.Button(c, "选择目录…").Clicked() {
					a.chooseProjectDir()
				}
				if ui.Button(c, "使用默认目录").Clicked() {
					if home, err := mygo.App.Path(mygo.PathUserData); err == nil {
						a.newProj.dir = filepath.Join(home, "新项目")
					}
				}
			})
			field("模块路径", &a.newProj.module)
			field("产品名/slug", &a.newProj.name)
			field("窗口标题", &a.newProj.title)
			field("短品牌名", &a.newProj.brand)
			field("应用标识", &a.newProj.id)

			ui.Row(c).Gap(12).Children(func() {
				ui.Column(c).Gap(4).Grow(1).Children(func() {
					ui.Text(c, "素材").FontSize(a.fs(11)).TextColor(a.tok.TextMuted)
					ui.Select(c, &a.newProj.assets, []string{"none", "demo", "full"})
				})
				ui.Column(c).Gap(4).Grow(1).Children(func() {
					ui.Text(c, "demo 角色目录").FontSize(a.fs(11)).TextColor(a.tok.TextMuted)
					ui.TextInput(c, &a.newProj.demo)
				})
			})
			field("保留页面（逗号分隔，留空=全部；settings 必留）", &a.newProj.pages)

			ui.Row(c).Gap(14).Wrap().Children(func() {
				ui.Checkbox(c, &a.newProj.noIntro, "不要开场动画")
				ui.Checkbox(c, &a.newProj.noUpdate, "不要自动更新")
				ui.Checkbox(c, &a.newProj.skipBuild, "跳过编译自检")
			})

			ui.Row(c).Gap(8).Children(func() {
				if ui.Button(c, "生成项目").Disabled(a.genBusy).Clicked() {
					a.generateProject()
				}
				if a.genBusy {
					ui.Text(c, "生成中（会跑一次 go build，请稍候）…").
						FontSize(a.fs(11)).TextColor(a.tok.Accent)
				}
			})
			if a.genStatus != "" {
				ui.Text(c, a.genStatus).FontSize(a.fs(10)).
					TextColor(a.tok.TextMuted).Wrap()
			}
			ui.Text(c, "生成目录必须不存在或为空；生成后可直接 go run ./... 启动。").
				FontSize(a.fs(10)).TextColor(a.tok.TextFaint)
		})
	})
}

func (a *App) chooseProjectDir() {
	go func() {
		dirs, err := mygo.Dialog.Open(mygo.OpenDialogOptions{
			Title:             "选择生成目录",
			ButtonLabel:       "选择",
			Directory:         true,
			CreateDirectories: true,
		})
		if err != nil || len(dirs) == 0 {
			return
		}
		mygo.RunOnMain(func() { a.newProj.dir = dirs[0] })
	}()
}

func (a *App) generateProject() {
	f := a.newProj
	if strings.TrimSpace(f.dir) == "" {
		a.genStatus = "请先选择目标目录。"
		return
	}
	var keep []string
	for _, p := range strings.Split(f.pages, ",") {
		if p = strings.TrimSpace(p); p != "" {
			keep = append(keep, p)
		}
	}
	a.genBusy = true
	a.genStatus = "正在生成 " + f.dir + " …"
	go func() {
		res, err := scaffold.Generate(scaffold.Options{
			Root:       "", // 用内嵌模板，不依赖源码树
			Dir:        f.dir,
			Module:     strings.TrimSpace(f.module),
			Name:       strings.TrimSpace(f.name),
			Title:      strings.TrimSpace(f.title),
			Brand:      strings.TrimSpace(f.brand),
			Identifier: strings.TrimSpace(f.id),
			Assets:     f.assets,
			DemoAsset:  strings.TrimSpace(f.demo),
			KeepPages:  keep,
			NoIntro:    f.noIntro,
			NoUpdate:   f.noUpdate,
			SkipBuild:  f.skipBuild,
		})
		mygo.RunOnMain(func() {
			a.genBusy = false
			if err != nil {
				a.genStatus = "生成失败：" + err.Error()
				return
			}
			msg := fmt.Sprintf("已生成 %s（%d 个文件；页面 %s；素材 %s）。",
				res.Dir, res.Files, strings.Join(res.Pages, ", "), res.Assets)
			if res.BuildOutput != "" {
				msg += "\n" + res.BuildOutput
			}
			a.genStatus = msg
		})
	}()
}
