package view

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"bestuav-hanabi/internal/assets"
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
			if a.genStatus != "" {
				ui.Text(c, a.genStatus).FontSize(a.fs(10)).
					TextColor(a.tok.TextMuted).Wrap()
			}

			field := func(label string, val *string) {
				ui.Field(c, label, func() {
					ui.TextInput(c, val)
				})
			}
			hint := func(s string) {
				ui.Text(c, s).FontSize(a.fs(9)).TextColor(a.tok.TextFaint).Wrap()
			}
			field("目标目录", &a.newProj.dir)
			hint("项目生成到哪里。目录必须不存在或为空；源码、go.mod、资源都会写到这里，生成后可直接 go run ./... 启动。")
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
			hint("Go module 路径（如 example.com/acme/app）。会改写 go.mod 与所有 internal/… 的 import，新项目应换成自己的域名。")
			field("产品名/slug", &a.newProj.name)
			hint("英文短名（如 my-app）。用于配置目录名、窗口状态键、可执行文件名与发布产物命名；不要带空格/中文。")
			field("窗口标题", &a.newProj.title)
			hint("窗口标题、关于页与诊断包里的展示名，可以中文。")
			field("短品牌名", &a.newProj.brand)
			hint("侧栏左上角的短品牌（2~6 个字符最合适），也用于标题栏；留空时默认同产品名。")
			field("应用标识", &a.newProj.id)
			hint("反向域名（如 com.acme.app）。单实例锁、更新源、配置目录与诊断包都用它区分应用，发布后不要随意改。")

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
			switch a.newProj.assets {
			case "none":
				hint("素材 none：不复制任何角色素材，界面走纯色降级，体积最小；之后可用「导入素材包」按需补图。")
			case "full":
				hint("素材 full：复制全部角色素材（约 20MB），适合直接发布完整版。")
			default:
				hint("素材 demo：只复制一个角色目录（下面的 demo 角色），体积小又能看到主题与壁纸效果。")
			}
			hint("demo 角色目录：assets/ 下要保留的角色目录名（如 ganyu）；只在 demo 模式生效，必须是模板里已有的角色。")
			field("保留页面（逗号分隔，留空=全部；settings 必留）", &a.newProj.pages)
			hint("侧栏保留哪些页面，例如 library,wallpaper,settings。settings 必须保留（设置入口）；留空表示全部保留。")

			ui.Row(c).Gap(14).Wrap().Children(func() {
				ui.Checkbox(c, &a.newProj.noIntro, "不要开场动画")
				ui.Checkbox(c, &a.newProj.noUpdate, "不要自动更新")
				ui.Checkbox(c, &a.newProj.skipBuild, "跳过编译自检")
			})
			hint("不要开场动画：删除 view 的 intro.go / opening.go 与相关测试，启动直接进主界面；适合不需要 6 秒开场的工具型项目。")
			hint("不要自动更新：删除 update.go，设置页显示「未启用」；没有更新源与签名密钥时建议勾选。")
			hint("跳过编译自检：生成后不跑 go build。速度快，但结果可能因环境问题不可编译，发布前请自己构建一次。")

			ui.Row(c).Gap(8).Children(func() {
				if ui.Button(c, "生成项目").Disabled(a.genBusy).Clicked() {
					a.generateProject()
				}
				if a.genBusy {
					ui.Text(c, "生成中（会跑一次 go build，请稍候）…").
						FontSize(a.fs(11)).TextColor(a.tok.Accent)
				}
			})
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
		a.updateNotice()
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
	a.updateNotice()
	go func() {
		res, err := scaffold.Generate(scaffold.Options{
			Root:       "", // 用内嵌模板，不依赖源码树
			AssetRoot:  assets.Dir(),
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
			} else {
				a.genStatus = fmt.Sprintf("已生成 %s（%d 个文件；页面 %s；素材 %s）。",
					res.Dir, res.Files, strings.Join(res.Pages, ", "), res.Assets)
			}
			a.updateNotice()
		})
		// 弹窗确认：生成在后台跑，用户可能已经滚到别处；只有弹窗能保证看到结果。
		if err != nil {
			_, _ = mygo.Dialog.Message(mygo.MessageOptions{
				Type:    mygo.MessageError,
				Title:   "生成失败",
				Message: "项目没有生成完整。",
				Detail:  err.Error(),
			})
			return
		}
		detail := fmt.Sprintf("目录：%s\n文件：%d\n页面：%s\n素材：%s",
			res.Dir, res.Files, strings.Join(res.Pages, ", "), res.Assets)
		if strings.TrimSpace(res.BuildOutput) != "" {
			out := res.BuildOutput
			if len(out) > 4000 {
				out = out[:4000] + "…"
			}
			detail += "\n\n" + out
		}
		_, _ = mygo.Dialog.Message(mygo.MessageOptions{
			Title:   "生成完成",
			Message: "项目已生成：" + res.Dir,
			Detail:  detail,
		})
	}()
}
