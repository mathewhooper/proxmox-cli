package commands_test

import (
	"testing"

	"proxmox-cli/commands"
)

func TestLoginCommand(t *testing.T) {
	cmd := commands.LoginCommand()
	if cmd == nil {
		t.Fatalf("Expected LoginCommand to be initialized, got nil")
	}

	if cmd.Use != "login" {
		t.Errorf("Expected command use to be 'login', got '%s'", cmd.Use)
	}

	if cmd.Short != "Log in to a Proxmox server" {
		t.Errorf("Expected command short description to be 'Log in to a Proxmox server', got '%s'", cmd.Short)
	}
}

func TestValidateLoginCommand(t *testing.T) {
	cmd := commands.ValidateLoginCommand()
	if cmd == nil {
		t.Fatalf("Expected ValidateLoginCommand to be initialized, got nil")
	}

	if cmd.Use != "validate" {
		t.Errorf("Expected command use to be 'validate', got '%s'", cmd.Use)
	}

	if cmd.Short != "Validate the current session" {
		t.Errorf("Expected command short description to be 'Validate the current session', got '%s'", cmd.Short)
	}
}

func TestLoginCommandFlags(t *testing.T) {
	cmd := commands.LoginCommand()
	flags := []string{"server", "username", "port", "httpScheme", "trust", "show-log"}

	for _, flag := range flags {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("Expected flag '%s' to be defined", flag)
		}
	}
}

func TestValidateLoginCommandFlags(t *testing.T) {
	cmd := commands.ValidateLoginCommand()
	flags := []string{"trust", "show-log"}

	for _, flag := range flags {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("Expected flag '%s' to be defined", flag)
		}
	}
}
