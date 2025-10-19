package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/sirupsen/logrus"
)

// Node represents a Proxmox cluster node
type Node struct {
	Node   string  `json:"node"`
	Status string  `json:"status"`
	CPU    float64 `json:"cpu,omitempty"`
	MaxCPU int     `json:"maxcpu,omitempty"`
	Mem    int64   `json:"mem,omitempty"`
	MaxMem int64   `json:"maxmem,omitempty"`
	Uptime int64   `json:"uptime,omitempty"`
	Level  string  `json:"level,omitempty"`
}

// NodeStatus represents detailed node status information
type NodeStatus struct {
	CPU        float64        `json:"cpu"`
	CPUInfo    NodeCPUInfo    `json:"cpuinfo"`
	Memory     NodeMemoryInfo `json:"memory"`
	RootFS     NodeRootFSInfo `json:"rootfs"`
	Swap       NodeSwapInfo   `json:"swap"`
	Uptime     int64          `json:"uptime"`
	LoadAvg    []float64      `json:"loadavg"`
	KVersion   string         `json:"kversion"`
	Wait       float64        `json:"wait"`
	PVEVersion string         `json:"pveversion"`
}

type NodeCPUInfo struct {
	CPUs    int    `json:"cpus"`
	Model   string `json:"model"`
	Sockets int    `json:"sockets"`
	MHZ     string `json:"mhz"`
}

type NodeMemoryInfo struct {
	Used  int64 `json:"used"`
	Total int64 `json:"total"`
	Free  int64 `json:"free"`
}

type NodeRootFSInfo struct {
	Used  int64 `json:"used"`
	Total int64 `json:"total"`
	Avail int64 `json:"avail"`
}

type NodeSwapInfo struct {
	Used  int64 `json:"used"`
	Total int64 `json:"total"`
	Free  int64 `json:"free"`
}

// NodeVersion represents version information
type NodeVersion struct {
	Version string `json:"version"`
	Release string `json:"release"`
	RepoID  string `json:"repoid"`
}

// NodeListResponse represents the API response for node list
type NodeListResponse struct {
	Data []Node `json:"data"`
}

// NodeStatusResponse represents the API response for node status
type NodeStatusResponse struct {
	Data NodeStatus `json:"data"`
}

// NodeVersionResponse represents the API response for node version
type NodeVersionResponse struct {
	Data NodeVersion `json:"data"`
}

// NodesService handles node-related operations
type NodesService struct {
	Logger         *logrus.Logger
	Trust          bool
	HttpService    HttpServiceInterface
	SessionService SessionServiceInterface
}

// NewNodesService creates a new NodesService with real dependencies
func NewNodesService(logger *logrus.Logger, trust bool) (*NodesService, error) {
	sessionService, err := NewSessionService(logger)
	if err != nil {
		return nil, err
	}

	return &NodesService{
		Logger:         logger,
		Trust:          trust,
		HttpService:    NewHttpService(logger, trust),
		SessionService: sessionService,
	}, nil
}

// NewNodesServiceWithDeps creates a NodesService with injected dependencies (for testing)
func NewNodesServiceWithDeps(logger *logrus.Logger, trust bool, httpService HttpServiceInterface, sessionService SessionServiceInterface) *NodesService {
	return &NodesService{
		Logger:         logger,
		Trust:          trust,
		HttpService:    httpService,
		SessionService: sessionService,
	}
}

// ListNodes retrieves a list of all nodes in the cluster
func (n *NodesService) ListNodes() ([]Node, error) {
	sessionData, err := n.SessionService.ReadSessionFile()
	if err != nil {
		n.Logger.Error("Error reading session file: ", err)
		return nil, err
	}

	uri := fmt.Sprintf("%s://%s:%d/api2/json/nodes",
		sessionData.HttpScheme, sessionData.Server, sessionData.Port)

	cookies := []*http.Cookie{
		{
			Name:  "PVEAuthCookie",
			Value: url.QueryEscape(sessionData.Response.Data.Ticket),
		},
	}

	resp, err := n.HttpService.Get(uri, nil, cookies)
	if err != nil {
		n.Logger.Error("Error listing nodes: ", err)
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		n.Logger.Error("Error reading response body: ", err)
		return nil, err
	}

	var result NodeListResponse
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		n.Logger.Error("Error parsing response JSON: ", err)
		return nil, err
	}

	return result.Data, nil
}

// GetNodeStatus retrieves detailed status information for a specific node
func (n *NodesService) GetNodeStatus(nodeName string) (*NodeStatus, error) {
	sessionData, err := n.SessionService.ReadSessionFile()
	if err != nil {
		n.Logger.Error("Error reading session file: ", err)
		return nil, err
	}

	uri := fmt.Sprintf("%s://%s:%d/api2/json/nodes/%s/status",
		sessionData.HttpScheme, sessionData.Server, sessionData.Port, nodeName)

	cookies := []*http.Cookie{
		{
			Name:  "PVEAuthCookie",
			Value: url.QueryEscape(sessionData.Response.Data.Ticket),
		},
	}

	resp, err := n.HttpService.Get(uri, nil, cookies)
	if err != nil {
		n.Logger.Error("Error getting node status: ", err)
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		n.Logger.Error("Error reading response body: ", err)
		return nil, err
	}

	var result NodeStatusResponse
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		n.Logger.Error("Error parsing response JSON: ", err)
		return nil, err
	}

	return &result.Data, nil
}

// GetNodeVersion retrieves version information for a specific node
func (n *NodesService) GetNodeVersion(nodeName string) (*NodeVersion, error) {
	sessionData, err := n.SessionService.ReadSessionFile()
	if err != nil {
		n.Logger.Error("Error reading session file: ", err)
		return nil, err
	}

	uri := fmt.Sprintf("%s://%s:%d/api2/json/nodes/%s/version",
		sessionData.HttpScheme, sessionData.Server, sessionData.Port, nodeName)

	cookies := []*http.Cookie{
		{
			Name:  "PVEAuthCookie",
			Value: url.QueryEscape(sessionData.Response.Data.Ticket),
		},
	}

	resp, err := n.HttpService.Get(uri, nil, cookies)
	if err != nil {
		n.Logger.Error("Error getting node version: ", err)
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		n.Logger.Error("Error reading response body: ", err)
		return nil, err
	}

	var result NodeVersionResponse
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		n.Logger.Error("Error parsing response JSON: ", err)
		return nil, err
	}

	return &result.Data, nil
}
