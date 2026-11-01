// 花火 HanaBi：以角色主题为核心的桌面 UI（学习自 HanaBi-ui-Genshin 的视觉与交互）。
//
// 窗口隐藏系统标题栏与窗口控件，由界面自绘（见 internal/view/app.go 的 titlebar）：
// 否则自带的标题栏会和界面里的那条叠成两条。
package main

import (
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"bestuav-hanabi/internal/view"
)

func main() {
	a := view.NewApp()

	mygo.App.WhenReady(func() {
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
