package intro

import (
	"math"
	"testing"
)

// TestTimeline 钉住时间轴：6.20s 结束，dt 被夹到 0.20。
func TestTimeline(t *testing.T) {
	in := New(1280, 800)
	if in.Finished() {
		t.Fatal("初始状态不应已结束")
	}

	// 单帧超大 dt 应被夹到 0.20，而不是一步跳完。
	in.Update(10)
	if in.T > 0.2001 {
		t.Errorf("dt 未被夹住：T=%.3f，期望 ≤0.20", in.T)
	}

	for i := 0; i < 100; i++ {
		in.Update(0.1)
		if in.Finished() {
			break
		}
	}
	if !in.Finished() {
		t.Fatalf("10 秒后仍未结束，T=%.3f", in.T)
	}
	if in.T != Duration {
		t.Errorf("结束时应停在 %.2f，实际 %.3f", Duration, in.T)
	}
}

// TestPose 检查标识的起止位置：中央偏下 → 侧栏停靠点。
func TestPose(t *testing.T) {
	const w, h = 1280.0, 800.0
	const dockX, dockY = 37.0, 35.0

	in := New(w, h)
	start := in.Pose(dockX, dockY)
	if math.Abs(float64(start.X-w/2)) > 0.5 {
		t.Errorf("起始应水平居中：x=%.1f，期望 %.1f", start.X, w/2)
	}
	if math.Abs(float64(start.Y-h*0.66)) > 0.5 {
		t.Errorf("起始应在 66%% 高度：y=%.1f，期望 %.1f", start.Y, h*0.66)
	}
	if start.Opacity != 0 {
		t.Errorf("起始应完全透明（0.18s 才开始显现），实际 %.3f", start.Opacity)
	}

	// 走到结尾。
	for !in.Finished() {
		in.Update(0.1)
	}
	end := in.Pose(dockX, dockY)
	if math.Abs(float64(end.X-dockX)) > 0.5 || math.Abs(float64(end.Y-dockY)) > 0.5 {
		t.Errorf("结束应停在 (%.0f,%.0f)，实际 (%.1f,%.1f)", dockX, dockY, end.X, end.Y)
	}
	if math.Abs(float64(end.Radius-dockR)) > 0.5 {
		t.Errorf("结束半径应为 %.1f（品牌 glyph 大小），实际 %.1f", dockR, end.Radius)
	}
}

// TestParticles 检查粒子都在合理范围，且分布是确定性的。
func TestParticles(t *testing.T) {
	a, b := New(1280, 800), New(1280, 800)
	for i := 0; i < LantCount; i++ {
		if a.lant[i] != b.lant[i] {
			t.Fatalf("粒子初始化不确定：第 %d 盏不一致", i)
		}
	}
	for i := 0; i < SnowCount; i++ {
		if a.snow[i] != b.snow[i] {
			t.Fatalf("粒子初始化不确定：第 %d 片雪不一致", i)
		}
	}

	// 跑 3 秒后，可见的灯应有一半以上在画面内。
	a2 := New(1280, 800)
	for i := 0; i < 30; i++ {
		a2.Update(0.1)
	}
	inFrame := 0
	for i := 0; i < LantCount; i++ {
		_, y, alpha, size := a2.LanternAt(i, 800)
		if alpha > 0.05 && y > -0.1 && y < 1.1 && size > 0 {
			inFrame++
		}
	}
	if inFrame < LantCount/2 {
		t.Errorf("3 秒时画面内只有 %d/%d 盏灯，构图可能不对", inFrame, LantCount)
	}
}

// TestSkip 检查跳过能在 0.9s 内走完，且不会跳过 6.20s。
func TestSkip(t *testing.T) {
	in := New(1280, 800)
	for i := 0; i < 10; i++ {
		in.Update(0.1) // 走到 1.0s
	}
	in.Skip()

	elapsed := float32(0)
	for !in.Finished() {
		in.Update(1.0 / 60)
		elapsed += 1.0 / 60
		if elapsed > 2 {
			t.Fatalf("跳过超过 2 秒仍未结束（T=%.2f），可能没走快进分支", in.T)
		}
	}
	if elapsed > 1.0 {
		t.Errorf("跳用过慢：耗时 %.2fs，期望 ≤0.9s", elapsed)
	}
}

// TestOpacityAndHandoff 检查淡出与交接的单调性。
func TestOpacityAndHandoff(t *testing.T) {
	in := New(1280, 800)
	prevO, prevK := in.Opacity(), in.HandoffK()
	for !in.Finished() {
		in.Update(1.0 / 60)
		o, k := in.Opacity(), in.HandoffK()
		if o > prevO+1e-5 {
			t.Fatalf("不透明度不应上升：t=%.2f %.4f → %.4f", in.T, prevO, o)
		}
		if k < prevK-1e-5 {
			t.Fatalf("交接进度不应回退：t=%.2f %.4f → %.4f", in.T, prevK, k)
		}
		prevO, prevK = o, k
	}
	if prevO != 0 {
		t.Errorf("结束时不应透明度应为 0，实际 %.3f", prevO)
	}
	if prevK < 0.999 {
		t.Errorf("结束时交接进度应到 1，实际 %.3f", prevK)
	}
}
