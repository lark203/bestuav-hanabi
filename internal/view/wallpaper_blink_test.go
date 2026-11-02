package view

import (
	"math"
	"testing"
)

// TestBlinkAmount 钉住眨眼的节奏：窗口外为 0、闭合到 1、而且不会一直不动。
func TestBlinkAmount(t *testing.T) {
	// 0 秒正好在窗口起点：应当立刻进入闭合过程。
	if v := blinkAmount(0); v != 0 {
		t.Errorf("t=0 应为 0，实际 %v", v)
	}
	// 闭合中（blinkClose 的中点）应当介于 0 和 1 之间。
	mid := blinkAmount(blinkClose / 2)
	if mid <= 0.05 || mid >= 0.95 {
		t.Errorf("闭合中途应为中间值，实际 %v", mid)
	}
	// 保持段应当完全闭合。
	hold := blinkAmount(blinkClose + blinkHold/2)
	if hold < 0.999 {
		t.Errorf("保持段应完全闭合，实际 %v", hold)
	}
	// 睁开段结束之后应当回到 0。
	after := blinkAmount(blinkClose + blinkHold + blinkOpen + 0.02)
	if after != 0 {
		t.Errorf("睁眼后应为 0，实际 %v", after)
	}
	// 一个周期内只有一次眨眼；10 秒内至少两次（奇数轮会提前）。
	blinks := 0
	for i := 0; i < 1000; i++ {
		tm := float64(i) * 0.01
		if blinkAmount(tm) > 0.999 && blinkAmount(tm-0.01) <= 0.999 {
			blinks++
		}
	}
	if blinks < 2 || blinks > 3 {
		t.Errorf("10 秒内眨眼次数 %d，期望 2~3 次", blinks)
	}
}

// TestBlinkAmountPeriodic 检查间隔不是死板的等距：奇数轮提前 blinkJitter 秒。
func TestBlinkAmountPeriodic(t *testing.T) {
	first := -1.0
	for i := 0; i < 1000; i++ {
		tm := float64(i) * 0.01
		if blinkAmount(tm) > 0.999 && (first < 0 || tm-first > 1) {
			if first < 0 {
				first = tm
				continue
			}
			gap := tm - first
			if math.Abs(gap-blinkPeriod) < 0.05 {
				t.Errorf("第二轮应当提前 %.1fs，实际间隔 %.2fs", blinkJitter, gap)
			}
			return
		}
	}
	t.Error("没找到第二次眨眼")
}
