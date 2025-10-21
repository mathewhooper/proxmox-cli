package services

import (
	"fmt"
	"os"

	"encoding/json"

	"path/filepath"

	"github.com/sirupsen/logrus"
)

// SessionServiceInterface allows mocking of session file operations for AuthService.
type SessionServiceInterface interface {
	WriteSessionFile(sessionData SessionData) error
	ReadSessionFile() (SessionData, error)
	UpdateSessionField(field string, value interface{}) error
}

// SessionService manages Proxmox session data persistence
type SessionService struct {
	homeDir string
	logger  *logrus.Logger
}

// SessionData represents the Proxmox session information stored locally
type SessionData struct {
	Server     string              `json:"server"`
	Port       int                 `json:"port"`
	HttpScheme string              `json:"httpScheme"`
	Response   SessionDataResponse `json:"response"`
}

// SessionDataResponse represents the authentication response from Proxmox API
type SessionDataResponse struct {
	Data struct {
		Username            string `json:"username"`
		Ticket              string `json:"ticket"`
		CSRFPreventionToken string `json:"CSRFPreventionToken"`
	} `json:"data"`
}

// NewSessionService creates a new SessionService instance
func NewSessionService(logger *logrus.Logger) (*SessionService, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	return &SessionService{
		homeDir: dir,
		logger:  logger,
	}, nil
}

// WriteSessionFile writes session data to the session file
func (s *SessionService) WriteSessionFile(sessionData SessionData) error {
	filePath, err := s.getSessionFilepath()
	if err != nil {
		return err
	}

	//nolint:gosec // G304: File path is from trusted getSessionFilepath method
	file, err := os.Create(filePath) // #nosec G304 -- File path from trusted getSessionFilepath
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(file)
	return encoder.Encode(sessionData)
}

// ReadSessionFile reads and validates session data from the session file
func (s *SessionService) ReadSessionFile() (SessionData, error) {
	filePath, err := s.getSessionFilepath()
	if err != nil {
		return SessionData{}, err
	}

	//nolint:gosec // G304: File path is from trusted getSessionFilepath method
	file, err := os.Open(filePath) // #nosec G304 -- File path from trusted getSessionFilepath
	if err != nil {
		return SessionData{}, err
	}
	//nolint:errcheck // Best effort close in defer
	defer func() { _ = file.Close() }()

	var sessionData SessionData
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&sessionData)
	if err != nil {
		return SessionData{}, err
	}

	// Validate that all fields in sessionData are not empty values
	if sessionData.Server == "" {
		return SessionData{}, fmt.Errorf("invalid session: missing server")
	}
	if sessionData.Port <= 0 {
		return SessionData{}, fmt.Errorf("invalid session: missing or invalid port")
	}
	if sessionData.HttpScheme == "" {
		return SessionData{}, fmt.Errorf("invalid session: missing httpScheme")
	}

	if sessionData.Response.Data.Username == "" {
		return SessionData{}, fmt.Errorf("invalid session: missing 'username' field in data")
	}
	if sessionData.Response.Data.Ticket == "" {
		return SessionData{}, fmt.Errorf("invalid session: missing 'ticket' field in data")
	}
	if sessionData.Response.Data.CSRFPreventionToken == "" {
		return SessionData{}, fmt.Errorf("invalid session: missing 'CSRFPreventionToken' field in data")
	}

	return sessionData, nil
}

// UpdateSessionField updates a specific field in the session data
func (s *SessionService) UpdateSessionField(field string, value interface{}) error {
	sessionData, err := s.ReadSessionFile()
	if err != nil {
		return err
	}

	switch field {
	case "server":
		//nolint:errcheck // Type assertion is safe within switch on field name
		sessionData.Server = value.(string)
	case "port":
		//nolint:errcheck // Type assertion is safe within switch on field name
		sessionData.Port = value.(int)
	case "httpScheme":
		//nolint:errcheck // Type assertion is safe within switch on field name
		sessionData.HttpScheme = value.(string)
	case "response":
		var resp SessionDataResponse
		if str, ok := value.(string); ok {
			err = json.Unmarshal([]byte(str), &resp)
			if err != nil {
				return err
			}
			sessionData.Response = resp
		}
	}
	return s.WriteSessionFile(sessionData)
}

func (s *SessionService) getSessionFilepath() (string, error) {
	dirPath := filepath.Join(s.homeDir, ".proxmox")
	var err error
	if _, err = os.Stat(dirPath); os.IsNotExist(err) {
		//nolint:gosec // G301: Standard directory permissions for user config
		err = os.MkdirAll(dirPath, 0755) // #nosec G301 -- Standard permissions for user config directory
		if err != nil {
			return "", err
		}
	}
	filePath := filepath.Join(dirPath, "session")
	if _, err = os.Stat(filePath); os.IsNotExist(err) {
		var file *os.File
		//nolint:gosec // G304: File path is constructed from trusted homeDir
		file, err = os.Create(filePath) // #nosec G304 -- File path from trusted homeDir
		if err != nil {
			return "", err
		}
		//nolint:errcheck // Best effort close, file was just created
		_ = file.Close()
	}
	return filePath, nil
}
