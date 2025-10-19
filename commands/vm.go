package commands

import (
	"fmt"
	"proxmox-cli/config"
	"proxmox-cli/services"

	"github.com/spf13/cobra"
)

// VMCommand creates the parent command for VM operations
func VMCommand() *cobra.Command {
	var vmCmd = &cobra.Command{
		Use:   "vm",
		Short: "Manage Proxmox virtual machines (QEMU)",
	}

	vmCmd.AddCommand(ListVMsCommand())
	vmCmd.AddCommand(VMStatusCommand())
	vmCmd.AddCommand(StartVMCommand())
	vmCmd.AddCommand(StopVMCommand())
	vmCmd.AddCommand(ShutdownVMCommand())
	vmCmd.AddCommand(RebootVMCommand())
	vmCmd.AddCommand(ResetVMCommand())
	vmCmd.AddCommand(SuspendVMCommand())
	vmCmd.AddCommand(ResumeVMCommand())
	vmCmd.AddCommand(DeleteVMCommand())

	return vmCmd
}

// ListVMsCommand lists all VMs on a node
func ListVMsCommand() *cobra.Command {
	var nodeName string

	var cmd = &cobra.Command{
		Use:   "list",
		Short: "List all virtual machines on a node",
		Run: func(cmd *cobra.Command, args []string) {
			if nodeName == "" {
				fmt.Println("Error: node name is required")
				return
			}

			vmService, err := services.NewVMService(config.Logger, config.Trust)
			if err != nil {
				config.Logger.Error("Failed to initialize VM service: ", err)
				fmt.Println("Error: Failed to initialize VM service")
				return
			}

			vms, err := vmService.ListVMs(nodeName)
			if err != nil {
				config.Logger.Error("Failed to list VMs: ", err)
				fmt.Println("Error: Failed to list VMs")
				return
			}

			if len(vms) == 0 {
				fmt.Printf("No VMs found on node: %s\n", nodeName)
				return
			}

			fmt.Printf("%-8s %-20s %-10s %-10s %-15s %-15s\n", "VMID", "NAME", "STATUS", "CPU %", "MEMORY", "UPTIME")
			fmt.Println("========================================================================================")
			for _, vm := range vms {
				cpuPercent := fmt.Sprintf("%.2f%%", vm.CPU*100)
				memUsage := ""
				if vm.MaxMem > 0 {
					memUsage = fmt.Sprintf("%.2f%%", float64(vm.Mem)/float64(vm.MaxMem)*100)
				}
				uptime := formatUptime(vm.Uptime)

				fmt.Printf("%-8d %-20s %-10s %-10s %-15s %-15s\n",
					vm.VMID, vm.Name, vm.Status, cpuPercent, memUsage, uptime)
			}
		},
	}

	cmd.Flags().StringVarP(&nodeName, "node", "n", "", "Name of the node")
	cmd.MarkFlagRequired("node")

	return cmd
}

// VMStatusCommand gets the status of a specific VM
func VMStatusCommand() *cobra.Command {
	var nodeName string
	var vmid int

	var cmd = &cobra.Command{
		Use:   "status",
		Short: "Get the status of a specific VM",
		Run: func(cmd *cobra.Command, args []string) {
			if nodeName == "" {
				fmt.Println("Error: node name is required")
				return
			}
			if vmid == 0 {
				fmt.Println("Error: VM ID is required")
				return
			}

			vmService, err := services.NewVMService(config.Logger, config.Trust)
			if err != nil {
				config.Logger.Error("Failed to initialize VM service: ", err)
				fmt.Println("Error: Failed to initialize VM service")
				return
			}

			status, err := vmService.GetVMStatus(nodeName, vmid)
			if err != nil {
				config.Logger.Error("Failed to get VM status: ", err)
				fmt.Println("Error: Failed to get VM status")
				return
			}

			fmt.Printf("VM Status for VMID: %d\n", vmid)
			fmt.Println("================================================================================")
			fmt.Printf("Name:            %s\n", status.Name)
			fmt.Printf("Status:          %s\n", status.Status)
			fmt.Printf("QMP Status:      %s\n", status.QMPStatus)
			fmt.Printf("CPU Usage:       %.2f%%\n", status.CPU*100)
			fmt.Printf("CPU Cores:       %d\n", status.CPUs)
			if status.MaxMem > 0 {
				fmt.Printf("Memory Used:     %s / %s (%.2f%%)\n",
					formatBytes(status.Mem), formatBytes(status.MaxMem),
					float64(status.Mem)/float64(status.MaxMem)*100)
			}
			fmt.Printf("Uptime:          %s\n", formatUptime(status.Uptime))
		},
	}

	cmd.Flags().StringVarP(&nodeName, "node", "n", "", "Name of the node")
	cmd.Flags().IntVarP(&vmid, "vmid", "i", 0, "VM ID")
	cmd.MarkFlagRequired("node")
	cmd.MarkFlagRequired("vmid")

	return cmd
}

