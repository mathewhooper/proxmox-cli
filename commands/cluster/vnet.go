package cluster

import (
	"fmt"

	"proxmox-cli/config"
	"proxmox-cli/services"

	"encoding/json"

	"github.com/spf13/cobra"
)

// VnetCommand creates a new Cobra command for managing virtual networks (VNet).
func VnetCommand() *cobra.Command {
	var vnetCmd = &cobra.Command{
		Use:   "vnet",
		Short: "Manage virtual networks (VNet) in Proxmox",
	}

	// Add subcommands for VNet management
	vnetCmd.AddCommand(createVnetCommand())
	vnetCmd.AddCommand(deleteVnetCommand())
	vnetCmd.AddCommand(updateVnetCommand())

	return vnetCmd
}

// createVnetCommand creates a new Cobra command for creating a VNet.
func createVnetCommand() *cobra.Command {
	var vnetName string

	var createCmd = &cobra.Command{
		Use:   "create",
		Short: "Create a new virtual network (VNet)",
		Run: func(cmd *cobra.Command, args []string) {
			if vnetName == "" {
				fmt.Println("Error: VNet name is required")
				return
			}

			// Logic to create a VNet
			fmt.Printf("Creating VNet: Name=%s\n", vnetName)
		},
	}

	createCmd.Flags().StringVarP(&vnetName, "name", "n", "", "Name of the VNet to create")

	return createCmd
}

// deleteVnetCommand creates a new Cobra command for deleting a VNet.
func deleteVnetCommand() *cobra.Command {
	var vnetName string

	var deleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete an existing virtual network (VNet)",
		Run: func(cmd *cobra.Command, args []string) {
			if vnetName == "" {
				fmt.Println("Error: VNet name is required")
				return
			}

			// Logic to delete a VNet
			fmt.Printf("Deleting VNet: Name=%s\n", vnetName)
		},
	}

	deleteCmd.Flags().StringVarP(&vnetName, "name", "n", "", "Name of the VNet to delete")

	return deleteCmd
}

// updateVnetCommand creates a new Cobra command for updating a VNet.
func updateVnetCommand() *cobra.Command {
	var vnetName string
	var newConfig string

	var updateCmd = &cobra.Command{
		Use:   "update",
		Short: "Update the configuration of an existing virtual network (VNet)",
		Run: func(cmd *cobra.Command, args []string) {
			if vnetName == "" {
				fmt.Println("Error: VNet name is required")
				return
			}

			if newConfig == "" {
				fmt.Println("Error: New configuration is required")
				return
			}

			// Logic to update a VNet
			fmt.Printf("Updating VNet: Name=%s, NewConfig=%s\n", vnetName, newConfig)
		},
	}

	updateCmd.Flags().StringVarP(&vnetName, "name", "n", "", "Name of the VNet to update")
	updateCmd.Flags().StringVarP(&newConfig, "config", "c", "", "New configuration for the VNet")

	return updateCmd
}

func createVnet(vnetName string, trust bool) bool {
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

	uri := fmt.Sprintf("%s://%s:%d/api2/json/cluster/sdn", sessionData.HttpScheme, sessionData.Server, sessionData.Port)

	httpService := services.NewHttpService(config.Logger, trust)

	payload := fmt.Sprintf("name=%s", vnetName)
	headers := map[string]string{
		"Content-Type":        "application/x-www-form-urlencoded; charset=UTF-8",
		"CSRFPreventionToken": sessionData.Response.Data.CSRFPreventionToken,
	}

	body, err := httpService.Post(uri, payload, headers, nil)
	if err != nil {
		config.Logger.Error("Error creating vnet: ", err)
		return false
	}

	config.Logger.Info("Response: ", body)

	var resp services.SessionDataResponse
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		config.Logger.Error("Error parsing response JSON: ", err)
		return false
	}

	return true
}
