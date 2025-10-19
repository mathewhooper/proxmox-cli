package commands

import (
	"fmt"
	"proxmox-cli/config"
	"proxmox-cli/services"

	"github.com/spf13/cobra"
)

// StorageCommand creates the parent command for storage operations
func StorageCommand() *cobra.Command {
	var storageCmd = &cobra.Command{
		Use:   "storage",
		Short: "Manage Proxmox storage",
	}

	storageCmd.AddCommand(ListStorageCommand())
	storageCmd.AddCommand(StorageContentCommand())

	return storageCmd
}

// ListStorageCommand lists all storage
func ListStorageCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "list",
		Short: "List all storage",
		Run: func(cmd *cobra.Command, args []string) {
			storageService, err := services.NewStorageService(config.Logger, config.Trust)
			if err != nil {
				config.Logger.Error("Failed to initialize storage service: ", err)
				fmt.Println("Error: Failed to initialize storage service")
				return
			}

			storages, err := storageService.ListStorage()
			if err != nil {
				config.Logger.Error("Failed to list storage: ", err)
				fmt.Println("Error: Failed to list storage")
				return
			}

			if len(storages) == 0 {
				fmt.Println("No storage found")
				return
			}

			fmt.Printf("%-20s %-15s %-10s %-10s %-30s\n", "STORAGE", "TYPE", "SHARED", "ACTIVE", "CONTENT")
			fmt.Println("==========================================================================================")
			for _, storage := range storages {
				shared := "No"
				if storage.Shared == 1 {
					shared = "Yes"
				}
				active := "No"
				if storage.Active == 1 {
					active = "Yes"
				}

				fmt.Printf("%-20s %-15s %-10s %-10s %-30s\n",
					storage.Storage, storage.Type, shared, active, storage.Content)
			}
		},
	}

	return cmd
}

// StorageContentCommand lists content of a specific storage
func StorageContentCommand() *cobra.Command {
	var nodeName string
	var storageName string

	var cmd = &cobra.Command{
		Use:   "content",
		Short: "List content of a specific storage",
		Run: func(cmd *cobra.Command, args []string) {
			if nodeName == "" || storageName == "" {
				fmt.Println("Error: node name and storage name are required")
				return
			}

			storageService, err := services.NewStorageService(config.Logger, config.Trust)
			if err != nil {
				config.Logger.Error("Failed to initialize storage service: ", err)
				fmt.Println("Error: Failed to initialize storage service")
				return
			}

			contents, err := storageService.ListStorageContent(nodeName, storageName)
			if err != nil {
				config.Logger.Error("Failed to list storage content: ", err)
				fmt.Println("Error: Failed to list storage content")
				return
			}

			if len(contents) == 0 {
				fmt.Printf("No content found in storage: %s\n", storageName)
				return
			}

			fmt.Printf("%-50s %-15s %-15s %-8s\n", "VOLUME ID", "FORMAT", "SIZE", "VMID")
			fmt.Println("==========================================================================================")
			for _, content := range contents {
				vmid := "-"
				if content.VMID != 0 {
					vmid = fmt.Sprintf("%d", content.VMID)
				}

				fmt.Printf("%-50s %-15s %-15s %-8s\n",
					content.VolID, content.Format, formatBytes(content.Size), vmid)
			}
		},
	}

	cmd.Flags().StringVarP(&nodeName, "node", "n", "", "Name of the node")
	cmd.Flags().StringVarP(&storageName, "storage", "s", "", "Name of the storage")
	cmd.MarkFlagRequired("node")
	cmd.MarkFlagRequired("storage")

	return cmd
}
