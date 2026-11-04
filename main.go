// 花火 HanaBi：以角色主题为核心的桌面 UI。
//
// 窗口隐藏系统标题栏与窗口控件，由界面自绘（见 internal/view/app.go 的 titlebar）。
//
// 这里只做启动编排：参数 → 单实例锁 → 重置偏好（可选）→ 数据库 → App → 窗口。
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"bestuav-hanabi/internal/appmeta"
	"bestuav-hanabi/internal/apppath"
	"bestuav-hanabi/internal/assets"
	"bestuav-hanabi/internal/cli"
	"bestuav-hanabi/internal/settings"
	"bestuav-hanabi/internal/store"
	"bestuav-hanabi/internal/themepack"
	"bestuav-hanabi/internal/view"
	"bestuav-hanabi/internal/visual"
)

func main() {
	// 壁纸解码/缩放是几十 MB 的短命分配：默认 GOGC=100 会先翻倍再回收。
	debug.SetGCPercent(50)
	debug.SetMemoryLimit(384 << 20)

	opts, err := cli.Parse(os.Args[1:], os.Stderr)
	if err != nil {
		if err == flag.ErrHelp {
			return
		}
		fmt.Fprintln(os.Stderr, "参数错误:", err)
		os.Exit(2)
	}
	if opts.Version {
		cli.AttachConsole()
		fmt.Printf("%s %s (%s/%s)\n", appmeta.Title, appmeta.BuildVersion(), runtime.GOOS, runtime.GOARCH)
		return
	}

	var (
		app *view.App
		win *mygo.Window
	)

	// 单实例：重复启动只唤出已有窗口（并可带 --page 跳页）。
	if !mygo.App.RequestSingleInstanceLock() {
		cli.AttachConsole()
		fmt.Fprintln(os.Stderr, "已有一个实例在运行。")
		return
	}
	mygo.App.OnSecondInstance(func(args []string, _ string) {
		page := cli.PageFromArgs(args)
		mygo.RunOnMain(func() {
			if app == nil {
				return
			}
			if page != "" {
				app.SetPage(page)
			}
			if win != nil {
				win.Show()
				win.Focus()
			}
		})
	})

	if opts.ResetSettings {
		if backup, err := settings.Reset(); err != nil {
			cli.AttachConsole()
			fmt.Fprintln(os.Stderr, "重置偏好失败:", err)
		} else if backup != "" {
			cli.AttachConsole()
			fmt.Fprintln(os.Stderr, "已重置偏好，旧文件备份在:", backup)
		}
	}

	// 数据库：失败不影响启动（页面状态退化为内存）。
	var db *store.Store
	if !opts.SafeMode {
		if path, err := apppath.DBPath(); err == nil {
			if s, err := store.Open(path); err == nil {
				db = s
				defer func() { _ = db.Close() }()
			}
		}
	}

	mygo.App.WhenReady(func() {
		// 用户导入的素材写在这里（安装目录可能只读，macOS .app 尤其如此）。
		if d, err := apppath.Dir(); err == nil {
			assets.SetUserDir(filepath.Join(d, "assets"))
			// 外部主题包：先注册再创建 App，主题库与启动主题都能看到。
			if themesDir, err := apppath.ThemesDir(); err == nil {
				chs, errs := themepack.LoadDir(themesDir)
				for _, e := range errs {
					fmt.Fprintln(os.Stderr, "外部主题加载失败:", e)
				}
				visual.RegisterExternal(chs)
			}
		}
		app = view.NewAppWithSettings(settings.Load())
		app.AttachStore(db)
		if opts.SafeMode {
			app.SetSafeMode(true)
		}
		if opts.Page != "" {
			app.SetPage(opts.Page)
		}
		mygo.App.OnQuit(app.FlushSettings)

		win = mygo.NewWindow(mygo.WindowOptions{
			Title:     appmeta.Title,
			Width:     1280,
			Height:    800,
			MinWidth:  1000,
			MinHeight: 640,
			StateKey:  appmeta.Name,

			// 隐藏标题栏但保留阴影/圆角/缩放边（Frameless 在 Windows 上会连阴影一起丢）。
			TitleBarStyle:     mygo.TitleBarHidden,
			HideWindowButtons: true,

			Content: ui.View(app.View),
		})
		app.SetWindow(win)
	})

	if err := mygo.App.Run(); err != nil {
		panic(err)
	}
}
