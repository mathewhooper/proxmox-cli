package commands

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func NewLoginCommand() *cobra.Command {
	var server string
    var username string
    var httpScheme string
	var port int
	var trust bool

	// Create the login command
	var loginCmd = &cobra.Command{
		Use:   "login",
		Short: "Log in to a Proxmox server",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Print("Enter Password: ")
			passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
			fmt.Println() // Print a newline after password input
			if err != nil {
				fmt.Println("Error reading password:", err)
				return
			}
			password := string(passwordBytes)
			loginToProxmox(server, port, httpScheme, username, password, trust)
		},
	}

	// Add flags to the login command
	loginCmd.Flags().StringVarP(&server, "server", "s", "", "Proxmox server URL")
	loginCmd.Flags().StringVarP(&username, "username", "u", "", "Username for Proxmox")
	loginCmd.Flags().IntVarP(&port, "port", "P", 8006, "Proxmox server port")
	loginCmd.Flags().StringVarP(&httpScheme, "httpScheme", "S", "https", "HTTP scheme (http or https)")
	loginCmd.Flags().BoolVarP(&trust, "trust", "t", false, "Trust SSL certificates")

	// Mark flags as required
	loginCmd.MarkFlagRequired("server")
	loginCmd.MarkFlagRequired("username")

	return loginCmd
}

func ValidateLoginCommand() *cobra.Command {
	var trust bool

    var validateCmd = &cobra.Command{
        Use:   "validate",
        Short: "Validate the current session",
        Run: func(cmd *cobra.Command, args []string) {
            if validateSession(trust) {
                fmt.Println("Session is valid.")
            } else {
                fmt.Println("Session is invalid.")
            }
        },
    }

    validateCmd.Flags().BoolVarP(&trust, "trust", "t", false, "Trust SSL certificates")
    
    return validateCmd
}

func loginToProxmox(server string, port int, httpScheme string, username string, password string, trust bool) {
	uri := fmt.Sprintf("%s://%s:%d/api2/json/access/ticket", httpScheme, server, port)

	// Create a custom HTTP client if trust is true
	client := &http.Client{}
	if trust {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{Transport: tr}
	}

	// Create the request payload
	payload := fmt.Sprintf("username=%s&password=%s&realm=pam&new-format=1", username, password)

	// Make the HTTP request
	req, err := http.NewRequest("POST", uri, bytes.NewBuffer([]byte(payload)))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}
	// Set charset=UTF-8 in the Content-Type header
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error logging in:", err)
		return
	}
	defer resp.Body.Close()

	// Print the response status code
	fmt.Println("Status Code:", resp.StatusCode)

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}

    // Check if the status code is 200
	if resp.StatusCode != 200 {
		fmt.Printf("Error: Received status code %d\n", resp.StatusCode)
        fmt.Println("Response:", string(body))
		return
	}

	// Print the response
	fmt.Println("Response:", string(body))

	// Store the session data
	sessionData := map[string]interface{}{
		"server":     server,
		"port":       port,
		"httpScheme": httpScheme,
		"response":   string(body),
	}

	usr, err := user.Current()
	if err != nil {
		fmt.Println("Error getting user profile directory:", err)
		return
	}

	dirPath := filepath.Join(usr.HomeDir, ".proxmox")

	// Check if the directory already exists
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			fmt.Println("Error creating .proxmox directory:", err)
			return
		}
	}

	filePath := filepath.Join(dirPath, "session")
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println("Error creating session file:", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(sessionData); err != nil {
		fmt.Println("Error writing session data to file:", err)
		return
	}

	fmt.Println("Authenticated!")
}

