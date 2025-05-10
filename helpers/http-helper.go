package helpers

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
)

// CreateHTTPClient creates an HTTP client with optional SSL verification skipping
func CreateHTTPClient(trust bool) *http.Client {
	if trust {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		return &http.Client{Transport: tr}
	}
	return &http.Client{}
}

// CreateHTTPRequest creates an HTTP request with the specified method, URL, payload, headers, and cookies
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
    return req, nil
}

// HandleHTTPResponse processes an HTTP response and returns the body as a string if the status code is 200
func HandleHTTPResponse(resp *http.Response) (string, error) {
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
