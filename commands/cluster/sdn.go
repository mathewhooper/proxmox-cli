package cluster

import (
	"fmt"

	"github.com/spf13/cobra"
)

// SDNCommand creates a new Cobra command for SDN-related operations.
func SDNCommand() *cobra.Command {
	var sdnCmd = &cobra.Command{
		Use:   "sdn",
		Short: "Manage Software Defined Networking (SDN) in Proxmox",
	}

	// Add the create-zone and delete-zone subcommands
	sdnCmd.AddCommand(ZoneCommand())
	sdnCmd.AddCommand(VnetCommand())
	sdnCmd.AddCommand(applyZoneConfigCommand())

	return sdnCmd
}

// applyZoneConfigCommand creates a new Cobra command for applying the configuration of an SDN zone.
func applyZoneConfigCommand() *cobra.Command {
	var zoneName string

	var applyZoneCmd = &cobra.Command{
		Use:   "apply-config",
		Short: "Apply the configuration of an SDN zone",
		Run: func(cmd *cobra.Command, args []string) {
			if zoneName == "" {
				fmt.Println("Error: zone name is required")
				return
			}

			// Logic to apply the configuration of an SDN zone
			fmt.Printf("Applying configuration for SDN zone: Name=%s\n", zoneName)
		},
	}

	applyZoneCmd.Flags().StringVarP(&zoneName, "name", "n", "", "Name of the SDN zone to apply configuration")

	return applyZoneCmd
}
