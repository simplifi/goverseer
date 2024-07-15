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

func init() {
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start the goverseer service",
		Run: func(cmd *cobra.Command, args []string) {
			if config, err := cmd.Flags().GetString("config"); err != nil {
				log.Fatalf("error getting config flag: %v", err)
			} else {
				start(config)
			}
		},
	}

	startCmd.Flags().StringP(
		"config",
		"c",
		"/etc/goverseer.yaml",
		"A configuration file for the goverseer service")

	rootCmd.AddCommand(startCmd)
}

// loadConfig loads the configuration file
func loadConfig(configFile string) *config.Config {
	cfg, err := config.FromFile(configFile)
	if err != nil {
		log.Fatalf("error loading configuration: %v", err)
	}
	return cfg
}

// start starts the goverseer service
func start(configFile string) {
	cfg := loadConfig(configFile)

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
