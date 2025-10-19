package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/sirupsen/logrus"
)

// VM represents a virtual machine in Proxmox
type VM struct {
	VMID    int     `json:"vmid"`
	Name    string  `json:"name"`
	Status  string  `json:"status"`
	CPU     float64 `json:"cpu,omitempty"`
	CPUs    int     `json:"cpus,omitempty"`
	MaxCPU  int     `json:"maxcpu,omitempty"`
	Mem     int64   `json:"mem,omitempty"`
	MaxMem  int64   `json:"maxmem,omitempty"`
	Disk    int64   `json:"disk,omitempty"`
	MaxDisk int64   `json:"maxdisk,omitempty"`
	Uptime  int64   `json:"uptime,omitempty"`
	Node    string  `json:"node,omitempty"`
}

// VMConfig represents VM configuration
type VMConfig struct {
	Name        string `json:"name,omitempty"`
	Memory      int    `json:"memory,omitempty"`
	Cores       int    `json:"cores,omitempty"`
	Sockets     int    `json:"sockets,omitempty"`
	OSType      string `json:"ostype,omitempty"`
	Boot        string `json:"boot,omitempty"`
	Bootdisk    string `json:"bootdisk,omitempty"`
	Description string `json:"description,omitempty"`
}

// VMStatus represents VM status details
type VMStatus struct {
	Status    string  `json:"status"`
	VMID      int     `json:"vmid"`
	CPU       float64 `json:"cpu,omitempty"`
	CPUs      int     `json:"cpus,omitempty"`
	Mem       int64   `json:"mem,omitempty"`
	MaxMem    int64   `json:"maxmem,omitempty"`
	Uptime    int64   `json:"uptime,omitempty"`
	Name      string  `json:"name,omitempty"`
	QMPStatus string  `json:"qmpstatus,omitempty"`
}

// VMListResponse represents the API response for VM list
type VMListResponse struct {
	Data []VM `json:"data"`
}

// VMStatusResponse represents the API response for VM status
type VMStatusResponse struct {
	Data VMStatus `json:"data"`
}

// VMCreateResponse represents the API response for VM creation
type VMCreateResponse struct {
	Data string `json:"data"` // UPID task ID
}

// VMService handles VM-related operations
type VMService struct {
	Logger         *logrus.Logger
	Trust          bool
	HttpService    HttpServiceInterface
	SessionService SessionServiceInterface
}

// NewVMService creates a new VMService with real dependencies
func NewVMService(logger *logrus.Logger, trust bool) (*VMService, error) {
	sessionService, err := NewSessionService(logger)
	if err != nil {
		return nil, err
	}

	return &VMService{
		Logger:         logger,
		Trust:          trust,
		HttpService:    NewHttpService(logger, trust),
		SessionService: sessionService,
	}, nil
}

// NewVMServiceWithDeps creates a VMService with injected dependencies (for testing)
func NewVMServiceWithDeps(logger *logrus.Logger, trust bool, httpService HttpServiceInterface, sessionService SessionServiceInterface) *VMService {
	return &VMService{
		Logger:         logger,
		Trust:          trust,
		HttpService:    httpService,
		SessionService: sessionService,
	}
}

// ListVMs retrieves a list of all VMs on a specific node
func (v *VMService) ListVMs(nodeName string) ([]VM, error) {
	sessionData, err := v.SessionService.ReadSessionFile()
	if err != nil {
		v.Logger.Error("Error reading session file: ", err)
		return nil, err
	}

	uri := fmt.Sprintf("%s://%s:%d/api2/json/nodes/%s/qemu",
		sessionData.HttpScheme, sessionData.Server, sessionData.Port, nodeName)

	cookies := []*http.Cookie{
		{
			Name:  "PVEAuthCookie",
			Value: url.QueryEscape(sessionData.Response.Data.Ticket),
		},
	}

	resp, err := v.HttpService.Get(uri, nil, cookies)
	if err != nil {
		v.Logger.Error("Error listing VMs: ", err)
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		v.Logger.Error("Error reading response body: ", err)
		return nil, err
	}

	var result VMListResponse
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		v.Logger.Error("Error parsing response JSON: ", err)
		return nil, err
	}

	return result.Data, nil
}

