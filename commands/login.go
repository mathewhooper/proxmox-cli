package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"proxmox-cli/config"
	"proxmox-cli/helpers"

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
	var trust bool
	var logLevel bool

	var loginCmd = &cobra.Command{
		Use:   "login",
		Short: "Log in to a Proxmox server",
		Run: func(cmd *cobra.Command, args []string) {
            config.Logger.Info("Logging in to Proxmox server...")
            if logLevel  {
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
			loginToProxmox(server, port, httpScheme, username, password, trust)
		},
	}

	loginCmd.Flags().StringVarP(&server, "server", "s", "", "Proxmox server URL")
	loginCmd.Flags().StringVarP(&username, "username", "u", "", "Username for Proxmox")
	loginCmd.Flags().IntVarP(&port, "port", "P", 8006, "Proxmox server port")
	loginCmd.Flags().StringVarP(&httpScheme, "httpScheme", "S", "https", "HTTP scheme (http or https)")
	loginCmd.Flags().BoolVarP(&trust, "trust", "t", false, "Trust SSL certificates")
    loginCmd.Flags().BoolVarP(&logLevel, "show-log", "l", false, "Set the log level to error")

	loginCmd.MarkFlagRequired("server")
	loginCmd.MarkFlagRequired("username")

	return loginCmd
}

// ValidateLoginCommand creates a new Cobra command for validating the current session.
//
// Returns:
// - *cobra.Command: The validate command.
func ValidateLoginCommand() *cobra.Command {
	var trust bool
	var logLevel bool

    var validateCmd = &cobra.Command{
        Use:   "validate",
        Short: "Validate the current session",
    }

    validateCmd.Flags().BoolVarP(&trust, "trust", "t", false, "Trust SSL certificates")
    validateCmd.Flags().BoolVarP(&logLevel, "show-log", "l", false, "Set the log level to error")

    validateCmd.Run = func(cmd *cobra.Command, args []string) {
        config.Logger.Info("Validating session...")
        if logLevel  {
            config.SetLogLevel(logrus.InfoLevel)
        }

        if validateSession(trust) {
            fmt.Println("Session is valid.")
        } else {
            fmt.Println("Session is invalid.")
        }
    }
    
    return validateCmd
}

// loginToProxmox handles the login process to a Proxmox server.
//
// Parameters:
// - server: The Proxmox server URL.
// - port: The Proxmox server port.
// - httpScheme: The HTTP scheme (http or https).
// - username: The username for Proxmox.
// - password: The password for Proxmox.
// - trust: A boolean indicating whether to trust SSL certificates.
func loginToProxmox(server string, port int, httpScheme string, username string, password string, trust bool) {
    uri := fmt.Sprintf("%s://%s:%d/api2/json/access/ticket", httpScheme, server, port)
    client := helpers.CreateHTTPClient(trust)

    payload := fmt.Sprintf("username=%s&password=%s&realm=pam&new-format=1", username, password)
    headers := map[string]string{
        "Content-Type": "application/x-www-form-urlencoded; charset=UTF-8",
    }

    req, err := helpers.CreateHTTPRequest("POST", uri, payload, headers, nil)
    if err != nil {
        config.Logger.Error("Error creating request: ", err)
        return
    }

    resp, err := client.Do(req)
    if (err != nil) {
        config.Logger.Error("Error logging in: ", err)
        return
    }

    body, err := helpers.HandleHTTPResponse(resp)
    if err != nil {
        config.Logger.Error("Error: ", err)
        return
    }

    fmt.Println("Response:", body)

    sessionData := map[string]interface{}{
        "server":     server,
        "port":       port,
        "httpScheme": httpScheme,
        "response":   body,
    }

    if err := helpers.WriteSessionFile(sessionData); err != nil {
        config.Logger.Error("Error writing session data to file: ", err)
    } else {
        config.Logger.Info("Authenticated!")
    }
}

// validateSession validates the current session by checking the session file and making a request to the Proxmox server.
//
// Parameters:
// - trust: A boolean indicating whether to trust SSL certificates.
//
// Returns:
// - bool: True if the session is valid, false otherwise.
func validateSession(trust bool) bool {
    sessionData, err := helpers.ReadSessionFile()
    if err != nil {
        config.Logger.Error("Error reading session file: ", err)
        return false
    }

    server, ok := sessionData["server"].(string)
    if !ok || server == "" {
        config.Logger.Error("Invalid session: missing server information")
        return false
    }

    httpScheme, ok := sessionData["httpScheme"].(string)
    if !ok || httpScheme == "" {
        config.Logger.Error("Invalid session: missing HTTP scheme")
        return false
    }

    port, ok := sessionData["port"].(float64)
    if !ok || port <= 0 {
        config.Logger.Error("Invalid session: missing or invalid port")
        return false
    }

    uri := fmt.Sprintf("%s://%s:%d/api2/json/access/ticket", httpScheme, server, int(port))

    response, ok := sessionData["response"].(string)
    if !ok || response == "" {
        config.Logger.Error("Invalid session: missing response information")
        return false
    }

    var responseData map[string]interface{}
    if err := json.Unmarshal([]byte(response), &responseData); err != nil {
        config.Logger.Error("Error parsing response JSON: ", err)
        return false
    }

    data, ok := responseData["data"]
    if !ok {
        config.Logger.Error("Invalid session: missing 'data' field in response")
        return false
    }

    username, ok := data.(map[string]interface{})["username"]
	if !ok {
		config.Logger.Error("Invalid session: missing 'username' field in data")
		return false
	}

    ticket, ok := data.(map[string]interface{})["ticket"]
    if !ok {
        config.Logger.Error("Invalid session: missing 'ticket' field in data")
        return false
    }

    csrfToken, ok := data.(map[string]interface{})["CSRFPreventionToken"]
    if !ok {
        config.Logger.Error("Invalid session: missing 'CSRFPreventionToken' field in data")
        return false
    }

    client := helpers.CreateHTTPClient(trust)

    ticketStr, ok := ticket.(string)
    if !ok {
        config.Logger.Error("Error: ticket is not a string")
        return false
    }

    payload := fmt.Sprintf("username=%s&password=%s", username, url.QueryEscape(ticketStr))
    headers := map[string]string{
        "Content-Type":          "application/x-www-form-urlencoded; charset=UTF-8",
    }

    req, err := helpers.CreateHTTPRequest("POST", uri, payload, headers, []*http.Cookie{
        {
            Name:  "PVEAuthCookie",
            Value: url.QueryEscape(ticketStr),
        },
    })
    if err != nil {
        config.Logger.Error("Error creating POST request: ", err)
        return false
    }

    req.Header = http.Header{}
	req.Header["CSRFPreventionToken"] = []string{csrfToken.(string)} 

    resp, err := client.Do(req)
    if err != nil {
        config.Logger.Error("Error validating session: ", err)
        return false
    }

    bodyBytes, err := helpers.HandleHTTPResponse(resp)
    if err != nil {
        config.Logger.Error("Session validation failed: ", err)
        return false
    }

    helpers.UpdateSessionField("response", bodyBytes)

    return true
}