func validateSession(trust bool) bool {
	usr, err := user.Current()
	if err != nil {
		fmt.Println("Error getting user profile directory:", err)
		return false
	}

	sessionFilePath := filepath.Join(usr.HomeDir, ".proxmox", "session")
	file, err := os.Open(sessionFilePath)
	if err != nil {
		fmt.Println("Error opening session file:", err)
		return false
	}
	defer file.Close()

	var sessionData map[string]interface{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&sessionData); err != nil {
		fmt.Println("Error decoding session file:", err)
		return false
	}

	server, ok := sessionData["server"].(string)
	if !ok || server == "" {
		fmt.Println("Invalid session: missing server information")
		return false
	}

	httpScheme, ok := sessionData["httpScheme"].(string)
	if !ok || httpScheme == "" {
		fmt.Println("Invalid session: missing HTTP scheme")
		return false
	}

	port, ok := sessionData["port"].(float64) // JSON numbers are decoded as float64
	if !ok || port <= 0 {
		fmt.Println("Invalid session: missing or invalid port")
		return false
	}

	uri := fmt.Sprintf("%s://%s:%d/api2/json/access/ticket", httpScheme, server, int(port))

    response, ok := sessionData["response"].(string)
	if !ok || server == "" {
		fmt.Println("Invalid session: missing response information")
		return false
	}

	// Parse the response body to extract the 'data' field
	var responseData map[string]interface{}
	if err := json.Unmarshal([]byte(response), &responseData); err != nil {
		fmt.Println("Error parsing response JSON:", err)
		return false
	}

	data, ok := responseData["data"]
	if !ok {
		fmt.Println("Invalid session: missing 'data' field in response")
		return false
	}

	// Extract the 'ticket' field from the 'data' object
	ticket, ok := data.(map[string]interface{})["ticket"]
	if !ok {
		fmt.Println("Invalid session: missing 'ticket' field in data")
		return false
	}

	fmt.Println("Extracted ticket:", ticket)

	// Extract the 'username' field from the 'data' object
	username, ok := data.(map[string]interface{})["username"]
	if !ok {
		fmt.Println("Invalid session: missing 'username' field in data")
		return false
	}

	fmt.Println("Extracted username:", username)

	// Extract the 'CSRFPreventionToken' field from the 'data' object
	csrfToken, ok := data.(map[string]interface{})["CSRFPreventionToken"]
	if !ok {
		fmt.Println("Invalid session: missing 'CSRFPreventionToken' field in data")
		return false
	}

	fmt.Println("Extracted CSRFPreventionToken:", csrfToken)

	// Create a custom HTTP client if trust is true
	client := &http.Client{}
	if trust {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{Transport: tr}
	}

	// Prepare the payload for the POST request
	ticketStr, ok := ticket.(string)
	if !ok {
		fmt.Println("Error: ticket is not a string")
		return false
	}
	payload := fmt.Sprintf("username=%s&password=%s", username, url.QueryEscape(ticketStr))
    fmt.Println("Payload:", payload)

	// Make the HTTP POST request
	req, err := http.NewRequest("POST", uri, bytes.NewBuffer([]byte(payload)))
	if err != nil {
		fmt.Println("Error creating POST request:", err)
		return false
	}
	// Set charset=UTF-8 in the Content-Type header
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")

	// Use a custom approach to set headers with exact casing
	req.Header = http.Header{} // Reset headers to avoid normalization
	req.Header["CSRFPreventionToken"] = []string{csrfToken.(string)} // Set exact casing

	// Add the PVEAuthCookie to the request as a cookie
	cookie := &http.Cookie{
		Name:  "PVEAuthCookie",
		Value: url.QueryEscape(ticketStr), // Use URL encoding instead of base64
	}
	req.AddCookie(cookie)

    fmt.Println("Cookie:", cookie)

	// Print the full HTTP request for debugging
	fmt.Println("--- HTTP Request ---")
	fmt.Printf("Method: %s\n", req.Method)
	fmt.Printf("URL: %s\n", req.URL.String())
	fmt.Println("Headers:")
	for key, values := range req.Header {
		for _, value := range values {
			fmt.Printf("  %s: %s\n", key, value)
		}
	}
	if req.Body != nil {
		bodyBytes, _ := io.ReadAll(req.Body)
		fmt.Printf("Body: %s\n", string(bodyBytes))
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Reassign the body after reading
	}
	fmt.Println("--------------------")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error validating session:", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("Session validation failed with status code %d\n", resp.StatusCode)
		//return false
	}

    // Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return false
	}

    // Print the full response body for debugging
	fmt.Println("Response Body:", string(body))

	if resp.StatusCode != 200 {
		fmt.Printf("Session validation failed with status code %d\n", resp.StatusCode)
		//return false
	}

    fmt.Println("Response:", string(body))

	// Print the full HTTP response for debugging
	fmt.Println("--- HTTP Response ---")
	fmt.Printf("Status Code: %d\n", resp.StatusCode)
	fmt.Println("Headers:")
	for key, values := range resp.Header {
		for _, value := range values {
			fmt.Printf("  %s: %s\n", key, value)
		}
	}
	fmt.Println("Body:")
	fmt.Println(string(body))
	fmt.Println("--------------------")

	return true
}
