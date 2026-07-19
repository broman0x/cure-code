//go:build !windows

package cmd

func flushWindowsInput() {
	// No-op on non-Windows platforms
}
