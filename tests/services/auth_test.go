// auth_test.go contains unit tests for the AuthService struct in the services package.
//
// These tests verify the behavior of authentication logic, including login and session validation.
package tests

import (
	"fmt"
	"net/http"
	"testing"

	"proxmox-cli/services"

	"github.com/sirupsen/logrus"
)

// mockHTTPService is a mock implementation of the HTTP service used for testing AuthService.
type mockHTTPService struct {
	postFunc   func(uri, payload string, headers map[string]string, cookies []*http.Cookie) (string, error)
	getFunc    func(url string, headers map[string]string, cookies []*http.Cookie) (*http.Response, error)
	putFunc    func(uri, payload string, headers map[string]string, cookies []*http.Cookie) (string, error)
	deleteFunc func(url string, headers map[string]string, cookies []*http.Cookie) (string, error)
}

// Post mocks the HTTP POST request for testing purposes.
func (m *mockHTTPService) Post(uri, payload string, headers map[string]string, cookies []*http.Cookie) (string, error) {
	return m.postFunc(uri, payload, headers, cookies)
}

// Get mocks the HTTP GET request for testing purposes.
func (m *mockHTTPService) Get(url string, headers map[string]string, cookies []*http.Cookie) (*http.Response, error) {
	if m.getFunc != nil {
		return m.getFunc(url, headers, cookies)
	}
	return nil, nil
}

// Put mocks the HTTP PUT request for testing purposes.
func (m *mockHTTPService) Put(uri, payload string, headers map[string]string, cookies []*http.Cookie) (string, error) {
	if m.putFunc != nil {
		return m.putFunc(uri, payload, headers, cookies)
	}
	return "", nil
}

// Delete mocks the HTTP DELETE request for testing purposes.
func (m *mockHTTPService) Delete(url string, headers map[string]string, cookies []*http.Cookie) (string, error) {
	if m.deleteFunc != nil {
		return m.deleteFunc(url, headers, cookies)
	}
	return "", nil
}

// mockSessionService is a mock implementation of the session service used for testing AuthService.
type mockSessionService struct {
	writeSessionFileFunc   func(data services.SessionData) error
	readSessionFileFunc    func() (services.SessionData, error)
	updateSessionFieldFunc func(field string, value interface{}) error
}

// WriteSessionFile mocks writing session data to a file.
func (m *mockSessionService) WriteSessionFile(data services.SessionData) error {
	return m.writeSessionFileFunc(data)
}

// ReadSessionFile mocks reading session data from a file.
func (m *mockSessionService) ReadSessionFile() (services.SessionData, error) {
	return m.readSessionFileFunc()
}

// UpdateSessionField mocks updating a field in the session file.
func (m *mockSessionService) UpdateSessionField(field string, value interface{}) error {
	return m.updateSessionFieldFunc(field, value)
}

func TestAuthService_ValidateSession_Success(t *testing.T) {
	logger := logrus.New()
	validSession := services.SessionData{
		Server:     "localhost",
		Port:       8006,
		HttpScheme: "https",
		Response: services.SessionDataResponse{
			Data: struct {
				Username            string `json:"username"`
				Ticket              string `json:"ticket"`
				CSRFPreventionToken string `json:"CSRFPreventionToken"`
			}{
				Username:            "user",
				Ticket:              "ticket",
				CSRFPreventionToken: "csrf",
			},
		},
	}
	mockHTTP := &mockHTTPService{
		postFunc: func(uri, payload string, headers map[string]string, cookies []*http.Cookie) (string, error) {
			return `{"data":{}}`, nil
		},
	}
	mockSession := &mockSessionService{
		writeSessionFileFunc:   func(data services.SessionData) error { return nil },
		readSessionFileFunc:    func() (services.SessionData, error) { return validSession, nil },
		updateSessionFieldFunc: func(field string, value interface{}) error { return nil },
	}
	authService := services.NewAuthServiceWithDeps(logger, true, mockHTTP, mockSession)
	valid := authService.ValidateSession()
	if !valid {
		t.Error("expected ValidateSession to return true for valid session")
	}
}

func TestAuthService_ValidateSession_HttpFailure(t *testing.T) {
	logger := logrus.New()
	validSession := services.SessionData{
		Server:     "localhost",
		Port:       8006,
		HttpScheme: "https",
		Response: services.SessionDataResponse{
			Data: struct {
				Username            string `json:"username"`
				Ticket              string `json:"ticket"`
				CSRFPreventionToken string `json:"CSRFPreventionToken"`
			}{
				Username:            "user",
				Ticket:              "ticket",
				CSRFPreventionToken: "csrf",
			},
		},
	}
	mockHTTP := &mockHTTPService{
		postFunc: func(uri, payload string, headers map[string]string, cookies []*http.Cookie) (string, error) {
			return "", fmt.Errorf("network error")
		},
	}
	mockSession := &mockSessionService{
		writeSessionFileFunc:   func(data services.SessionData) error { return nil },
		readSessionFileFunc:    func() (services.SessionData, error) { return validSession, nil },
		updateSessionFieldFunc: func(field string, value interface{}) error { return nil },
	}
	authService := services.NewAuthServiceWithDeps(logger, true, mockHTTP, mockSession)
	valid := authService.ValidateSession()
	if valid {
		t.Error("expected ValidateSession to return false on HTTP failure")
	}
}

