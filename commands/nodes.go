package commands

import (
	"fmt"
	"proxmox-cli/config"
	"proxmox-cli/services"

	"github.com/spf13/cobra"
)

// NodesCommand creates the parent command for node operations
func NodesCommand() *cobra.Command {
	var nodesCmd = &cobra.Command{
		Use:   "nodes",
		Short: "Manage Proxmox cluster nodes",
	}

	nodesCmd.AddCommand(ListNodesCommand())
	nodesCmd.AddCommand(NodeStatusCommand())
	nodesCmd.AddCommand(NodeVersionCommand())

	return nodesCmd
}

// ListNodesCommand lists all nodes in the cluster
func ListNodesCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "list",
		Short: "List all cluster nodes",
		Run: func(cmd *cobra.Command, args []string) {
			nodesService, err := services.NewNodesService(config.Logger, config.Trust)
			if err != nil {
				config.Logger.Error("Failed to initialize nodes service: ", err)
				fmt.Println("Error: Failed to initialize nodes service")
				return
			}

			nodes, err := nodesService.ListNodes()
			if err != nil {
				config.Logger.Error("Failed to list nodes: ", err)
				fmt.Println("Error: Failed to list nodes")
				return
			}

			if len(nodes) == 0 {
				fmt.Println("No nodes found")
				return
			}

			fmt.Printf("%-20s %-10s %-10s %-15s %-15s\n", "NODE", "STATUS", "CPU %", "MEMORY", "UPTIME")
			fmt.Println("================================================================================")
			for _, node := range nodes {
				cpuPercent := fmt.Sprintf("%.2f%%", node.CPU*100)
				memUsage := ""
				if node.MaxMem > 0 {
					memUsage = fmt.Sprintf("%.2f%%", float64(node.Mem)/float64(node.MaxMem)*100)
				}
				uptime := formatUptime(node.Uptime)

				fmt.Printf("%-20s %-10s %-10s %-15s %-15s\n",
					node.Node, node.Status, cpuPercent, memUsage, uptime)
			}
		},
	}

	return cmd
}

// NodeStatusCommand gets detailed status for a specific node
func NodeStatusCommand() *cobra.Command {
	var nodeName string

	var cmd = &cobra.Command{
		Use:   "status",
		Short: "Get detailed status for a specific node",
		Run: func(cmd *cobra.Command, args []string) {
			if nodeName == "" {
				fmt.Println("Error: node name is required")
				return
			}

			nodesService, err := services.NewNodesService(config.Logger, config.Trust)
			if err != nil {
				config.Logger.Error("Failed to initialize nodes service: ", err)
				fmt.Println("Error: Failed to initialize nodes service")
				return
			}

			status, err := nodesService.GetNodeStatus(nodeName)
			if err != nil {
				config.Logger.Error("Failed to get node status: ", err)
				fmt.Println("Error: Failed to get node status")
				return
			}

			fmt.Printf("Node Status for: %s\n", nodeName)
			fmt.Println("================================================================================")
			fmt.Printf("CPU Usage:       %.2f%%\n", status.CPU*100)
			fmt.Printf("CPU Model:       %s\n", status.CPUInfo.Model)
			fmt.Printf("CPU Cores:       %d\n", status.CPUInfo.CPUs)
			fmt.Printf("Memory Used:     %s / %s (%.2f%%)\n",
				formatBytes(status.Memory.Used), formatBytes(status.Memory.Total),
				float64(status.Memory.Used)/float64(status.Memory.Total)*100)
			fmt.Printf("Swap Used:       %s / %s\n",
				formatBytes(status.Swap.Used), formatBytes(status.Swap.Total))
			fmt.Printf("Root FS Used:    %s / %s (%.2f%%)\n",
				formatBytes(status.RootFS.Used), formatBytes(status.RootFS.Total),
				float64(status.RootFS.Used)/float64(status.RootFS.Total)*100)
			fmt.Printf("Uptime:          %s\n", formatUptime(status.Uptime))
			fmt.Printf("Load Average:    %.2f, %.2f, %.2f\n",
				status.LoadAvg[0], status.LoadAvg[1], status.LoadAvg[2])
			fmt.Printf("Kernel Version:  %s\n", status.KVersion)
			fmt.Printf("PVE Version:     %s\n", status.PVEVersion)
		},
	}

	cmd.Flags().StringVarP(&nodeName, "name", "n", "", "Name of the node")
	cmd.MarkFlagRequired("name")

	return cmd
}

// NodeVersionCommand gets version information for a specific node
func NodeVersionCommand() *cobra.Command {
	var nodeName string

	var cmd = &cobra.Command{
		Use:   "version",
		Short: "Get version information for a specific node",
		Run: func(cmd *cobra.Command, args []string) {
			if nodeName == "" {
				fmt.Println("Error: node name is required")
				return
			}

			nodesService, err := services.NewNodesService(config.Logger, config.Trust)
			if err != nil {
				config.Logger.Error("Failed to initialize nodes service: ", err)
				fmt.Println("Error: Failed to initialize nodes service")
				return
			}

			version, err := nodesService.GetNodeVersion(nodeName)
			if err != nil {
				config.Logger.Error("Failed to get node version: ", err)
				fmt.Println("Error: Failed to get node version")
				return
			}

			fmt.Printf("Node: %s\n", nodeName)
			fmt.Printf("Version: %s\n", version.Version)
			fmt.Printf("Release: %s\n", version.Release)
			fmt.Printf("Repo ID: %s\n", version.RepoID)
		},
	}

	cmd.Flags().StringVarP(&nodeName, "name", "n", "", "Name of the node")
	cmd.MarkFlagRequired("name")

	return cmd
}

// Helper function to format bytes into human-readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
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
