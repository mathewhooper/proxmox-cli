package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/sirupsen/logrus"
)

// Storage represents a storage definition
type Storage struct {
	Storage string `json:"storage"`
	Type    string `json:"type"`
	Content string `json:"content,omitempty"`
	Shared  int    `json:"shared,omitempty"`
	Active  int    `json:"active,omitempty"`
	Enabled int    `json:"enabled,omitempty"`
}

// StorageContent represents content within a storage
type StorageContent struct {
	VolID  string `json:"volid"`
	Format string `json:"format,omitempty"`
	Size   int64  `json:"size,omitempty"`
	Used   int64  `json:"used,omitempty"`
	VMID   int    `json:"vmid,omitempty"`
	CTime  int64  `json:"ctime,omitempty"`
}

// StorageListResponse represents the API response for storage list
type StorageListResponse struct {
	Data []Storage `json:"data"`
}

// StorageContentResponse represents the API response for storage content
type StorageContentResponse struct {
	Data []StorageContent `json:"data"`
}

// StorageService handles storage-related operations
type StorageService struct {
	Logger         *logrus.Logger
	Trust          bool
	HttpService    HttpServiceInterface
	SessionService SessionServiceInterface
}

// NewStorageService creates a new StorageService with real dependencies
func NewStorageService(logger *logrus.Logger, trust bool) (*StorageService, error) {
	sessionService, err := NewSessionService(logger)
	if err != nil {
		return nil, err
	}

	return &StorageService{
		Logger:         logger,
		Trust:          trust,
		HttpService:    NewHttpService(logger, trust),
		SessionService: sessionService,
	}, nil
}

// NewStorageServiceWithDeps creates a StorageService with injected dependencies (for testing)
func NewStorageServiceWithDeps(logger *logrus.Logger, trust bool, httpService HttpServiceInterface, sessionService SessionServiceInterface) *StorageService {
	return &StorageService{
		Logger:         logger,
		Trust:          trust,
		HttpService:    httpService,
		SessionService: sessionService,
	}
}

// ListStorage retrieves a list of all storage
func (s *StorageService) ListStorage() ([]Storage, error) {
	sessionData, err := s.SessionService.ReadSessionFile()
	if err != nil {
		s.Logger.Error("Error reading session file: ", err)
		return nil, err
	}

	uri := fmt.Sprintf("%s://%s:%d/api2/json/storage",
		sessionData.HttpScheme, sessionData.Server, sessionData.Port)

	cookies := []*http.Cookie{
		{
			Name:  "PVEAuthCookie",
			Value: url.QueryEscape(sessionData.Response.Data.Ticket),
		},
	}

	resp, err := s.HttpService.Get(uri, nil, cookies)
	if err != nil {
		s.Logger.Error("Error listing storage: ", err)
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		s.Logger.Error("Error reading response body: ", err)
		return nil, err
	}

	var result StorageListResponse
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		s.Logger.Error("Error parsing response JSON: ", err)
		return nil, err
	}

	return result.Data, nil
}

// ListStorageContent retrieves the content of a specific storage on a node
func (s *StorageService) ListStorageContent(nodeName, storageName string) ([]StorageContent, error) {
	sessionData, err := s.SessionService.ReadSessionFile()
	if err != nil {
		s.Logger.Error("Error reading session file: ", err)
		return nil, err
	}

	uri := fmt.Sprintf("%s://%s:%d/api2/json/nodes/%s/storage/%s/content",
		sessionData.HttpScheme, sessionData.Server, sessionData.Port, nodeName, storageName)

	cookies := []*http.Cookie{
		{
			Name:  "PVEAuthCookie",
			Value: url.QueryEscape(sessionData.Response.Data.Ticket),
		},
	}

	resp, err := s.HttpService.Get(uri, nil, cookies)
	if err != nil {
		s.Logger.Error("Error listing storage content: ", err)
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		s.Logger.Error("Error reading response body: ", err)
		return nil, err
	}

	var result StorageContentResponse
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		s.Logger.Error("Error parsing response JSON: ", err)
		return nil, err
	}

	return result.Data, nil
}
