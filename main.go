package main

import (
	"proxmox-cli/commands"
	"proxmox-cli/commands/cluster"

	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "proxmox-cli",
		Short: "A CLI for interacting with Proxmox servers",
	}

	rootCmd.AddCommand(commands.LoginCommand())
    rootCmd.AddCommand(commands.ValidateLoginCommand())
    rootCmd.AddCommand(cluster.ClusterCommand())

	rootCmd.Execute()
}