package cluster

import (
	"github.com/spf13/cobra"
)

// NewClusterCommand creates a new Cobra command for cluster-related operations.
func ClusterCommand() *cobra.Command {
	var clusterCmd = &cobra.Command{
		Use:   "cluster",
		Short: "Manage Proxmox clusters",
	}

	// Add subcommands here
    clusterCmd.AddCommand(SDNCommand())

	return clusterCmd
}
