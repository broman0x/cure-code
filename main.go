package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/broman0x/forge-code/cmd"
)

// [EN] main is the entry point of the application. It handles global panic recovery and starts the command execution.
// [ID] main adalah titik masuk aplikasi. Ini menangani pemulihan panic global dan memulai eksekusi perintah.
func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("\n------------------------------------------------")
			fmt.Printf("CRITICAL PANIC: %v\n", r)
			fmt.Println("------------------------------------------------")
		}
		pauseExit()
	}()

	if err := cmd.Execute(); err != nil {
		if err.Error() != "" {
			fmt.Println("\n[!] Error:", err)
		}
		os.Exit(1)
	}
}

func pauseExit() {
	fmt.Println("\nPress 'Enter' to close window...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}
