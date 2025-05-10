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

// Helper function to create an HTTP client
func createHTTPClient(trust bool) *http.Client {
    if trust {
        tr := &http.Transport{
            TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
        }
        return &http.Client{Transport: tr}
    }
    return &http.Client{}
}

// Helper function to create an HTTP request
func createHTTPRequest(method, url, payload string, headers map[string]string, cookies []*http.Cookie) (*http.Request, error) {
    req, err := http.NewRequest(method, url, bytes.NewBuffer([]byte(payload)))
    if err != nil {
        return nil, err
    }
    for key, value := range headers {
        req.Header.Set(key, value)
    }
    for _, cookie := range cookies {
        req.AddCookie(cookie)
    }
    return req, nil
}

// Helper function to handle HTTP responses
func handleHTTPResponse(resp *http.Response) (string, error) {
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }
    if resp.StatusCode != 200 {
        return "", fmt.Errorf("received status code %d: %s", resp.StatusCode, string(body))
    }
    return string(body), nil
}

// Helper function to read session data from file
func readSessionFile() (map[string]interface{}, error) {
    usr, err := user.Current()
    if err != nil {
        return nil, err
    }
    sessionFilePath := filepath.Join(usr.HomeDir, ".proxmox", "session")
    file, err := os.Open(sessionFilePath)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    var sessionData map[string]interface{}
    decoder := json.NewDecoder(file)
    if err := decoder.Decode(&sessionData); err != nil {
        return nil, err
    }
    return sessionData, nil
}

// Helper function to write session data to file
func writeSessionFile(data map[string]interface{}) error {
    usr, err := user.Current()
    if err != nil {
        return err
    }
    dirPath := filepath.Join(usr.HomeDir, ".proxmox")
    if _, err := os.Stat(dirPath); os.IsNotExist(err) {
        if err := os.MkdirAll(dirPath, 0755); err != nil {
            return err
        }
    }
    filePath := filepath.Join(dirPath, "session")
    file, err := os.Create(filePath)
    if err != nil {
        return err
    }
    defer file.Close()

    encoder := json.NewEncoder(file)
    return encoder.Encode(data)
}

// Refactor loginToProxmox to use helper functions
func loginToProxmox(server string, port int, httpScheme string, username string, password string, trust bool) {
    uri := fmt.Sprintf("%s://%s:%d/api2/json/access/ticket", httpScheme, server, port)
    client := createHTTPClient(trust)

    payload := fmt.Sprintf("username=%s&password=%s&realm=pam&new-format=1", username, password)
    headers := map[string]string{
        "Content-Type": "application/x-www-form-urlencoded; charset=UTF-8",
    }

    req, err := createHTTPRequest("POST", uri, payload, headers, nil)
    if err != nil {
        fmt.Println("Error creating request:", err)
        return
    }

    resp, err := client.Do(req)
    if err != nil {
        fmt.Println("Error logging in:", err)
        return
    }

    body, err := handleHTTPResponse(resp)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }

    fmt.Println("Response:", body)

    sessionData := map[string]interface{}{
        "server":     server,
        "port":       port,
        "httpScheme": httpScheme,
        "response":   body,
    }

    if err := writeSessionFile(sessionData); err != nil {
        fmt.Println("Error writing session data to file:", err)
    } else {
        fmt.Println("Authenticated!")
    }
}

// Refactor validateSession to use helper functions
func validateSession(trust bool) bool {
    sessionData, err := readSessionFile()
    if err != nil {
        fmt.Println("Error reading session file:", err)
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
    if !ok || response == "" {
        fmt.Println("Invalid session: missing response information")
        return false
    }

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

    username, ok := data.(map[string]interface{})["username"]
	if !ok {
		fmt.Println("Invalid session: missing 'username' field in data")
		return false
	}

    ticket, ok := data.(map[string]interface{})["ticket"]
    if !ok {
        fmt.Println("Invalid session: missing 'ticket' field in data")
        return false
    }

    csrfToken, ok := data.(map[string]interface{})["CSRFPreventionToken"]
    if !ok {
        fmt.Println("Invalid session: missing 'CSRFPreventionToken' field in data")
        return false
    }

    client := createHTTPClient(trust)

    ticketStr, ok := ticket.(string)
    if !ok {
        fmt.Println("Error: ticket is not a string")
        return false
    }

    payload := fmt.Sprintf("username=%s&password=%s", username, url.QueryEscape(ticketStr))
    headers := map[string]string{
        "Content-Type":          "application/x-www-form-urlencoded; charset=UTF-8",
    }

    req, err := createHTTPRequest("POST", uri, payload, headers, []*http.Cookie{
        {
            Name:  "PVEAuthCookie",
            Value: url.QueryEscape(ticketStr),
        },
    })
    if err != nil {
        fmt.Println("Error creating POST request:", err)
        return false
    }

    req.Header = http.Header{} // Reset headers to avoid normalization
	req.Header["CSRFPreventionToken"] = []string{csrfToken.(string)} 

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

    _, err = handleHTTPResponse(resp)
    if err != nil {
        fmt.Println("Session validation failed:", err)
        return false
    }

    fmt.Println("Session is valid.")
    return true
}