// GetVMStatus retrieves the current status of a specific VM
func (v *VMService) GetVMStatus(nodeName string, vmid int) (*VMStatus, error) {
	sessionData, err := v.SessionService.ReadSessionFile()
	if err != nil {
		v.Logger.Error("Error reading session file: ", err)
		return nil, err
	}

	uri := fmt.Sprintf("%s://%s:%d/api2/json/nodes/%s/qemu/%d/status/current",
		sessionData.HttpScheme, sessionData.Server, sessionData.Port, nodeName, vmid)

	cookies := []*http.Cookie{
		{
			Name:  "PVEAuthCookie",
			Value: url.QueryEscape(sessionData.Response.Data.Ticket),
		},
	}

	resp, err := v.HttpService.Get(uri, nil, cookies)
	if err != nil {
		v.Logger.Error("Error getting VM status: ", err)
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		v.Logger.Error("Error reading response body: ", err)
		return nil, err
	}

	var result VMStatusResponse
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		v.Logger.Error("Error parsing response JSON: ", err)
		return nil, err
	}

	return &result.Data, nil
}

// StartVM starts a VM
func (v *VMService) StartVM(nodeName string, vmid int) (string, error) {
	return v.vmStatusAction(nodeName, vmid, "start")
}

// StopVM stops a VM
func (v *VMService) StopVM(nodeName string, vmid int) (string, error) {
	return v.vmStatusAction(nodeName, vmid, "stop")
}

// ShutdownVM gracefully shuts down a VM
func (v *VMService) ShutdownVM(nodeName string, vmid int) (string, error) {
	return v.vmStatusAction(nodeName, vmid, "shutdown")
}

// RebootVM reboots a VM
func (v *VMService) RebootVM(nodeName string, vmid int) (string, error) {
	return v.vmStatusAction(nodeName, vmid, "reboot")
}

// ResetVM resets a VM
func (v *VMService) ResetVM(nodeName string, vmid int) (string, error) {
	return v.vmStatusAction(nodeName, vmid, "reset")
}

// SuspendVM suspends a VM
func (v *VMService) SuspendVM(nodeName string, vmid int) (string, error) {
	return v.vmStatusAction(nodeName, vmid, "suspend")
}

// ResumeVM resumes a suspended VM
func (v *VMService) ResumeVM(nodeName string, vmid int) (string, error) {
	return v.vmStatusAction(nodeName, vmid, "resume")
}

// vmStatusAction performs a status action on a VM (start, stop, etc.)
func (v *VMService) vmStatusAction(nodeName string, vmid int, action string) (string, error) {
	sessionData, err := v.SessionService.ReadSessionFile()
	if err != nil {
		v.Logger.Error("Error reading session file: ", err)
		return "", err
	}

	uri := fmt.Sprintf("%s://%s:%d/api2/json/nodes/%s/qemu/%d/status/%s",
		sessionData.HttpScheme, sessionData.Server, sessionData.Port, nodeName, vmid, action)

	headers := map[string]string{
		"Content-Type":        "application/x-www-form-urlencoded; charset=UTF-8",
		"CSRFPreventionToken": sessionData.Response.Data.CSRFPreventionToken,
	}

	cookies := []*http.Cookie{
		{
			Name:  "PVEAuthCookie",
			Value: url.QueryEscape(sessionData.Response.Data.Ticket),
		},
	}

	body, err := v.HttpService.Post(uri, "", headers, cookies)
	if err != nil {
		v.Logger.Error(fmt.Sprintf("Error performing %s on VM: ", action), err)
		return "", err
	}

	var result VMCreateResponse
	if err := json.Unmarshal([]byte(body), &result); err != nil {
		v.Logger.Error("Error parsing response JSON: ", err)
		return "", err
	}

	return result.Data, nil
}

// DeleteVM deletes a VM
func (v *VMService) DeleteVM(nodeName string, vmid int) (string, error) {
	sessionData, err := v.SessionService.ReadSessionFile()
	if err != nil {
		v.Logger.Error("Error reading session file: ", err)
		return "", err
	}

	uri := fmt.Sprintf("%s://%s:%d/api2/json/nodes/%s/qemu/%d",
		sessionData.HttpScheme, sessionData.Server, sessionData.Port, nodeName, vmid)

	headers := map[string]string{
		"CSRFPreventionToken": sessionData.Response.Data.CSRFPreventionToken,
	}

	cookies := []*http.Cookie{
		{
			Name:  "PVEAuthCookie",
			Value: url.QueryEscape(sessionData.Response.Data.Ticket),
		},
	}

	body, err := v.HttpService.Delete(uri, headers, cookies)
	if err != nil {
		v.Logger.Error("Error deleting VM: ", err)
		return "", err
	}

	var result VMCreateResponse
	if err := json.Unmarshal([]byte(body), &result); err != nil {
		v.Logger.Error("Error parsing response JSON: ", err)
		return "", err
	}

	return result.Data, nil
}
