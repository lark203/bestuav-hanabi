//go:build windows

package cli

import "syscall"

// attachConsole 接回父进程的控制台（GUI 子系统程序默认没有控制台）。
func attachConsole() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	proc := kernel32.NewProc("AttachConsole")
	// ATTACH_PARENT_PROCESS = -1
	_, _, _ = proc.Call(^uintptr(0))
}
