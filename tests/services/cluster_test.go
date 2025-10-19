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

func TestClusterService_ListResources_Success(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	mockHTTP := &mockHttpService{
		getFunc: func(url string, headers map[string]string, cookies []*http.Cookie) (*http.Response, error) {
			body := `{"data": [
				{
					"id": "node/pve1",
					"type": "node",
					"node": "pve1",
					"status": "online",
					"maxcpu": 8,
					"cpu": 0.15,
					"maxmem": 17179869184,
					"mem": 8589934592,
					"uptime": 86400,
					"level": ""
				},
				{
					"id": "qemu/100",
					"type": "qemu",
					"node": "pve1",
					"status": "running",
					"name": "test-vm",
					"vmid": 100,
					"maxcpu": 2,
					"cpu": 0.25,
					"maxmem": 4294967296,
					"mem": 2147483648,
					"maxdisk": 34359738368,
					"disk": 10737418240,
					"uptime": 3600
				},
				{
					"id": "storage/pve1/local",
					"type": "storage",
					"node": "pve1",
					"status": "available"
				}
			]}`
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(body)),
			}, nil
		},
	}

	mockSession := &mockSessionService{
		readSessionFileFunc: func() (services.SessionData, error) {
			return getValidSessionData(), nil
		},
	}

	clusterService := services.NewClusterServiceWithDeps(logger, true, mockHTTP, mockSession)
	resources, err := clusterService.ListResources()

	assert.NoError(t, err)
	assert.Len(t, resources, 3)

	// Check node resource
	assert.Equal(t, "node/pve1", resources[0].ID)
	assert.Equal(t, "node", resources[0].Type)
	assert.Equal(t, "pve1", resources[0].Node)
	assert.Equal(t, "online", resources[0].Status)
	assert.Equal(t, 8, resources[0].MaxCPU)

	// Check VM resource
	assert.Equal(t, "qemu/100", resources[1].ID)
	assert.Equal(t, "qemu", resources[1].Type)
	assert.Equal(t, "test-vm", resources[1].Name)
	assert.Equal(t, 100, resources[1].VMID)
	assert.Equal(t, "running", resources[1].Status)

	// Check storage resource
	assert.Equal(t, "storage/pve1/local", resources[2].ID)
	assert.Equal(t, "storage", resources[2].Type)
	assert.Equal(t, "available", resources[2].Status)
}

func TestClusterService_ListResources_SessionError(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	mockHTTP := &mockHttpService{}
	mockSession := &mockSessionService{
		readSessionFileFunc: func() (services.SessionData, error) {
			return services.SessionData{}, assert.AnError
		},
	}

	clusterService := services.NewClusterServiceWithDeps(logger, true, mockHTTP, mockSession)
	resources, err := clusterService.ListResources()

	assert.Error(t, err)
	assert.Nil(t, resources)
}

func TestClusterService_ListResources_HttpError(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	mockHTTP := &mockHttpService{
		getFunc: func(url string, headers map[string]string, cookies []*http.Cookie) (*http.Response, error) {
			return nil, assert.AnError
		},
	}

	mockSession := &mockSessionService{
		readSessionFileFunc: func() (services.SessionData, error) {
			return getValidSessionData(), nil
		},
	}

	clusterService := services.NewClusterServiceWithDeps(logger, true, mockHTTP, mockSession)
	resources, err := clusterService.ListResources()

	assert.Error(t, err)
	assert.Nil(t, resources)
}

func TestClusterService_ListResources_InvalidJSON(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	mockHTTP := &mockHttpService{
		getFunc: func(url string, headers map[string]string, cookies []*http.Cookie) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader("invalid json")),
			}, nil
		},
	}

	mockSession := &mockSessionService{
		readSessionFileFunc: func() (services.SessionData, error) {
			return getValidSessionData(), nil
		},
	}

	clusterService := services.NewClusterServiceWithDeps(logger, true, mockHTTP, mockSession)
	resources, err := clusterService.ListResources()

	assert.Error(t, err)
	assert.Nil(t, resources)
}

