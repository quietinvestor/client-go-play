package main

import (
	"os"

	"github.com/quietinvestor/client-go-play/internal/cmd"
)

func main() {
	rootCmd := cmd.NewRootCmd()
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
