package main

import (
	"proxmox-cli/commands"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var logger = logrus.New()

func init() {
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logger.SetLevel(logrus.InfoLevel)
}

func main() {
	// Create the root command
	var rootCmd = &cobra.Command{
		Use:   "proxmox-cli",
		Short: "A CLI for interacting with Proxmox servers",
	}

	// Add the login command from the commands package
	rootCmd.AddCommand(commands.NewLoginCommand())
    rootCmd.AddCommand(commands.ValidateLoginCommand())

	// Execute the root command
	rootCmd.Execute()
}