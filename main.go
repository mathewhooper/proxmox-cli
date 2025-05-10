package main

import (
	"proxmox-cli/commands"

	"github.com/spf13/cobra"
)

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