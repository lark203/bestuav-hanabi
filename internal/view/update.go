package view

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

// 自动更新：mygo.Updater 的检查/下载/安装是网络与文件操作，放到 goroutine；
// 结果通过 channel 回到 UI 线程，进度用原子变量。开发构建 Enabled()=false，
// 设置页只显示“未启用”，不会发起请求。

type updatePhase uint8

const (
	updateIdle updatePhase = iota
	updateChecking
	updateUpToDate
	updateAvailable
	updateInstalling
	updateReady
	updateError
)

type updateEvent struct {
	up        *mygo.Update
	err       error
	installed bool
}

type updateState struct {
	enabled      bool
	enabledKnown bool
	phase        updatePhase
	available    *mygo.Update
	err          string
	down         atomic.Int64
	total        atomic.Int64
	events       chan updateEvent
}

// pollUpdate 把 goroutine 的结果搬到 UI 状态；检查/安装期间持续请求下一帧。
func (a *App) pollUpdate(c *ui.Context) {
	if a.upd.events == nil {
		return
	}
	for {
		select {
		case ev := <-a.upd.events:
			switch {
			case ev.installed:
				if ev.err != nil {
					a.upd.phase = updateError
					a.upd.err = updateErrorText(ev.err)
				} else {
					a.upd.phase = updateReady
					a.upd.err = ""
				}
			case ev.err != nil:
				a.upd.phase = updateError
				a.upd.err = updateErrorText(ev.err)
			case ev.up == nil:
				a.upd.phase = updateUpToDate
				a.upd.err = ""
			default:
				a.upd.phase = updateAvailable
				a.upd.available = ev.up
				a.upd.err = ""
			}
			c.Invalidate()
		default:
			if a.upd.phase == updateChecking || a.upd.phase == updateInstalling {
				c.Invalidate()
			}
			return
		}
	}
}

// updateEnabled 只探测一次：Enabled 会试写安装目录，不适合每帧调用。
func (a *App) updateEnabled() bool {
	if !a.upd.enabledKnown {
		a.upd.enabled = mygo.Updater.Enabled()
		a.upd.enabledKnown = true
	}
	return a.upd.enabled
}

// startUpdateCheck 起一个带超时的检查。
func (a *App) startUpdateCheck() {
	if !a.updateEnabled() || a.upd.phase == updateChecking || a.upd.phase == updateInstalling {
		return
	}
	a.upd.phase = updateChecking
	a.upd.err = ""
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		up, err := mygo.Updater.Check(ctx)
		a.upd.events <- updateEvent{up: up, err: err}
	}()
}

// startUpdateInstall 下载并安装当前发现的版本。
func (a *App) startUpdateInstall() {
	if a.upd.phase != updateAvailable || a.upd.available == nil {
		return
	}
	up := a.upd.available
	a.upd.phase = updateInstalling
	a.upd.down.Store(0)
	a.upd.total.Store(0)
	a.upd.err = ""
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		err := up.Install(ctx, func(downloaded, total int64) {
			a.upd.down.Store(downloaded)
			a.upd.total.Store(total)
		})
		a.upd.events <- updateEvent{installed: true, err: err}
	}()
}

// updateStatusLine 是设置页里的一句话状态。
func (a *App) updateStatusLine() string {
	if !a.updateEnabled() {
		return "当前构建未启用自动更新（开发版或未配置更新源）。"
	}
	switch a.upd.phase {
	case updateChecking:
		return "正在检查更新…"
	case updateUpToDate:
		return "已是最新版本。"
	case updateAvailable:
		return "发现新版本 " + a.upd.available.Version + "。"
	case updateInstalling:
		return "正在下载并安装…"
	case updateReady:
		return "新版本已就绪，重启后生效。"
	case updateError:
		return "更新失败，可重试。"
	default:
		return "自动更新已启用。"
	}
}

// updatePercent 返回 0..100；总大小未知时返回 -1。
func (a *App) updatePercent() int {
	total := a.upd.total.Load()
	if total <= 0 {
		return -1
	}
	p := int(a.upd.down.Load() * 100 / total)
	if p < 0 {
		p = 0
	}
	if p > 100 {
		p = 100
	}
	return p
}

// updateNotes 只取发布说明的前几行，并把 Markdown 标题/粗体/列表压成纯文本，
// 避免长说明把设置面板撑爆，也避免把 "##" 直接显示出来。
func updateNotes(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	var lines []string
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "- ")
		line = strings.TrimPrefix(line, "* ")
		line = strings.ReplaceAll(line, "**", "")
		line = strings.ReplaceAll(line, "`", "")
		if line != "" {
			lines = append(lines, "· "+line)
		}
		if len(lines) == 3 {
			break
		}
	}
	out := strings.Join(lines, "\n")
	if runes := []rune(out); len(runes) > 160 {
		out = string(runes[:160]) + "…"
	}
	return out
}

// updateErrorText 把底层错误压成一行，去掉 mygo 前缀与过长的 URL。
func updateErrorText(err error) string {
	if err == nil {
		return "未知错误"
	}
	s := strings.TrimSpace(err.Error())
	s = strings.TrimPrefix(s, "mygo: ")
	s = strings.ReplaceAll(s, "\n", " ")
	if runes := []rune(s); len(runes) > 120 {
		s = string(runes[:120]) + "…"
	}
	if s == "" {
		s = "未知错误"
	}
	return s
}

// updateControls 画检查/安装/重启按钮与发布说明。
func (a *App) updateControls(c *ui.Context) {
	if !a.updateEnabled() {
		return
	}
	switch a.upd.phase {
	case updateChecking:
		ui.Progress(c, 0).Label("检查更新")
	case updateInstalling:
		if p := a.updatePercent(); p >= 0 {
			ui.Progress(c, float64(p)).Label(fmt.Sprintf("下载 %d%%", p))
		} else {
			ui.Progress(c, 0).Label("下载中")
		}
	case updateAvailable:
		if ui.Button(c, "下载并安装 "+a.upd.available.Version).Clicked() {
			a.startUpdateInstall()
		}
		if notes := updateNotes(a.upd.available.Notes); notes != "" {
			ui.Text(c, notes).FontSize(a.fs(10)).TextColor(a.tok.TextFaint)
		}
	case updateReady:
		if ui.Button(c, "立即重启并完成更新").Clicked() {
			mygo.App.Relaunch()
		}
	case updateError:
		ui.Row(c).Gap(8).Children(func() {
			if ui.Button(c, "重试").Clicked() {
				a.startUpdateCheck()
			}
		})
		ui.Text(c, a.upd.err).FontSize(a.fs(10)).TextColor(a.tok.Danger)
	default:
		if ui.Button(c, "检查更新").Clicked() {
			a.startUpdateCheck()
		}
	}
}
