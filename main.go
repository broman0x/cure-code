package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime"

	"github.com/broman0x/cure-code/cmd"
)

// [EN] main is the entry point of the application. It handles global panic recovery and starts the command execution.
// [ID] main adalah titik masuk aplikasi. Ini menangani pemulihan panic global dan memulai eksekusi perintah.
func main() {
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
