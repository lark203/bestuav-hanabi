package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestParseDefaults(t *testing.T) {
	o, err := Parse(nil, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if o.Version || o.ResetSettings || o.SafeMode || o.Page != "" {
		t.Errorf("默认值不对: %+v", o)
	}
}

func TestParseAll(t *testing.T) {
	o, err := Parse([]string{"--version", "--page=settings", "--reset-settings", "--safe-mode"}, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	if !o.Version || !o.ResetSettings || !o.SafeMode || o.Page != "settings" {
		t.Errorf("解析结果不对: %+v", o)
	}
}

func TestPageFromArgs(t *testing.T) {
	cases := map[string]string{
		"--page=library":       "library",
		"-page=settings":       "settings",
		"--page wallpaper":     "wallpaper",
		"":                     "",
		"--safe-mode --page=x": "x",
	}
	for args, want := range cases {
		if got := PageFromArgs(splitArgs(args)); got != want {
			t.Errorf("PageFromArgs(%q) = %q，期望 %q", args, got, want)
		}
	}
}

func splitArgs(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return strings.Fields(s)
}