// StartVMCommand starts a VM
func StartVMCommand() *cobra.Command {
	var nodeName string
	var vmid int

	var cmd = &cobra.Command{
		Use:   "start",
		Short: "Start a virtual machine",
		Run: func(cmd *cobra.Command, args []string) {
			if nodeName == "" || vmid == 0 {
				fmt.Println("Error: node name and VM ID are required")
				return
			}

			vmService, err := services.NewVMService(config.Logger, config.Trust)
			if err != nil {
				config.Logger.Error("Failed to initialize VM service: ", err)
				fmt.Println("Error: Failed to initialize VM service")
				return
			}

			taskID, err := vmService.StartVM(nodeName, vmid)
			if err != nil {
				config.Logger.Error("Failed to start VM: ", err)
				fmt.Println("Error: Failed to start VM")
				return
			}

			fmt.Printf("VM %d start initiated. Task ID: %s\n", vmid, taskID)
		},
	}

	cmd.Flags().StringVarP(&nodeName, "node", "n", "", "Name of the node")
	cmd.Flags().IntVarP(&vmid, "vmid", "i", 0, "VM ID")
	cmd.MarkFlagRequired("node")
	cmd.MarkFlagRequired("vmid")

	return cmd
}

// StopVMCommand stops a VM
func StopVMCommand() *cobra.Command {
	var nodeName string
	var vmid int

	var cmd = &cobra.Command{
		Use:   "stop",
		Short: "Stop a virtual machine",
		Run: func(cmd *cobra.Command, args []string) {
			if nodeName == "" || vmid == 0 {
				fmt.Println("Error: node name and VM ID are required")
				return
			}

			vmService, err := services.NewVMService(config.Logger, config.Trust)
			if err != nil {
				config.Logger.Error("Failed to initialize VM service: ", err)
				fmt.Println("Error: Failed to initialize VM service")
				return
			}

			taskID, err := vmService.StopVM(nodeName, vmid)
			if err != nil {
				config.Logger.Error("Failed to stop VM: ", err)
				fmt.Println("Error: Failed to stop VM")
				return
			}

			fmt.Printf("VM %d stop initiated. Task ID: %s\n", vmid, taskID)
		},
	}

	cmd.Flags().StringVarP(&nodeName, "node", "n", "", "Name of the node")
	cmd.Flags().IntVarP(&vmid, "vmid", "i", 0, "VM ID")
	cmd.MarkFlagRequired("node")
	cmd.MarkFlagRequired("vmid")

	return cmd
}

// ShutdownVMCommand gracefully shuts down a VM
func ShutdownVMCommand() *cobra.Command {
	var nodeName string
	var vmid int

	var cmd = &cobra.Command{
		Use:   "shutdown",
		Short: "Gracefully shutdown a virtual machine",
		Run: func(cmd *cobra.Command, args []string) {
			if nodeName == "" || vmid == 0 {
				fmt.Println("Error: node name and VM ID are required")
				return
			}

			vmService, err := services.NewVMService(config.Logger, config.Trust)
			if err != nil {
				config.Logger.Error("Failed to initialize VM service: ", err)
				fmt.Println("Error: Failed to initialize VM service")
				return
			}

			taskID, err := vmService.ShutdownVM(nodeName, vmid)
			if err != nil {
				config.Logger.Error("Failed to shutdown VM: ", err)
				fmt.Println("Error: Failed to shutdown VM")
				return
			}

			fmt.Printf("VM %d shutdown initiated. Task ID: %s\n", vmid, taskID)
		},
	}

	cmd.Flags().StringVarP(&nodeName, "node", "n", "", "Name of the node")
	cmd.Flags().IntVarP(&vmid, "vmid", "i", 0, "VM ID")
	cmd.MarkFlagRequired("node")
	cmd.MarkFlagRequired("vmid")

	return cmd
}

// RebootVMCommand reboots a VM
func RebootVMCommand() *cobra.Command {
	var nodeName string
	var vmid int

	var cmd = &cobra.Command{
		Use:   "reboot",
		Short: "Reboot a virtual machine",
		Run: func(cmd *cobra.Command, args []string) {
			if nodeName == "" || vmid == 0 {
				fmt.Println("Error: node name and VM ID are required")
				return
			}

			vmService, err := services.NewVMService(config.Logger, config.Trust)
			if err != nil {
				config.Logger.Error("Failed to initialize VM service: ", err)
				fmt.Println("Error: Failed to initialize VM service")
				return
			}

			taskID, err := vmService.RebootVM(nodeName, vmid)
			if err != nil {
				config.Logger.Error("Failed to reboot VM: ", err)
				fmt.Println("Error: Failed to reboot VM")
				return
			}

			fmt.Printf("VM %d reboot initiated. Task ID: %s\n", vmid, taskID)
		},
	}

	cmd.Flags().StringVarP(&nodeName, "node", "n", "", "Name of the node")
	cmd.Flags().IntVarP(&vmid, "vmid", "i", 0, "VM ID")
	cmd.MarkFlagRequired("node")
	cmd.MarkFlagRequired("vmid")

	return cmd
}

