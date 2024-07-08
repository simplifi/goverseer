package cli

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/simplifi/goverseer/internal/goverseer/config"
	"github.com/simplifi/goverseer/internal/goverseer/manager"
	"github.com/spf13/cobra"
)

var (
	configDir string
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
		&configDir,
		"config-dir",
		"c",
		"/etc/goverseer.d",
		"Directory containing configuration files")
	rootCmd.AddCommand(startCmd)
}

// loadConfigs loads the configuration files from the config directory
func loadConfigs() []*config.Config {
	cfgs, err := config.FromPath(configDir)
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}
	return cfgs
}

// validateConfigs validates the configuration files
func validateConfigs(cfgs []*config.Config) {
	var errors []error

	for _, cfg := range cfgs {
		if err := cfg.Validate(); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		log.Printf("Validation errors found in %d configuration(s):\n", len(errors))
		for _, err := range errors {
			log.Printf("- %s\n", err.Error())
		}
		log.Fatalf("\nExiting due to validation errors")
	}
}

// start starts the goverseer service
func start() {
	cfgs := loadConfigs()
	validateConfigs(cfgs)

	mgr := manager.NewManager(cfgs)
	go func() {
		if err := mgr.Run(); err != nil {
			log.Fatalf("Manager run error: %v", err)
		}
	}()

	// Listen for OS signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigCh
	log.Printf("\nReceived signal: %s", sig)
	log.Println("Shutting down...")
	mgr.Stop()
}
