package view

import "github.com/egoist/mygo/ui"

// pageScrollReserve 是页面纵向滚动条预留的右侧空档。
//
// mygo 的滚动条是覆盖式绘制的（不占布局宽度），所以滚动内容必须自己让出
// 条宽 + 视觉透气，否则输入框/卡片会贴到滚动条上。所有可滚动页面统一走
// pageScroll，新页面不要再手写 ui.Scroll + Padding。
const pageScrollReserve = 24

// pageScroll 是页面滚动体的统一容器：铺满剩余高度，右侧给滚动条留空。
func pageScroll(c *ui.Context, gap float32, build func()) {
	ui.Scroll(c).Grow(1).Children(func() {
		ui.Column(c).Padding(4, pageScrollReserve, 10, 0).Gap(gap).Children(build)
	})
}
