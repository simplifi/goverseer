package cli

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{Use: "goverseer"}

func init() {
	// Disable timestamps in the log output for the CLI
	log.SetFlags(0)
}

// Execute the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
