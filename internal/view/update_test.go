package view

import (
	"errors"
	"strings"
	"testing"
)

func TestUpdateNotes(t *testing.T) {
	got := updateNotes("## 0.2.0\n\n- 新增角色主题效果\n- **修复** 更新检查\n- `优化` 启动速度\n- 多余的一行")
	want := "· 新增角色主题效果\n· 修复 更新检查\n· 优化 启动速度"
	if got != want {
		t.Fatalf("发布说明清理不对：\n%q\nwant\n%q", got, want)
	}

	long := strings.Repeat("很长的说明", 80)
	if out := updateNotes(long); len([]rune(out)) > 161 || !strings.HasSuffix(out, "…") {
		t.Errorf("超长说明应截断：len=%d", len(out))
	}
	if out := updateNotes("## 只有标题"); out != "" {
		t.Errorf("只有标题时应为空，实际 %q", out)
	}
}

func TestUpdateErrorText(t *testing.T) {
	e := errors.New("mygo: checking for updates: Get \"https://example.com\": dial tcp: timeout")
	got := updateErrorText(e)
	if strings.HasPrefix(got, "mygo:") || strings.Contains(got, "\n") {
		t.Errorf("错误文本不应带前缀或换行：%q", got)
	}
	if updateErrorText(nil) != "未知错误" {
		// 调用方不会传 nil，但兜底不应 panic。
		t.Error("nil 错误应有兜底")
	}
	long := errors.New(strings.Repeat("x", 300))
	if out := updateErrorText(long); len(out) > 123 || !strings.HasSuffix(out, "…") {
		t.Errorf("超长错误应截断：len=%d", len(out))
	}
}

func TestUpdatePercent(t *testing.T) {
	a := &App{}
	if p := a.updatePercent(); p != -1 {
		t.Errorf("总大小未知时应为 -1，实际 %d", p)
	}
	a.upd.total.Store(200)
	a.upd.down.Store(50)
	if p := a.updatePercent(); p != 25 {
		t.Errorf("50/200 应为 25，实际 %d", p)
	}
	a.upd.down.Store(300)
	if p := a.updatePercent(); p != 100 {
		t.Errorf("超过总量应夹到 100，实际 %d", p)
	}
}
