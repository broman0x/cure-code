//go:build windows

package cmd

import (
	"syscall"
)

func flushWindowsInput() {
	handle, err := syscall.GetStdHandle(syscall.STD_INPUT_HANDLE)
	if err == nil {
		kernel32 := syscall.NewLazyDLL("kernel32.dll")
		flush := kernel32.NewProc("FlushConsoleInputBuffer")
		flush.Call(uintptr(handle))
	}
}
