package main

import (
	"proxmox-cli/commands"
	"proxmox-cli/commands/cluster"
	"proxmox-cli/config"

	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "proxmox-cli",
		Short: "A comprehensive CLI for interacting with Proxmox VE API",
		Long: `proxmox-cli is a command-line tool that provides access to the Proxmox VE API.
It supports managing nodes, virtual machines, containers, storage, networking, and more.`,
	}

	// Add persistent trust flag
	rootCmd.PersistentFlags().BoolVarP(&config.Trust, "trust", "t", false, "Trust SSL certificates")

	// Authentication commands
	rootCmd.AddCommand(commands.LoginCommand())
	rootCmd.AddCommand(commands.ValidateLoginCommand())

	// Resource management commands
	rootCmd.AddCommand(commands.NodesCommand())
	rootCmd.AddCommand(commands.VMCommand())
	rootCmd.AddCommand(commands.StorageCommand())
	rootCmd.AddCommand(cluster.ClusterCommand())

	rootCmd.Execute()
}
