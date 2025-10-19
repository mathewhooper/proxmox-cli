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

type SessionService struct {
	homeDir string
	logger  *logrus.Logger
}

type SessionData struct {
	Server     string              `json:"server"`
	Port       int                 `json:"port"`
	HttpScheme string              `json:"httpScheme"`
	Response   SessionDataResponse `json:"response"`
}

type SessionDataResponse struct {
	Data struct {
		Username            string `json:"username"`
		Ticket              string `json:"ticket"`
		CSRFPreventionToken string `json:"CSRFPreventionToken"`
	} `json:"data"`
}

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

func (s *SessionService) WriteSessionFile(sessionData SessionData) error {
	filePath, err := s.getSessionFilepath()
	if err != nil {
		return err
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(file)
	return encoder.Encode(sessionData)
}

func (s *SessionService) ReadSessionFile() (SessionData, error) {
	filePath, err := s.getSessionFilepath()
	if err != nil {
		return SessionData{}, err
	}

	file, err := os.Open(filePath)
	if err != nil {
		return SessionData{}, err
	}
	defer file.Close()

	var sessionData SessionData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&sessionData); err != nil {
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

func (s *SessionService) UpdateSessionField(field string, value interface{}) error {
	sessionData, err := s.ReadSessionFile()
	if err != nil {
		return err
	}

	switch field {
	case "server":
		sessionData.Server = value.(string)
	case "port":
		sessionData.Port = value.(int)
	case "httpScheme":
		sessionData.HttpScheme = value.(string)
	case "response":
		var resp SessionDataResponse
		if str, ok := value.(string); ok {
			if err := json.Unmarshal([]byte(str), &resp); err != nil {
				return err
			}
			sessionData.Response = resp
		}
	}
	return s.WriteSessionFile(sessionData)
}

func (s *SessionService) getSessionFilepath() (string, error) {
	dirPath := filepath.Join(s.homeDir, ".proxmox")
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return "", err
		}
	}
	filePath := filepath.Join(dirPath, "session")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		file, err := os.Create(filePath)
		if err != nil {
			return "", err
		}
		file.Close()
	}
	return filePath, nil
}
