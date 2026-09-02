package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	cfgFile string
	verbose bool
)

// RootCmd is the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "porta",
	Short: "PORTA — Expose your local multi-service application to the world with one command",
	Long: `PORTA is a developer utility and runtime engine that binds multiple local applications 
running on localhost into a single consolidated reverse proxy gateway and exposes them to the public 
internet via a secure encrypted HTTPS tunnel.`,
}

func init() {
	RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "porta.yaml", "config file path")
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose debug logging")

	RootCmd.AddCommand(initCmd)
	RootCmd.AddCommand(startCmd)
	RootCmd.AddCommand(statusCmd)
	RootCmd.AddCommand(logsCmd)
	RootCmd.AddCommand(configCmd)
	RootCmd.AddCommand(doctorCmd)
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
