package tests

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"proxmox-cli/services"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNodesService_ListNodes_Success(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)

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
				Username:            "user@pam",
				Ticket:              "ticket123",
				CSRFPreventionToken: "csrf123",
			},
		},
	}

	mockHTTP := &mockHTTPService{
		getFunc: func(url string, headers map[string]string, cookies []*http.Cookie) (*http.Response, error) {
			body := `{"data": [
				{"node": "pve1", "status": "online", "cpu": 0.15, "maxcpu": 8, "mem": 8589934592, "maxmem": 17179869184, "uptime": 86400}
			]}`
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(body)),
			}, nil
		},
	}

	mockSession := &mockSessionService{
		readSessionFileFunc: func() (services.SessionData, error) {
			return validSession, nil
		},
	}

	nodesService := services.NewNodesServiceWithDeps(logger, true, mockHTTP, mockSession)
	nodes, err := nodesService.ListNodes()

	assert.NoError(t, err)
	assert.Len(t, nodes, 1)
	assert.Equal(t, "pve1", nodes[0].Node)
	assert.Equal(t, "online", nodes[0].Status)
	assert.Equal(t, 0.15, nodes[0].CPU)
}

func TestNodesService_ListNodes_SessionError(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	mockHTTP := &mockHTTPService{}
	mockSession := &mockSessionService{
		readSessionFileFunc: func() (services.SessionData, error) {
			return services.SessionData{}, assert.AnError
		},
	}

	nodesService := services.NewNodesServiceWithDeps(logger, true, mockHTTP, mockSession)
	nodes, err := nodesService.ListNodes()

	assert.Error(t, err)
	assert.Nil(t, nodes)
}

func TestNodesService_ListNodes_HttpError(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)

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
				Username:            "user@pam",
				Ticket:              "ticket123",
				CSRFPreventionToken: "csrf123",
			},
		},
	}

	mockHTTP := &mockHTTPService{
		getFunc: func(url string, headers map[string]string, cookies []*http.Cookie) (*http.Response, error) {
			return nil, assert.AnError
		},
	}

	mockSession := &mockSessionService{
		readSessionFileFunc: func() (services.SessionData, error) {
			return validSession, nil
		},
	}

	nodesService := services.NewNodesServiceWithDeps(logger, true, mockHTTP, mockSession)
	nodes, err := nodesService.ListNodes()

	assert.Error(t, err)
	assert.Nil(t, nodes)
}

func TestNodesService_GetNodeStatus_Success(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)

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
				Username:            "user@pam",
				Ticket:              "ticket123",
				CSRFPreventionToken: "csrf123",
			},
		},
	}

	mockHTTP := &mockHTTPService{
		getFunc: func(url string, headers map[string]string, cookies []*http.Cookie) (*http.Response, error) {
			body := `{"data": {
				"cpu": 0.25,
				"cpuinfo": {"cpus": 8, "model": "Intel Core i7", "sockets": 1, "mhz": "2400"},
				"memory": {"used": 8589934592, "total": 17179869184, "free": 8589934592},
				"rootfs": {"used": 10737418240, "total": 107374182400, "avail": 96636764160},
				"swap": {"used": 0, "total": 8589934592, "free": 8589934592},
				"uptime": 86400,
				"loadavg": [0.5, 0.4, 0.3],
				"kversion": "Linux 5.15.0",
				"wait": 0.01,
				"pveversion": "pve-manager/7.0-1"
			}}`
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(body)),
			}, nil
		},
	}

	mockSession := &mockSessionService{
		readSessionFileFunc: func() (services.SessionData, error) {
			return validSession, nil
		},
	}

	nodesService := services.NewNodesServiceWithDeps(logger, true, mockHTTP, mockSession)
	status, err := nodesService.GetNodeStatus("pve1")

	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.Equal(t, 0.25, status.CPU)
	assert.Equal(t, 8, status.CPUInfo.CPUs)
	assert.Equal(t, int64(86400), status.Uptime)
}

func TestNodesService_GetNodeStatus_HttpError(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)

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
				Username:            "user@pam",
				Ticket:              "ticket123",
				CSRFPreventionToken: "csrf123",
			},
		},
	}

	mockHTTP := &mockHTTPService{
		getFunc: func(url string, headers map[string]string, cookies []*http.Cookie) (*http.Response, error) {
			return nil, assert.AnError
		},
	}

	mockSession := &mockSessionService{
		readSessionFileFunc: func() (services.SessionData, error) {
			return validSession, nil
		},
	}

	nodesService := services.NewNodesServiceWithDeps(logger, true, mockHTTP, mockSession)
	status, err := nodesService.GetNodeStatus("pve1")

	assert.Error(t, err)
	assert.Nil(t, status)
}

func TestNodesService_GetNodeVersion_Success(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)

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
				Username:            "user@pam",
				Ticket:              "ticket123",
				CSRFPreventionToken: "csrf123",
			},
		},
	}

	mockHTTP := &mockHTTPService{
		getFunc: func(url string, headers map[string]string, cookies []*http.Cookie) (*http.Response, error) {
			body := `{"data": {
				"version": "7.0",
				"release": "7.0-1",
				"repoid": "abc123def"
			}}`
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(body)),
			}, nil
		},
	}

	mockSession := &mockSessionService{
		readSessionFileFunc: func() (services.SessionData, error) {
			return validSession, nil
		},
	}

	nodesService := services.NewNodesServiceWithDeps(logger, true, mockHTTP, mockSession)
	version, err := nodesService.GetNodeVersion("pve1")

	assert.NoError(t, err)
	assert.NotNil(t, version)
	assert.Equal(t, "7.0", version.Version)
	assert.Equal(t, "7.0-1", version.Release)
	assert.Equal(t, "abc123def", version.RepoID)
}

func TestNodesService_GetNodeVersion_InvalidJSON(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)

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
				Username:            "user@pam",
				Ticket:              "ticket123",
				CSRFPreventionToken: "csrf123",
			},
		},
	}

	mockHTTP := &mockHTTPService{
		getFunc: func(url string, headers map[string]string, cookies []*http.Cookie) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader("invalid json")),
			}, nil
		},
	}

	mockSession := &mockSessionService{
		readSessionFileFunc: func() (services.SessionData, error) {
			return validSession, nil
		},
	}

	nodesService := services.NewNodesServiceWithDeps(logger, true, mockHTTP, mockSession)
	version, err := nodesService.GetNodeVersion("pve1")

	assert.Error(t, err)
	assert.Nil(t, version)
}
