package view

import "testing"

// TestSpringProgress 钉住主题切换的弹簧曲线：
// 起点 0、单调、比匀速更早到达（0.1s 已经过半），到 0.42s 基本到 1。
func TestSpringProgress(t *testing.T) {
	if v := springProgress(0); v != 0 {
		t.Errorf("t=0 应为 0，实际 %v", v)
	}
	if v := springProgress(0.42); v < 0.999 {
		t.Errorf("0.42s 应基本到位，实际 %v", v)
	}
	prev := float32(0)
	for i := 1; i <= 42; i++ {
		v := springProgress(float64(i) / 100)
		if v < prev-1e-6 {
			t.Fatalf("t=%.2fs 出现回退：%v < %v", float64(i)/100, v, prev)
		}
		prev = v
	}
	// 弹簧起步比 slowerstep 快：0.1s 时已经过半（匀速是 0.24）。
	if v := springProgress(0.1); v < 0.6 {
		t.Errorf("0.1s 应已过半（弹簧起步快），实际 %v", v)
	}
}