// ResetVMCommand resets a VM
func ResetVMCommand() *cobra.Command {
	var nodeName string
	var vmid int

	var cmd = &cobra.Command{
		Use:   "reset",
		Short: "Reset a virtual machine",
		Run: func(cmd *cobra.Command, args []string) {
			if nodeName == "" || vmid == 0 {
				fmt.Println("Error: node name and VM ID are required")
				return
			}

			vmService, err := services.NewVMService(config.Logger, config.Trust)
			if err != nil {
				config.Logger.Error("Failed to initialize VM service: ", err)
				fmt.Println("Error: Failed to initialize VM service")
				return
			}

			taskID, err := vmService.ResetVM(nodeName, vmid)
			if err != nil {
				config.Logger.Error("Failed to reset VM: ", err)
				fmt.Println("Error: Failed to reset VM")
				return
			}

			fmt.Printf("VM %d reset initiated. Task ID: %s\n", vmid, taskID)
		},
	}

	cmd.Flags().StringVarP(&nodeName, "node", "n", "", "Name of the node")
	cmd.Flags().IntVarP(&vmid, "vmid", "i", 0, "VM ID")
	cmd.MarkFlagRequired("node")
	cmd.MarkFlagRequired("vmid")

	return cmd
}

// SuspendVMCommand suspends a VM
func SuspendVMCommand() *cobra.Command {
	var nodeName string
	var vmid int

	var cmd = &cobra.Command{
		Use:   "suspend",
		Short: "Suspend a virtual machine",
		Run: func(cmd *cobra.Command, args []string) {
			if nodeName == "" || vmid == 0 {
				fmt.Println("Error: node name and VM ID are required")
				return
			}

			vmService, err := services.NewVMService(config.Logger, config.Trust)
			if err != nil {
				config.Logger.Error("Failed to initialize VM service: ", err)
				fmt.Println("Error: Failed to initialize VM service")
				return
			}

			taskID, err := vmService.SuspendVM(nodeName, vmid)
			if err != nil {
				config.Logger.Error("Failed to suspend VM: ", err)
				fmt.Println("Error: Failed to suspend VM")
				return
			}

			fmt.Printf("VM %d suspend initiated. Task ID: %s\n", vmid, taskID)
		},
	}

	cmd.Flags().StringVarP(&nodeName, "node", "n", "", "Name of the node")
	cmd.Flags().IntVarP(&vmid, "vmid", "i", 0, "VM ID")
	cmd.MarkFlagRequired("node")
	cmd.MarkFlagRequired("vmid")

	return cmd
}

// ResumeVMCommand resumes a suspended VM
func ResumeVMCommand() *cobra.Command {
	var nodeName string
	var vmid int

	var cmd = &cobra.Command{
		Use:   "resume",
		Short: "Resume a suspended virtual machine",
		Run: func(cmd *cobra.Command, args []string) {
			if nodeName == "" || vmid == 0 {
				fmt.Println("Error: node name and VM ID are required")
				return
			}

			vmService, err := services.NewVMService(config.Logger, config.Trust)
			if err != nil {
				config.Logger.Error("Failed to initialize VM service: ", err)
				fmt.Println("Error: Failed to initialize VM service")
				return
			}

			taskID, err := vmService.ResumeVM(nodeName, vmid)
			if err != nil {
				config.Logger.Error("Failed to resume VM: ", err)
				fmt.Println("Error: Failed to resume VM")
				return
			}

			fmt.Printf("VM %d resume initiated. Task ID: %s\n", vmid, taskID)
		},
	}

	cmd.Flags().StringVarP(&nodeName, "node", "n", "", "Name of the node")
	cmd.Flags().IntVarP(&vmid, "vmid", "i", 0, "VM ID")
	cmd.MarkFlagRequired("node")
	cmd.MarkFlagRequired("vmid")

	return cmd
}

// DeleteVMCommand deletes a VM
func DeleteVMCommand() *cobra.Command {
	var nodeName string
	var vmid int

	var cmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete a virtual machine",
		Run: func(cmd *cobra.Command, args []string) {
			if nodeName == "" || vmid == 0 {
				fmt.Println("Error: node name and VM ID are required")
				return
			}

			vmService, err := services.NewVMService(config.Logger, config.Trust)
			if err != nil {
				config.Logger.Error("Failed to initialize VM service: ", err)
				fmt.Println("Error: Failed to initialize VM service")
				return
			}

			taskID, err := vmService.DeleteVM(nodeName, vmid)
			if err != nil {
				config.Logger.Error("Failed to delete VM: ", err)
				fmt.Println("Error: Failed to delete VM")
				return
			}

			fmt.Printf("VM %d deletion initiated. Task ID: %s\n", vmid, taskID)
		},
	}

	cmd.Flags().StringVarP(&nodeName, "node", "n", "", "Name of the node")
	cmd.Flags().IntVarP(&vmid, "vmid", "i", 0, "VM ID")
	cmd.MarkFlagRequired("node")
	cmd.MarkFlagRequired("vmid")

	return cmd
}
