// 花火 HanaBi：以角色主题为核心的桌面 UI（学习自 HanaBi-ui-Genshin 的视觉与交互）。
//
// 窗口隐藏系统标题栏与窗口控件，由界面自绘（见 internal/view/app.go 的 titlebar）：
// 否则自带的标题栏会和界面里的那条叠成两条。
package main

import (
	"runtime/debug"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"bestuav-hanabi/internal/settings"
	"bestuav-hanabi/internal/view"
)

func main() {
	// 壁纸解码/缩放是几十 MB 的短命分配：默认 GOGC=100 会先把堆翻倍再回收，
	// 切几次主题进程就涨到几百 MB。50 让 GC 早一点介入；再加一个软上限兜底，
	// 防止异常增长时把内存吃满（超过上限只会更勤快地 GC，不会直接失败）。
	debug.SetGCPercent(50)
	debug.SetMemoryLimit(384 << 20)

	mygo.App.WhenReady(func() {
		// 先读用户设置：启动主题、字号、动效开关都在这里恢复。
		a := view.NewAppWithSettings(settings.Load())
		// 退出前把还没落盘的拖动/开关写掉（平时是 300ms 防抖保存）。
		mygo.App.OnQuit(a.FlushSettings)
		win := mygo.NewWindow(mygo.WindowOptions{
			Title:     "花火 HanaBi",
			Width:     1280,
			Height:    800,
			MinWidth:  1000,
			MinHeight: 640,
			StateKey:  "bestuav-hanabi",

			// 隐藏标题栏但保留窗口阴影、圆角与可缩放边缘
			// （Frameless 在 Windows 上会连阴影一起丢掉）。
			TitleBarStyle: mygo.TitleBarHidden,
			// 连系统按钮也隐藏：按钮由界面自绘，并调用 Minimize/Maximize/Close。
			HideWindowButtons: true,

			Content: ui.View(a.View),
		})
		a.SetWindow(win)
	})

	if err := mygo.App.Run(); err != nil {
		panic(err)
	}
}
