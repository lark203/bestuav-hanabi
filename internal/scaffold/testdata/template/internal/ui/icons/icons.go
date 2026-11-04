// Package icons 提供菜单用的线性图标（Lucide 风格，stroke=currentColor，
// 颜色跟随 TextColor，选中时可用强调色）。
package icons

import "github.com/egoist/mygo/ui"

var (
	// Palette 主题库：调色盘。
	Palette = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round" d="M12 3a9 9 0 1 0 0 18h1.2a2.4 2.4 0 0 0 0-4.8h-1a2 2 0 0 1 0-4H16a5 5 0 0 0 0-10z"/><circle cx="7.6" cy="11.4" r="1.1" fill="currentColor"/><circle cx="9.8" cy="7.4" r="1.1" fill="currentColor"/></svg>`))

	// Picture 壁纸：山与日。
	Picture = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><rect x="3" y="4" width="18" height="16" rx="3" fill="none" stroke="currentColor" stroke-width="1.8"/><circle cx="8.6" cy="9.4" r="1.5" fill="currentColor"/><path fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" d="m4.5 17 4.8-4.8 3.6 3.6 2.6-2.6 4 4"/></svg>`))

	// Minimize 窗口最小化。
	Minimize = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" d="M5 12h14"/></svg>`))

	// Maximize 窗口最大化：方框。
	Maximize = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><rect x="6" y="6" width="12" height="12" rx="1.5" fill="none" stroke="currentColor" stroke-width="1.6"/></svg>`))

	// Close 关闭：叉。
	Close = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" d="M6.5 6.5l11 11M17.5 6.5l-11 11"/></svg>`))

	// Sliders 设置：滑块。
	Sliders = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" d="M4 7h9M18 7h2M4 12h3M12 12h8M4 17h11M20 17h0"/><circle cx="15.5" cy="7" r="2.2" fill="none" stroke="currentColor" stroke-width="1.8"/><circle cx="9.5" cy="12" r="2.2" fill="none" stroke="currentColor" stroke-width="1.8"/><circle cx="17.5" cy="17" r="2.2" fill="none" stroke="currentColor" stroke-width="1.8"/></svg>`))

	// Square 通用页面：圆角方块。
	Square = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><rect x="4" y="4" width="16" height="16" rx="3.5" fill="none" stroke="currentColor" stroke-width="1.8"/></svg>`))

	// Plus 新建项目：加号。
	Plus = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" d="M12 5v14M5 12h14"/></svg>`))
)
