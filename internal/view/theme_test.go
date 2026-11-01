package view

import (
	"testing"

	"bestuav-hanabi/internal/theme"
)

// TestThemeSwitch 检查切换主题时令牌与角色一起换，且非法 id 不会把界面弄坏。
func TestThemeSwitch(t *testing.T) {
	a := NewApp()
	if a.char.ID != theme.Default().ID {
		t.Fatalf("默认主题应为 %s，实际 %s", theme.Default().ID, a.char.ID)
	}

	// 切到火：令牌必须同步换掉（只换 char 会让界面还是旧配色）。
	if _, ok := theme.ByID("yoimiya"); !ok {
		t.Skip("没有 yoimiya 主题")
	}
	a.setCharacter("yoimiya")
	if a.char.ID != "yoimiya" {
		t.Fatalf("切换失败，当前是 %s", a.char.ID)
	}
	if a.tok.Accent != a.char.Tok.Accent {
		t.Error("令牌没有跟着角色一起换")
	}
	if a.tok.Accent == theme.Default().Tok.Accent {
		t.Error("切换后强调色仍与默认主题相同")
	}

	// 非法 id：保持原样，不 panic。
	a.setCharacter("no-such-character")
	if a.char.ID != "yoimiya" {
		t.Errorf("非法 id 不应改变当前主题，实际变成 %s", a.char.ID)
	}

	// 重复切同一个：幂等。
	before := a.tok
	a.setCharacter("yoimiya")
	if a.tok != before {
		t.Error("重复切换同一角色不应改变令牌")
	}
}
