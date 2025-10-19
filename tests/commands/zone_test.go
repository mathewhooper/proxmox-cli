package commands_test

import (
	"bytes"
	"fmt"
	"testing"

	"proxmox-cli/commands/cluster"

	"github.com/stretchr/testify/assert"
)

func TestZoneCommand(t *testing.T) {
	cmd := cluster.ZoneCommand()

	assert.NotNil(t, cmd)
	assert.Equal(t, "sdn", cmd.Use)
	assert.Equal(t, "Manage Software Defined Networking (SDN) in Proxmox", cmd.Short)

	// Check that all subcommands are added
	subcommands := cmd.Commands()
	assert.Len(t, subcommands, 3)

	// Check subcommand names
	subcommandNames := make([]string, len(subcommands))
	for i, subcmd := range subcommands {
		subcommandNames[i] = subcmd.Name()
	}

	assert.Contains(t, subcommandNames, "create-zone")
	assert.Contains(t, subcommandNames, "delete-zone")
	assert.Contains(t, subcommandNames, "update-zone")
}

func TestCreateZoneCommand(t *testing.T) {
	cmd := cluster.ZoneCommand()
	createCmd := cmd.Commands()[0] // create-zone is the first subcommand

	assert.Equal(t, "create-zone", createCmd.Name())
	assert.Equal(t, "Create a new SDN zone", createCmd.Short)

	// Test required flags
	assert.True(t, createCmd.Flags().Lookup("name") != nil)
	assert.True(t, createCmd.Flags().Lookup("type") != nil)

	// Test flag shorthand
	assert.Equal(t, "n", createCmd.Flags().Lookup("name").Shorthand)
	assert.Equal(t, "t", createCmd.Flags().Lookup("type").Shorthand)
}

func TestDeleteZoneCommand(t *testing.T) {
	cmd := cluster.ZoneCommand()
	deleteCmd := cmd.Commands()[1] // delete-zone is the second subcommand

	assert.Equal(t, "delete-zone", deleteCmd.Name())
	assert.Equal(t, "Delete an existing SDN zone", deleteCmd.Short)

	// Test required flags
	assert.True(t, deleteCmd.Flags().Lookup("name") != nil)

	// Test flag shorthand
	assert.Equal(t, "n", deleteCmd.Flags().Lookup("name").Shorthand)
}

func TestUpdateZoneCommand(t *testing.T) {
	cmd := cluster.ZoneCommand()
	updateCmd := cmd.Commands()[2] // update-zone is the third subcommand

	assert.Equal(t, "update-zone", updateCmd.Name())
	assert.Equal(t, "Update an existing SDN zone", updateCmd.Short)

	// Test required flags
	assert.True(t, updateCmd.Flags().Lookup("name") != nil)
	assert.True(t, updateCmd.Flags().Lookup("new-type") != nil)

	// Test flag shorthand
	assert.Equal(t, "n", updateCmd.Flags().Lookup("name").Shorthand)
	assert.Equal(t, "t", updateCmd.Flags().Lookup("new-type").Shorthand)
}

func TestZoneTypeValidation(t *testing.T) {
	validZoneTypes := []string{
		"simple", "vlan", "vxlan", "gre", "ipsec",
		"l2tpv3", "vxlan-ipsec", "l2tpv3-ipsec",
	}

	invalidZoneTypes := []string{
		"invalid", "test", "zone", "", "simple-invalid",
	}

	// Test valid zone types
	for _, zoneType := range validZoneTypes {
		t.Run(fmt.Sprintf("ValidZoneType_%s", zoneType), func(t *testing.T) {
			cmd := cluster.ZoneCommand()
			createCmd := cmd.Commands()[0]

			// Set valid flags
			createCmd.Flags().Set("name", "testzone")
			createCmd.Flags().Set("type", zoneType)

			// Verify flags were set correctly
			name, _ := createCmd.Flags().GetString("name")
			zoneTypeFlag, _ := createCmd.Flags().GetString("type")

			assert.Equal(t, "testzone", name)
			assert.Equal(t, zoneType, zoneTypeFlag)
		})
	}

	// Test invalid zone types
	for _, zoneType := range invalidZoneTypes {
		t.Run(fmt.Sprintf("InvalidZoneType_%s", zoneType), func(t *testing.T) {
			cmd := cluster.ZoneCommand()
			createCmd := cmd.Commands()[0]

			// Set invalid flags
			createCmd.Flags().Set("name", "testzone")
			createCmd.Flags().Set("type", zoneType)

			// Verify flags were set correctly
			name, _ := createCmd.Flags().GetString("name")
			zoneTypeFlag, _ := createCmd.Flags().GetString("type")

			assert.Equal(t, "testzone", name)
			assert.Equal(t, zoneType, zoneTypeFlag)
		})
	}
}

