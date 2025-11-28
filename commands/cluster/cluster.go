package cluster

import (
	"fmt"
	"proxmox-cli/config"
	"proxmox-cli/services"

	"github.com/spf13/cobra"
)

// ClusterCommand creates a new Cobra command for cluster-related operations.
//
//nolint:revive
func ClusterCommand() *cobra.Command {
	var clusterCmd = &cobra.Command{
		Use:   "cluster",
		Short: "Manage Proxmox clusters",
	}

	// Add subcommands here
	clusterCmd.AddCommand(SDNCommand())
	clusterCmd.AddCommand(ResourcesCommand())
	clusterCmd.AddCommand(StatusCommand())

	return clusterCmd
}

// ResourcesCommand lists all cluster resources
func ResourcesCommand() *cobra.Command {
	var resourceType string

	var cmd = &cobra.Command{
		Use:   "resources",
		Short: "List all cluster resources",
		Run: func(cmd *cobra.Command, args []string) {
			clusterService, err := services.NewClusterService(config.Logger, config.Trust)
			if err != nil {
				config.Logger.Error("Failed to initialize cluster service: ", err)
				fmt.Println("Error: Failed to initialize cluster service")
				return
			}

			resources, err := clusterService.ListResources()
			if err != nil {
				config.Logger.Error("Failed to list cluster resources: ", err)
				fmt.Println("Error: Failed to list cluster resources")
				return
			}

			// Filter by type if specified
			var filteredResources []services.ClusterResource
			if resourceType != "" {
				for _, r := range resources {
					if r.Type == resourceType {
						filteredResources = append(filteredResources, r)
					}
				}
			} else {
				filteredResources = resources
			}

			if len(filteredResources) == 0 {
				if resourceType != "" {
					fmt.Printf("No resources found of type: %s\n", resourceType)
				} else {
					fmt.Println("No resources found")
				}
				return
			}

			fmt.Printf("%-8s %-20s %-20s %-10s %-10s %-15s\n", "TYPE", "ID", "NAME", "NODE", "STATUS", "UPTIME")
			fmt.Println("========================================================================================")
			for _, resource := range filteredResources {
				uptime := formatUptime(resource.Uptime)
				fmt.Printf("%-8s %-20s %-20s %-10s %-10s %-15s\n",
					resource.Type, resource.ID, resource.Name, resource.Node, resource.Status, uptime)
			}
		},
	}

	cmd.Flags().StringVarP(&resourceType, "type", "t", "", "Filter by resource type (vm, node, storage, etc.)")

	return cmd
}

// StatusCommand gets cluster status
func StatusCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "status",
		Short: "Get cluster status",
		Run: func(cmd *cobra.Command, args []string) {
			clusterService, err := services.NewClusterService(config.Logger, config.Trust)
			if err != nil {
				config.Logger.Error("Failed to initialize cluster service: ", err)
				fmt.Println("Error: Failed to initialize cluster service")
				return
			}

			statuses, err := clusterService.GetStatus()
			if err != nil {
				config.Logger.Error("Failed to get cluster status: ", err)
				fmt.Println("Error: Failed to get cluster status")
				return
			}

			if len(statuses) == 0 {
				fmt.Println("No cluster status found")
				return
			}

			fmt.Println("Cluster Status:")
			fmt.Println("================================================================================")
			for _, status := range statuses {
				fmt.Printf("Type: %s\n", status.Type)
				if status.Name != "" {
					fmt.Printf("Name: %s\n", status.Name)
				}
				if status.Type == "cluster" {
					fmt.Printf("Nodes: %d\n", status.Nodes)
					fmt.Printf("Quorate: %d\n", status.Quorate)
					fmt.Printf("Version: %d\n", status.Version)
				}
				if status.IP != "" {
					fmt.Printf("IP: %s\n", status.IP)
				}
				if status.Online == 1 {
					fmt.Println("Online: Yes")
				}
				fmt.Println("---")
			}
		},
	}

	return cmd
}

// Helper function to format uptime into human-readable format
func formatUptime(seconds int64) string {
	if seconds == 0 {
		return "N/A"
	}

	days := seconds / 86400
	hours := (seconds % 86400) / 3600
	minutes := (seconds % 3600) / 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	} else if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	} else {
		return fmt.Sprintf("%dm", minutes)
	}
}
