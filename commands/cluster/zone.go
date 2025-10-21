package cluster

import (
	"fmt"

	"proxmox-cli/config"
	"proxmox-cli/services"

	"encoding/json"

	"github.com/spf13/cobra"
)

// ZoneCommand creates and returns the SDN (Software Defined Networking) management command
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

			// Zone type can be : simple, vlan, vxlan, gre, ipsec, l2tpv3, vxlan-ipsec, l2tpv3-ipsec
			// If the zone type is not one of the above, return an error
			if zoneType != "simple" && zoneType != "vlan" && zoneType != "vxlan" && zoneType != "gre" && zoneType != "ipsec" && zoneType != "l2tpv3" && zoneType != "vxlan-ipsec" && zoneType != "l2tpv3-ipsec" {
				fmt.Println("Error: invalid zone type")
				return
			}

			CreateZone(zoneName, zoneType, config.Trust)

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

			DeleteZone(zoneName, config.Trust)

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

			UpdateZone(zoneName, newZoneType, config.Trust)

			// Logic to update an SDN zone
			fmt.Printf("Updating SDN zone: Name=%s, New Type=%s\n", zoneName, newZoneType)
		},
	}

	updateZoneCmd.Flags().StringVarP(&zoneName, "name", "n", "", "Name of the SDN zone to update")
	updateZoneCmd.Flags().StringVarP(&newZoneType, "new-type", "t", "", "New type of the SDN zone")

	return updateZoneCmd
}

// CreateZone creates a new SDN zone in the Proxmox cluster
func CreateZone(zoneName string, zoneType string, trust bool) bool {
	sessionService, err := services.NewSessionService(config.Logger)
	if err != nil {
		config.Logger.Error("Error initializing session service: ", err)
		return false
	}

	sessionData, err := sessionService.ReadSessionFile()
	if err != nil {
		config.Logger.Error("Error reading session file: ", err)
		return false
	}

	uri := fmt.Sprintf("%s://%s:%d/api2/extjs/cluster/zones", sessionData.HttpScheme, sessionData.Server, sessionData.Port)

	httpService := services.NewHttpService(config.Logger, trust)

	payload := fmt.Sprintf("zone=%s&type=%s", zoneName, zoneType)
	headers := map[string]string{
		"Content-Type":        "application/x-www-form-urlencoded; charset=UTF-8",
		"CSRFPreventionToken": sessionData.Response.Data.CSRFPreventionToken,
	}

	body, err := httpService.Post(uri, payload, headers, nil)
	if err != nil {
		config.Logger.Error("Error creating zone: ", err)
		return false
	}

	config.Logger.Info("Response: ", body)

	var resp services.SessionDataResponse
	err = json.Unmarshal([]byte(body), &resp)
	if err != nil {
		config.Logger.Error("Error parsing response JSON: ", err)
		return false
	}

	return true
}

// UpdateZone updates an existing SDN zone in the Proxmox cluster
func UpdateZone(zoneName string, newZoneType string, trust bool) bool {
	sessionService, err := services.NewSessionService(config.Logger)
	if err != nil {
		config.Logger.Error("Error initializing session service: ", err)
		return false
	}

	sessionData, err := sessionService.ReadSessionFile()
	if err != nil {
		config.Logger.Error("Error reading session file: ", err)
		return false
	}

	uri := fmt.Sprintf("%s://%s:%d/api2/extjs/cluster/zones", sessionData.HttpScheme, sessionData.Server, sessionData.Port)

	httpService := services.NewHttpService(config.Logger, trust)

	payload := fmt.Sprintf("name=%s&type=%s", zoneName, newZoneType)
	headers := map[string]string{
		"Content-Type":        "application/x-www-form-urlencoded; charset=UTF-8",
		"CSRFPreventionToken": sessionData.Response.Data.CSRFPreventionToken,
	}

	body, err := httpService.Post(uri, payload, headers, nil)
	if err != nil {
		config.Logger.Error("Error updating zone: ", err)
		return false
	}

	config.Logger.Info("Response: ", body)

	var resp services.SessionDataResponse
	err = json.Unmarshal([]byte(body), &resp)
	if err != nil {
		config.Logger.Error("Error parsing response JSON: ", err)
		return false
	}

	return true
}

// DeleteZone deletes an existing SDN zone from the Proxmox cluster
func DeleteZone(zoneName string, trust bool) bool {
	sessionService, err := services.NewSessionService(config.Logger)
	if err != nil {
		config.Logger.Error("Error initializing session service: ", err)
		return false
	}

	sessionData, err := sessionService.ReadSessionFile()
	if err != nil {
		config.Logger.Error("Error reading session file: ", err)
		return false
	}

	uri := fmt.Sprintf("%s://%s:%d/api2/extjs/cluster/zones", sessionData.HttpScheme, sessionData.Server, sessionData.Port)

	httpService := services.NewHttpService(config.Logger, trust)

	payload := fmt.Sprintf("name=%s", zoneName)
	headers := map[string]string{
		"Content-Type":        "application/x-www-form-urlencoded; charset=UTF-8",
		"CSRFPreventionToken": sessionData.Response.Data.CSRFPreventionToken,
	}

	body, err := httpService.Post(uri, payload, headers, nil)
	if err != nil {
		config.Logger.Error("Error deleting zone: ", err)
		return false
	}

	config.Logger.Info("Response: ", body)

	var resp services.SessionDataResponse
	err = json.Unmarshal([]byte(body), &resp)
	if err != nil {
		config.Logger.Error("Error parsing response JSON: ", err)
		return false
	}

	return true
}