func TestAuthService_ValidateSession_UpdateSessionFieldFailure(t *testing.T) {
	logger := logrus.New()
	validSession := services.SessionData{
		Server:     "localhost",
		Port:       8006,
		HttpScheme: "https",
		Response: services.SessionDataResponse{
			Data: struct {
				Username            string `json:"username"`
				Ticket              string `json:"ticket"`
				CSRFPreventionToken string `json:"CSRFPreventionToken"`
			}{
				Username:            "user",
				Ticket:              "ticket",
				CSRFPreventionToken: "csrf",
			},
		},
	}
	mockHTTP := &mockHTTPService{
		postFunc: func(uri, payload string, headers map[string]string, cookies []*http.Cookie) (string, error) {
			return `{"data":{}}`, nil
		},
	}
	mockSession := &mockSessionService{
		writeSessionFileFunc:   func(data services.SessionData) error { return nil },
		readSessionFileFunc:    func() (services.SessionData, error) { return validSession, nil },
		updateSessionFieldFunc: func(field string, value interface{}) error { return fmt.Errorf("write error") },
	}
	authService := services.NewAuthServiceWithDeps(logger, true, mockHTTP, mockSession)
	valid := authService.ValidateSession()
	if valid {
		t.Error("expected ValidateSession to return false on update session field failure")
	}
}

func TestAuthService_ValidateSession_InvalidSessionData(t *testing.T) {
	logger := logrus.New()
	invalidSession := services.SessionData{
		Response: services.SessionDataResponse{
			Data: struct {
				Username            string `json:"username"`
				Ticket              string `json:"ticket"`
				CSRFPreventionToken string `json:"CSRFPreventionToken"`
			}{},
		},
	}
	mockHTTP := &mockHTTPService{
		postFunc: func(uri, payload string, headers map[string]string, cookies []*http.Cookie) (string, error) {
			return `{"data":{}}`, nil
		},
	}
	mockSession := &mockSessionService{
		writeSessionFileFunc: func(data services.SessionData) error { return nil },
		readSessionFileFunc: func() (services.SessionData, error) {
			return invalidSession, fmt.Errorf("invalid session: missing server")
		},
		updateSessionFieldFunc: func(field string, value interface{}) error { return nil },
	}
	authService := services.NewAuthServiceWithDeps(logger, true, mockHTTP, mockSession)
	valid := authService.ValidateSession()
	if valid {
		t.Error("expected ValidateSession to return false for invalid session data")
	}
}

func TestAuthService_LoginToProxmox_Success(t *testing.T) {
	logger := logrus.New()
	mockHTTP := &mockHTTPService{
		postFunc: func(uri, payload string, headers map[string]string, cookies []*http.Cookie) (string, error) {
			return `{"data":{"username":"user","ticket":"ticket","CSRFPreventionToken":"csrf"}}`, nil
		},
	}
	mockSession := &mockSessionService{
		writeSessionFileFunc:   func(data services.SessionData) error { return nil },
		readSessionFileFunc:    func() (services.SessionData, error) { return services.SessionData{}, nil },
		updateSessionFieldFunc: func(field string, value interface{}) error { return nil },
	}
	authService := services.NewAuthServiceWithDeps(logger, true, mockHTTP, mockSession)
	err := authService.LoginToProxmox("localhost", 8006, "https", "user", "pass")
	if err != nil {
		t.Errorf("expected LoginToProxmox to succeed, got error: %v", err)
	}
}

func TestAuthService_LoginToProxmox_HttpFailure(t *testing.T) {
	logger := logrus.New()
	mockHTTP := &mockHTTPService{
		postFunc: func(uri, payload string, headers map[string]string, cookies []*http.Cookie) (string, error) {
			return "", fmt.Errorf("network error")
		},
	}
	mockSession := &mockSessionService{
		writeSessionFileFunc:   func(data services.SessionData) error { return nil },
		readSessionFileFunc:    func() (services.SessionData, error) { return services.SessionData{}, nil },
		updateSessionFieldFunc: func(field string, value interface{}) error { return nil },
	}
	authService := services.NewAuthServiceWithDeps(logger, true, mockHTTP, mockSession)
	err := authService.LoginToProxmox("localhost", 8006, "https", "user", "pass")
	if err == nil {
		t.Error("expected LoginToProxmox to return error on HTTP failure")
	}
}

func TestAuthService_LoginToProxmox_InvalidJson(t *testing.T) {
	logger := logrus.New()
	mockHTTP := &mockHTTPService{
		postFunc: func(uri, payload string, headers map[string]string, cookies []*http.Cookie) (string, error) {
			return "not json", nil
		},
	}
	mockSession := &mockSessionService{
		writeSessionFileFunc:   func(data services.SessionData) error { return nil },
		readSessionFileFunc:    func() (services.SessionData, error) { return services.SessionData{}, nil },
		updateSessionFieldFunc: func(field string, value interface{}) error { return nil },
	}
	authService := services.NewAuthServiceWithDeps(logger, true, mockHTTP, mockSession)
	err := authService.LoginToProxmox("localhost", 8006, "https", "user", "pass")
	if err == nil {
		t.Error("expected LoginToProxmox to return error on invalid JSON")
	}
}

func TestAuthService_LoginToProxmox_WriteSessionFileFailure(t *testing.T) {
	logger := logrus.New()
	mockHTTP := &mockHTTPService{
		postFunc: func(uri, payload string, headers map[string]string, cookies []*http.Cookie) (string, error) {
			return `{"data":{"username":"user","ticket":"ticket","CSRFPreventionToken":"csrf"}}`, nil
		},
	}
	mockSession := &mockSessionService{
		writeSessionFileFunc:   func(data services.SessionData) error { return fmt.Errorf("write error") },
		readSessionFileFunc:    func() (services.SessionData, error) { return services.SessionData{}, nil },
		updateSessionFieldFunc: func(field string, value interface{}) error { return nil },
	}
	authService := services.NewAuthServiceWithDeps(logger, true, mockHTTP, mockSession)
	err := authService.LoginToProxmox("localhost", 8006, "https", "user", "pass")
	if err == nil {
		t.Error("expected LoginToProxmox to return error on write session file failure")
	}
}