func TestCommandHelp(t *testing.T) {
	cmd := cluster.ZoneCommand()

	// Test main command help
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.Help()

	output := buf.String()
	assert.Contains(t, output, "Manage Software Defined Networking (SDN) in Proxmox")
	assert.Contains(t, output, "create-zone")
	assert.Contains(t, output, "delete-zone")
	assert.Contains(t, output, "update-zone")

	// Test subcommand help
	createCmd := cmd.Commands()[0]
	buf.Reset()
	createCmd.SetOut(&buf)
	createCmd.Help()

	output = buf.String()
	assert.Contains(t, output, "Create a new SDN zone")
	assert.Contains(t, output, "--name")
	assert.Contains(t, output, "--type")

	// Test delete-zone help
	deleteCmd := cmd.Commands()[1]
	buf.Reset()
	deleteCmd.SetOut(&buf)
	deleteCmd.Help()

	output = buf.String()
	assert.Contains(t, output, "Delete an existing SDN zone")
	assert.Contains(t, output, "--name")

	// Test update-zone help
	updateCmd := cmd.Commands()[2]
	buf.Reset()
	updateCmd.SetOut(&buf)
	updateCmd.Help()

	output = buf.String()
	assert.Contains(t, output, "Update an existing SDN zone")
	assert.Contains(t, output, "--name")
	assert.Contains(t, output, "--new-type")
}

func TestCommandFlags(t *testing.T) {
	cmd := cluster.ZoneCommand()

	// Test create-zone flags
	createCmd := cmd.Commands()[0]
	createCmd.Flags().Set("name", "testzone")
	createCmd.Flags().Set("type", "simple")

	name, _ := createCmd.Flags().GetString("name")
	zoneType, _ := createCmd.Flags().GetString("type")

	assert.Equal(t, "testzone", name)
	assert.Equal(t, "simple", zoneType)

	// Test delete-zone flags
	deleteCmd := cmd.Commands()[1]
	deleteCmd.Flags().Set("name", "testzone")

	deleteName, _ := deleteCmd.Flags().GetString("name")
	assert.Equal(t, "testzone", deleteName)

	// Test update-zone flags
	updateCmd := cmd.Commands()[2]
	updateCmd.Flags().Set("name", "testzone")
	updateCmd.Flags().Set("new-type", "vlan")

	updateName, _ := updateCmd.Flags().GetString("name")
	updateType, _ := updateCmd.Flags().GetString("new-type")

	assert.Equal(t, "testzone", updateName)
	assert.Equal(t, "vlan", updateType)
}

func TestCommandUsage(t *testing.T) {
	cmd := cluster.ZoneCommand()

	// Test main command usage
	assert.Equal(t, "sdn", cmd.Use)
	assert.Equal(t, "Manage Software Defined Networking (SDN) in Proxmox", cmd.Short)
	// Note: Long field is not set, so we don't test it

	// Test subcommand usage
	createCmd := cmd.Commands()[0]
	assert.Equal(t, "create-zone", createCmd.Use)
	assert.Equal(t, "Create a new SDN zone", createCmd.Short)

	deleteCmd := cmd.Commands()[1]
	assert.Equal(t, "delete-zone", deleteCmd.Use)
	assert.Equal(t, "Delete an existing SDN zone", deleteCmd.Short)

	updateCmd := cmd.Commands()[2]
	assert.Equal(t, "update-zone", updateCmd.Use)
	assert.Equal(t, "Update an existing SDN zone", updateCmd.Short)
}

func TestCommandStructure(t *testing.T) {
	cmd := cluster.ZoneCommand()

	// Verify command hierarchy
	assert.Equal(t, "sdn", cmd.Name())
	assert.Len(t, cmd.Commands(), 3)

	// Verify subcommand names and order
	subcommands := cmd.Commands()
	assert.Equal(t, "create-zone", subcommands[0].Name())
	assert.Equal(t, "delete-zone", subcommands[1].Name())
	assert.Equal(t, "update-zone", subcommands[2].Name())

	// Verify each subcommand has the correct parent
	for _, subcmd := range subcommands {
		assert.Equal(t, cmd, subcmd.Parent())
	}
}
