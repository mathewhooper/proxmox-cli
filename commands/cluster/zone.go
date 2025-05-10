package cluster

import (
	"fmt"

	"github.com/spf13/cobra"
)

func ZoneCommand() *cobra.Command {
    var zoneCmd = &cobra.Command{
        Use:   "sdn",
        Short: "Manage Software Defined Networking (SDN) in Proxmox",
    }

    // Add the create-zone, delete-zone, update-zone, and apply-config subcommands
    zoneCmd.AddCommand(createZoneCommand())
    zoneCmd.AddCommand(deleteZoneCommand())
    zoneCmd.AddCommand(updateZoneCommand())

    return zoneCmd
}

// createZoneCommand creates a new Cobra command for creating a new SDN zone.
func createZoneCommand() *cobra.Command {
	var zoneName string
	var zoneType string

	var createZoneCmd = &cobra.Command{
		Use:   "create-zone",
		Short: "Create a new SDN zone",
		Run: func(cmd *cobra.Command, args []string) {
			if zoneName == "" || zoneType == "" {
				fmt.Println("Error: zone name and type are required")
				return
			}

			// Logic to create a new SDN zone
			fmt.Printf("Creating SDN zone: Name=%s, Type=%s\n", zoneName, zoneType)
		},
	}

	createZoneCmd.Flags().StringVarP(&zoneName, "name", "n", "", "Name of the SDN zone")
	createZoneCmd.Flags().StringVarP(&zoneType, "type", "t", "", "Type of the SDN zone")

	return createZoneCmd
}

// deleteZoneCommand creates a new Cobra command for deleting an SDN zone.
func deleteZoneCommand() *cobra.Command {
	var zoneName string

	var deleteZoneCmd = &cobra.Command{
		Use:   "delete-zone",
		Short: "Delete an existing SDN zone",
		Run: func(cmd *cobra.Command, args []string) {
			if zoneName == "" {
				fmt.Println("Error: zone name is required")
				return
			}

			// Logic to delete an SDN zone
			fmt.Printf("Deleting SDN zone: Name=%s\n", zoneName)
		},
	}

	deleteZoneCmd.Flags().StringVarP(&zoneName, "name", "n", "", "Name of the SDN zone to delete")

	return deleteZoneCmd
}

// updateZoneCommand creates a new Cobra command for updating an existing SDN zone.
func updateZoneCommand() *cobra.Command {
	var zoneName string
	var newZoneType string

	var updateZoneCmd = &cobra.Command{
		Use:   "update-zone",
		Short: "Update an existing SDN zone",
		Run: func(cmd *cobra.Command, args []string) {
			if zoneName == "" || newZoneType == "" {
				fmt.Println("Error: zone name and new type are required")
				return
			}

			// Logic to update an SDN zone
			fmt.Printf("Updating SDN zone: Name=%s, New Type=%s\n", zoneName, newZoneType)
		},
	}

	updateZoneCmd.Flags().StringVarP(&zoneName, "name", "n", "", "Name of the SDN zone to update")
	updateZoneCmd.Flags().StringVarP(&newZoneType, "new-type", "t", "", "New type of the SDN zone")

	return updateZoneCmd
}