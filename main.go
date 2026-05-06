package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime"

	"github.com/broman0x/cure-code/cmd"
	"github.com/broman0x/cure-code/internal/config"
)

// [EN] main is the entry point of the application. It handles global panic recovery and starts the command execution.
// [ID] main adalah titik masuk aplikasi. Ini menangani pemulihan panic global dan memulai eksekusi perintah.
func main() {
	// [EN] Ensure config directories exist before any command runs
	// [ID] Pastikan direktori konfigurasi ada sebelum perintah apapun berjalan
	if err := config.EnsureConfigDirs(); err != nil {
		fmt.Printf("[!] Config dir error: %v\n", err)
	}

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("\n------------------------------------------------")
			fmt.Printf("CRITICAL PANIC: %v\n", r)
			fmt.Println("------------------------------------------------")
			os.Exit(1)
		}
	}()

	if err := cmd.Execute(); err != nil {
		if err.Error() != "" {
			fmt.Println("\n[!] Error:", err)
		}
		if !cmd.SkipPause {
			pauseExit()
		}
		os.Exit(1)
	}
	
	if !cmd.SkipPause {
		pauseExit()
	}
}

func pauseExit() {
	if runtime.GOOS != "windows" {
		return
	}
	fmt.Println("\nPress 'Enter' to close window...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}
