//go:build !windows

package cli

// attachConsole 其他平台本来就有标准输出。
func attachConsole() {}
