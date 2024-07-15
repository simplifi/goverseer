package cli

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/simplifi/goverseer/internal/goverseer/config"
	"github.com/simplifi/goverseer/internal/goverseer/overseer"
	"github.com/spf13/cobra"
)

var (
	configFile string
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the goverseer service",
	Run: func(cmd *cobra.Command, args []string) {
		start()
	},
}

func init() {
	startCmd.Flags().StringVarP(
		&configFile,
		"config",
		"c",
		"/etc/goverseer.yaml",
		"A configuration file for the goverseer service")
	rootCmd.AddCommand(startCmd)
}

// loadConfig loads the configuration file
func loadConfig() *config.Config {
	cfg, err := config.FromFile(configFile)
	if err != nil {
		log.Fatalf("error loading configuration: %v", err)
	}
	return cfg
}

// start starts the goverseer service
func start() {
	cfg := loadConfig()

	overseer, err := overseer.New(cfg)
	if err != nil {
		log.Fatalf("overseer error: %v", err)
	}

	// Listen for OS signals and wait
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-signalChan
		log.Printf("\nreceived signal: %s", sig)
		log.Println("shutting down...")
		overseer.Stop()
	}()

	overseer.Run()
}
