package commands

import (
	"fmt"
	"os"

	"proxmox-cli/config"
	"proxmox-cli/services"

	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// LoginCommand creates a new Cobra command for logging in to a Proxmox server.
//
// Returns:
// - *cobra.Command: The login command.
func LoginCommand() *cobra.Command {
	var server string
	var username string
	var httpScheme string
	var port int
	var logLevel bool

	var loginCmd = &cobra.Command{
		Use:   "login",
		Short: "Log in to a Proxmox server",
		Run: func(cmd *cobra.Command, args []string) {
			config.Logger.Info("Logging in to Proxmox server...")
			if logLevel {
				config.SetLogLevel(logrus.InfoLevel)
			}
			fmt.Print("Enter Password: ")
			passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
			fmt.Println()
			if err != nil {
				config.Logger.Error("Error reading password: ", err)
				return
			}
			password := string(passwordBytes)
			authService := services.NewAuthService(config.Logger, config.Trust)
			err = authService.LoginToProxmox(server, port, httpScheme, username, password)
			if err != nil {
				config.Logger.Error("Login failed: ", err)
			}
		},
	}

	loginCmd.Flags().StringVarP(&server, "server", "s", "", "Proxmox server URL")
	loginCmd.Flags().StringVarP(&username, "username", "u", "", "Username for Proxmox")
	loginCmd.Flags().IntVarP(&port, "port", "P", 8006, "Proxmox server port")
	loginCmd.Flags().StringVarP(&httpScheme, "httpScheme", "S", "https", "HTTP scheme (http or https)")
	loginCmd.Flags().BoolVarP(&logLevel, "show-log", "l", false, "Set the log level to error")

	//nolint:errcheck // Flag is defined above, so this cannot fail
	_ = loginCmd.MarkFlagRequired("server")
	//nolint:errcheck // Flag is defined above, so this cannot fail
	_ = loginCmd.MarkFlagRequired("username")

	return loginCmd
}

// ValidateLoginCommand creates a new Cobra command for validating the current session.
//
// Returns:
// - *cobra.Command: The validate command.
func ValidateLoginCommand() *cobra.Command {
	var logLevel bool

	var validateCmd = &cobra.Command{
		Use:   "validate",
		Short: "Validate the current session",
	}

	validateCmd.Flags().BoolVarP(&logLevel, "show-log", "l", false, "Set the log level to error")

	validateCmd.Run = func(cmd *cobra.Command, args []string) {
		config.Logger.Info("Validating session...")
		if logLevel {
			config.SetLogLevel(logrus.InfoLevel)
		}

		authService := services.NewAuthService(config.Logger, config.Trust)
		if authService.ValidateSession() {
			fmt.Println("Session is valid.")
		} else {
			fmt.Println("Session is invalid.")
		}
	}

	return validateCmd
}