func TestClusterService_GetStatus_Success(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	mockHTTP := &mockHttpService{
		getFunc: func(url string, headers map[string]string, cookies []*http.Cookie) (*http.Response, error) {
			body := `{"data": [
				{
					"type": "cluster",
					"id": "cluster",
					"name": "mycluster",
					"nodes": 3,
					"quorate": 1,
					"version": 15
				},
				{
					"type": "node",
					"id": "node/pve1",
					"name": "pve1",
					"ip": "192.168.1.10",
					"online": 1,
					"local": 1
				},
				{
					"type": "node",
					"id": "node/pve2",
					"name": "pve2",
					"ip": "192.168.1.11",
					"online": 1,
					"local": 0
				}
			]}`
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(body)),
			}, nil
		},
	}

	mockSession := &mockSessionService{
		readSessionFileFunc: func() (services.SessionData, error) {
			return getValidSessionData(), nil
		},
	}

	clusterService := services.NewClusterServiceWithDeps(logger, true, mockHTTP, mockSession)
	statuses, err := clusterService.GetStatus()

	assert.NoError(t, err)
	assert.Len(t, statuses, 3)

	// Check cluster status
	assert.Equal(t, "cluster", statuses[0].Type)
	assert.Equal(t, "mycluster", statuses[0].Name)
	assert.Equal(t, 3, statuses[0].Nodes)
	assert.Equal(t, 1, statuses[0].Quorate)
	assert.Equal(t, 15, statuses[0].Version)

	// Check node 1 status
	assert.Equal(t, "node", statuses[1].Type)
	assert.Equal(t, "pve1", statuses[1].Name)
	assert.Equal(t, "192.168.1.10", statuses[1].IP)
	assert.Equal(t, 1, statuses[1].Online)
	assert.Equal(t, 1, statuses[1].Local)

	// Check node 2 status
	assert.Equal(t, "node", statuses[2].Type)
	assert.Equal(t, "pve2", statuses[2].Name)
	assert.Equal(t, "192.168.1.11", statuses[2].IP)
	assert.Equal(t, 1, statuses[2].Online)
	assert.Equal(t, 0, statuses[2].Local)
}

func TestClusterService_GetStatus_SessionError(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	mockHTTP := &mockHttpService{}
	mockSession := &mockSessionService{
		readSessionFileFunc: func() (services.SessionData, error) {
			return services.SessionData{}, assert.AnError
		},
	}

	clusterService := services.NewClusterServiceWithDeps(logger, true, mockHTTP, mockSession)
	statuses, err := clusterService.GetStatus()

	assert.Error(t, err)
	assert.Nil(t, statuses)
}

func TestClusterService_GetStatus_HttpError(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	mockHTTP := &mockHttpService{
		getFunc: func(url string, headers map[string]string, cookies []*http.Cookie) (*http.Response, error) {
			return nil, assert.AnError
		},
	}

	mockSession := &mockSessionService{
		readSessionFileFunc: func() (services.SessionData, error) {
			return getValidSessionData(), nil
		},
	}

	clusterService := services.NewClusterServiceWithDeps(logger, true, mockHTTP, mockSession)
	statuses, err := clusterService.GetStatus()

	assert.Error(t, err)
	assert.Nil(t, statuses)
}

func TestClusterService_GetStatus_InvalidJSON(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	mockHTTP := &mockHttpService{
		getFunc: func(url string, headers map[string]string, cookies []*http.Cookie) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader("not valid json")),
			}, nil
		},
	}

	mockSession := &mockSessionService{
		readSessionFileFunc: func() (services.SessionData, error) {
			return getValidSessionData(), nil
		},
	}

	clusterService := services.NewClusterServiceWithDeps(logger, true, mockHTTP, mockSession)
	statuses, err := clusterService.GetStatus()

	assert.Error(t, err)
	assert.Nil(t, statuses)
}

func TestClusterService_GetStatus_EmptyData(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	mockHTTP := &mockHttpService{
		getFunc: func(url string, headers map[string]string, cookies []*http.Cookie) (*http.Response, error) {
			body := `{"data": []}`
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(body)),
			}, nil
		},
	}

	mockSession := &mockSessionService{
		readSessionFileFunc: func() (services.SessionData, error) {
			return getValidSessionData(), nil
		},
	}

	clusterService := services.NewClusterServiceWithDeps(logger, true, mockHTTP, mockSession)
	statuses, err := clusterService.GetStatus()

	assert.NoError(t, err)
	assert.Len(t, statuses, 0)
}
