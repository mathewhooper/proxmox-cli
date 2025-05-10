package helpers_test

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"proxmox-cli/helpers"
)

func TestCreateHTTPClient(t *testing.T) {
	client := helpers.CreateHTTPClient(true)
	if client == nil {
		t.Errorf("Expected HTTP client, got nil")
	}
}

func TestCreateHTTPRequest(t *testing.T) {
	method := "POST"
	url := "http://example.com"
	payload := "key=value"
	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}
	cookies := []*http.Cookie{
		{Name: "test", Value: "value"},
	}

	req, err := helpers.CreateHTTPRequest(method, url, payload, headers, cookies)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if req.Method != method {
		t.Errorf("Expected method %s, got %s", method, req.Method)
	}

	if req.URL.String() != url {
		t.Errorf("Expected URL %s, got %s", url, req.URL.String())
	}

	if req.Header.Get("Content-Type") != headers["Content-Type"] {
		t.Errorf("Expected Content-Type header %s, got %s", headers["Content-Type"], req.Header.Get("Content-Type"))
	}

	if len(req.Cookies()) != 1 || req.Cookies()[0].Name != "test" {
		t.Errorf("Expected cookie with name 'test', got %v", req.Cookies())
	}
}

func TestHandleHTTPResponse(t *testing.T) {
	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString("response body")),
	}

	body, err := helpers.HandleHTTPResponse(resp)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if body != "response body" {
		t.Errorf("Expected body 'response body', got %s", body)
	}

	resp.StatusCode = 500
	resp.Body = io.NopCloser(bytes.NewBufferString("error body"))
	_, err = helpers.HandleHTTPResponse(resp)
	if err == nil {
		t.Errorf("Expected error for status code 500, got nil")
	}
}

func TestLogHTTPRequestContent(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://example.com", nil)
	helpers.LogHTTPRequestContent(req)
	// Verify logs manually or mock config.Logger to capture output
}

func TestLogHTTPResponseContent(t *testing.T) {
	resp := &http.Response{
		StatusCode: 200,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewBufferString("response body")),
	}
	helpers.LogHTTPResponseContent(resp, "response body")
	// Verify logs manually or mock config.Logger to capture output
}
