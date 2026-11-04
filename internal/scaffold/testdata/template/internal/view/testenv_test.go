package view

import (
	"bestuav-hanabi/internal/visual"
)

// 测试用：去掉角色素材（壁纸/头像/动态遮罩）。
//
// 离屏测试的 Tester 是软件光栅器，带壁纸着色器时 1920×1080 一帧要几十秒；
// 页面/绘制测试只关心界面本身，素材缺失正好走「纯色降级」分支。
func stripped(a *App) *App {
	ch := a.char
	ch.Bg, ch.Icon = "", ""
	ch.Motion = visual.Motion{}
	a.char, a.fromChar = ch, ch
	a.tok, a.fromTok = ch.Tok, ch.Tok
	a.fromBg, a.fromFlow = nil, nil
	return a
}
