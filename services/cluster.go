package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/sirupsen/logrus"
)

// ClusterResource represents a cluster resource (VM, container, node, storage, etc.)
type ClusterResource struct {
	ID      string  `json:"id"`
	Type    string  `json:"type"`
	Node    string  `json:"node,omitempty"`
	Status  string  `json:"status,omitempty"`
	Name    string  `json:"name,omitempty"`
	VMID    int     `json:"vmid,omitempty"`
	MaxCPU  int     `json:"maxcpu,omitempty"`
	CPU     float64 `json:"cpu,omitempty"`
	MaxMem  int64   `json:"maxmem,omitempty"`
	Mem     int64   `json:"mem,omitempty"`
	MaxDisk int64   `json:"maxdisk,omitempty"`
	Disk    int64   `json:"disk,omitempty"`
	Uptime  int64   `json:"uptime,omitempty"`
	Level   string  `json:"level,omitempty"`
}

// ClusterStatus represents cluster status information
type ClusterStatus struct {
	Type    string `json:"type"`
	ID      string `json:"id,omitempty"`
	Name    string `json:"name"`
	Nodes   int    `json:"nodes,omitempty"`
	Quorate int    `json:"quorate,omitempty"`
	Version int    `json:"version,omitempty"`
	IP      string `json:"ip,omitempty"`
	Online  int    `json:"online,omitempty"`
	Local   int    `json:"local,omitempty"`
}

// ClusterResourcesResponse represents the API response for cluster resources
type ClusterResourcesResponse struct {
	Data []ClusterResource `json:"data"`
}

// ClusterStatusResponse represents the API response for cluster status
type ClusterStatusResponse struct {
	Data []ClusterStatus `json:"data"`
}

// ClusterService handles cluster-related operations
type ClusterService struct {
	Logger         *logrus.Logger
	Trust          bool
	HTTPService    HTTPServiceInterface
	SessionService SessionServiceInterface
}

// NewClusterService creates a new ClusterService with real dependencies
func NewClusterService(logger *logrus.Logger, trust bool) (*ClusterService, error) {
	sessionService, err := NewSessionService(logger)
	if err != nil {
		return nil, err
	}

	return &ClusterService{
		Logger:         logger,
		Trust:          trust,
		HTTPService:    NewHttpService(logger, trust),
		SessionService: sessionService,
	}, nil
}

// NewClusterServiceWithDeps creates a ClusterService with injected dependencies (for testing)
func NewClusterServiceWithDeps(logger *logrus.Logger, trust bool, httpService HTTPServiceInterface, sessionService SessionServiceInterface) *ClusterService {
	return &ClusterService{
		Logger:         logger,
		Trust:          trust,
		HTTPService:    httpService,
		SessionService: sessionService,
	}
}

// ListResources retrieves a list of all cluster resources
func (c *ClusterService) ListResources() ([]ClusterResource, error) {
	sessionData, err := c.SessionService.ReadSessionFile()
	if err != nil {
		c.Logger.Error("Error reading session file: ", err)
		return nil, err
	}

	uri := fmt.Sprintf("%s://%s:%d/api2/json/cluster/resources",
		sessionData.HttpScheme, sessionData.Server, sessionData.Port)

	cookies := []*http.Cookie{
		{
			Name:  "PVEAuthCookie",
			Value: url.QueryEscape(sessionData.Response.Data.Ticket),
		},
	}

	resp, err := c.HTTPService.Get(uri, nil, cookies)
	if err != nil {
		c.Logger.Error("Error listing cluster resources: ", err)
		return nil, err
	}
	//nolint:errcheck // Best effort close in defer
	defer func() { _ = resp.Body.Close() }()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		c.Logger.Error("Error reading response body: ", err)
		return nil, err
	}

	var result ClusterResourcesResponse
	err = json.Unmarshal(bodyBytes, &result)
	if err != nil {
		c.Logger.Error("Error parsing response JSON: ", err)
		return nil, err
	}

	return result.Data, nil
}

// GetStatus retrieves cluster status information
func (c *ClusterService) GetStatus() ([]ClusterStatus, error) {
	sessionData, err := c.SessionService.ReadSessionFile()
	if err != nil {
		c.Logger.Error("Error reading session file: ", err)
		return nil, err
	}

	uri := fmt.Sprintf("%s://%s:%d/api2/json/cluster/status",
		sessionData.HttpScheme, sessionData.Server, sessionData.Port)

	cookies := []*http.Cookie{
		{
			Name:  "PVEAuthCookie",
			Value: url.QueryEscape(sessionData.Response.Data.Ticket),
		},
	}

	resp, err := c.HTTPService.Get(uri, nil, cookies)
	if err != nil {
		c.Logger.Error("Error getting cluster status: ", err)
		return nil, err
	}
	//nolint:errcheck // Best effort close in defer
	defer func() { _ = resp.Body.Close() }()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		c.Logger.Error("Error reading response body: ", err)
		return nil, err
	}

	var result ClusterStatusResponse
	err = json.Unmarshal(bodyBytes, &result)
	if err != nil {
		c.Logger.Error("Error parsing response JSON: ", err)
		return nil, err
	}

	return result.Data, nil
}
