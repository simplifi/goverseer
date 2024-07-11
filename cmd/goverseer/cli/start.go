package cli

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/simplifi/goverseer/internal/goverseer/config"
	"github.com/simplifi/goverseer/internal/goverseer/overseer"
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
		log.Fatalf("error loading configuration: %v", err)
	}
	return cfgs
}

// start starts the goverseer service
func start() {
	wg := sync.WaitGroup{}
	stop := make(chan struct{})

	cfgs := loadConfigs()

	for _, cfg := range cfgs {
		overseer, err := overseer.New(cfg, stop)
		if err != nil {
			log.Fatalf("overseer run error: %v", err)
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			overseer.Run()
		}()
	}

	// Listen for OS signals and wait
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-signalChan
	log.Printf("\nreceived signal: %s", sig)
	log.Println("shutting down...")

	close(stop)
	wg.Wait()
	log.Println("shutdown complete")
}
