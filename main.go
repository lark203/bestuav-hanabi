// 花火 HanaBi：以角色主题为核心的桌面 UI（学习自 HanaBi-ui-Genshin 的视觉与交互）。
//
// 骨架阶段只做：窗口 + 主题令牌 + 三栏布局 + 主题库占位页。
// 动态壁纸、开场动画按评估文档分阶段接入（见 docs/）。
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
			Content:   ui.View(a.View),
		})
		a.SetWindow(win)
	})

	if err := mygo.App.Run(); err != nil {
		panic(err)
	}
}
