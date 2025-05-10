package helpers

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"

	"proxmox-cli/config"
)

// CreateHTTPClient creates an HTTP client with optional SSL verification skipping.
// If `trust` is true, the client will skip SSL certificate verification.
// Otherwise, it will use the default HTTP client settings.
//
// Parameters:
// - trust: A boolean indicating whether to skip SSL verification.
//
// Returns:
// - *http.Client: The configured HTTP client.
func CreateHTTPClient(trust bool) *http.Client {
	if trust {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		return &http.Client{Transport: tr}
	}
	return &http.Client{}
}

// CreateHTTPRequest creates an HTTP request with the specified method, URL, payload, headers, and cookies.
//
// Parameters:
// - method: The HTTP method (e.g., "GET", "POST").
// - url: The target URL for the request.
// - payload: The request body as a string.
// - headers: A map of header key-value pairs.
// - cookies: A slice of HTTP cookies to include in the request.
//
// Returns:
// - *http.Request: The created HTTP request.
// - error: An error if the request creation fails.
func CreateHTTPRequest(method, url, payload string, headers map[string]string, cookies []*http.Cookie) (*http.Request, error) {
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

    LogHTTPRequestContent(req)
    return req, nil
}

// HandleHTTPResponse processes an HTTP response and returns the body as a string if the status code is 200.
//
// Parameters:
// - resp: The HTTP response to process.
//
// Returns:
// - string: The response body as a string.
// - error: An error if the response status code is not 200 or if reading the body fails.
func HandleHTTPResponse(resp *http.Response) (string, error) {
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }
    if resp.StatusCode != 200 {
        return "", fmt.Errorf("received status code %d: %s", resp.StatusCode, string(body))
    }

    LogHTTPResponseContent(resp, string(body))
    return string(body), nil
}

// LogHTTPRequestContent logs the details of an HTTP request, including method, URL, headers, and body.
//
// Parameters:
// - req: The HTTP request to log.
func LogHTTPRequestContent(req *http.Request) {
	config.Logger.Info("--- HTTP Request ---")
	config.Logger.Infof("Method: %s", req.Method)
	config.Logger.Infof("URL: %s", req.URL.String())
	config.Logger.Info("Headers:")
	for key, values := range req.Header {
		for _, value := range values {
			config.Logger.Infof("  %s: %s", key, value)
		}
	}
	if req.Body != nil {
		bodyBytes, _ := io.ReadAll(req.Body)
		config.Logger.Infof("Body: %s", string(bodyBytes))
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Reassign the body after reading
	}
	config.Logger.Info("--------------------")
}

// LogHTTPResponseContent logs the details of an HTTP response, including status code, headers, and body.
//
// Parameters:
// - resp: The HTTP response to log.
func LogHTTPResponseContent(resp *http.Response, body string) {
	config.Logger.Info("--- HTTP Response ---")
	config.Logger.Infof("Status Code: %d", resp.StatusCode)
	config.Logger.Info("Headers:")
	for key, values := range resp.Header {
		for _, value := range values {
			config.Logger.Infof("  %s: %s", key, value)
		}
	}
	config.Logger.Infof("Body: %s", body)
	config.Logger.Info("--------------------")
}
