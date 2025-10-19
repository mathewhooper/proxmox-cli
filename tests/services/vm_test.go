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

func getValidSessionData() services.SessionData {
	return services.SessionData{
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
}

func TestVMService_ListVMs_Success(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	mockHTTP := &mockHttpService{
		getFunc: func(url string, headers map[string]string, cookies []*http.Cookie) (*http.Response, error) {
			body := `{"data": [
				{
					"vmid": 100,
					"name": "test-vm",
					"status": "running",
					"cpu": 0.25,
					"cpus": 2,
					"maxcpu": 2,
					"mem": 2147483648,
					"maxmem": 4294967296,
					"uptime": 3600
				},
				{
					"vmid": 101,
					"name": "test-vm2",
					"status": "stopped",
					"cpu": 0.0,
					"cpus": 4,
					"maxcpu": 4,
					"mem": 0,
					"maxmem": 8589934592,
					"uptime": 0
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

	vmService := services.NewVMServiceWithDeps(logger, true, mockHTTP, mockSession)
	vms, err := vmService.ListVMs("pve1")

	assert.NoError(t, err)
	assert.Len(t, vms, 2)
	assert.Equal(t, 100, vms[0].VMID)
	assert.Equal(t, "test-vm", vms[0].Name)
	assert.Equal(t, "running", vms[0].Status)
	assert.Equal(t, 101, vms[1].VMID)
	assert.Equal(t, "stopped", vms[1].Status)
}

func TestVMService_ListVMs_SessionError(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	mockHTTP := &mockHttpService{}
	mockSession := &mockSessionService{
		readSessionFileFunc: func() (services.SessionData, error) {
			return services.SessionData{}, assert.AnError
		},
	}

	vmService := services.NewVMServiceWithDeps(logger, true, mockHTTP, mockSession)
	vms, err := vmService.ListVMs("pve1")

	assert.Error(t, err)
	assert.Nil(t, vms)
}

func TestVMService_ListVMs_HttpError(t *testing.T) {
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

	vmService := services.NewVMServiceWithDeps(logger, true, mockHTTP, mockSession)
	vms, err := vmService.ListVMs("pve1")

	assert.Error(t, err)
	assert.Nil(t, vms)
}

func TestVMService_GetVMStatus_Success(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	mockHTTP := &mockHttpService{
		getFunc: func(url string, headers map[string]string, cookies []*http.Cookie) (*http.Response, error) {
			body := `{"data": {
				"status": "running",
				"vmid": 100,
				"cpu": 0.15,
				"cpus": 2,
				"mem": 2147483648,
				"maxmem": 4294967296,
				"uptime": 3600,
				"name": "test-vm",
				"qmpstatus": "running"
			}}`
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

	vmService := services.NewVMServiceWithDeps(logger, true, mockHTTP, mockSession)
	status, err := vmService.GetVMStatus("pve1", 100)

	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.Equal(t, "running", status.Status)
	assert.Equal(t, 100, status.VMID)
	assert.Equal(t, "test-vm", status.Name)
	assert.Equal(t, "running", status.QMPStatus)
}

func TestVMService_StartVM_Success(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	mockHTTP := &mockHttpService{
		postFunc: func(uri, payload string, headers map[string]string, cookies []*http.Cookie) (string, error) {
			return `{"data": "UPID:pve1:00001234:12345678:5F123456:qmstart:100:user@pam:"}`, nil
		},
	}

	mockSession := &mockSessionService{
		readSessionFileFunc: func() (services.SessionData, error) {
			return getValidSessionData(), nil
		},
	}

	vmService := services.NewVMServiceWithDeps(logger, true, mockHTTP, mockSession)
	taskID, err := vmService.StartVM("pve1", 100)

	assert.NoError(t, err)
	assert.NotEmpty(t, taskID)
	assert.Contains(t, taskID, "UPID:")
}

func TestVMService_StopVM_Success(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	mockHTTP := &mockHttpService{
		postFunc: func(uri, payload string, headers map[string]string, cookies []*http.Cookie) (string, error) {
			return `{"data": "UPID:pve1:00001234:12345678:5F123456:qmstop:100:user@pam:"}`, nil
		},
	}

	mockSession := &mockSessionService{
		readSessionFileFunc: func() (services.SessionData, error) {
			return getValidSessionData(), nil
		},
	}

	vmService := services.NewVMServiceWithDeps(logger, true, mockHTTP, mockSession)
	taskID, err := vmService.StopVM("pve1", 100)

	assert.NoError(t, err)
	assert.NotEmpty(t, taskID)
	assert.Contains(t, taskID, "UPID:")
}

func TestVMService_ShutdownVM_Success(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	mockHTTP := &mockHttpService{
		postFunc: func(uri, payload string, headers map[string]string, cookies []*http.Cookie) (string, error) {
			return `{"data": "UPID:pve1:00001234:12345678:5F123456:qmshutdown:100:user@pam:"}`, nil
		},
	}

	mockSession := &mockSessionService{
		readSessionFileFunc: func() (services.SessionData, error) {
			return getValidSessionData(), nil
		},
	}

	vmService := services.NewVMServiceWithDeps(logger, true, mockHTTP, mockSession)
	taskID, err := vmService.ShutdownVM("pve1", 100)

	assert.NoError(t, err)
	assert.NotEmpty(t, taskID)
}

func TestVMService_RebootVM_Success(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	mockHTTP := &mockHttpService{
		postFunc: func(uri, payload string, headers map[string]string, cookies []*http.Cookie) (string, error) {
			return `{"data": "UPID:pve1:00001234:12345678:5F123456:qmreboot:100:user@pam:"}`, nil
		},
	}

	mockSession := &mockSessionService{
		readSessionFileFunc: func() (services.SessionData, error) {
			return getValidSessionData(), nil
		},
	}

	vmService := services.NewVMServiceWithDeps(logger, true, mockHTTP, mockSession)
	taskID, err := vmService.RebootVM("pve1", 100)

	assert.NoError(t, err)
	assert.NotEmpty(t, taskID)
}

func TestVMService_ResetVM_Success(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	mockHTTP := &mockHttpService{
		postFunc: func(uri, payload string, headers map[string]string, cookies []*http.Cookie) (string, error) {
			return `{"data": "UPID:pve1:00001234:12345678:5F123456:qmreset:100:user@pam:"}`, nil
		},
	}

	mockSession := &mockSessionService{
		readSessionFileFunc: func() (services.SessionData, error) {
			return getValidSessionData(), nil
		},
	}

	vmService := services.NewVMServiceWithDeps(logger, true, mockHTTP, mockSession)
	taskID, err := vmService.ResetVM("pve1", 100)

	assert.NoError(t, err)
	assert.NotEmpty(t, taskID)
}

func TestVMService_SuspendVM_Success(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	mockHTTP := &mockHttpService{
		postFunc: func(uri, payload string, headers map[string]string, cookies []*http.Cookie) (string, error) {
			return `{"data": "UPID:pve1:00001234:12345678:5F123456:qmsuspend:100:user@pam:"}`, nil
		},
	}

	mockSession := &mockSessionService{
		readSessionFileFunc: func() (services.SessionData, error) {
			return getValidSessionData(), nil
		},
	}

	vmService := services.NewVMServiceWithDeps(logger, true, mockHTTP, mockSession)
	taskID, err := vmService.SuspendVM("pve1", 100)

	assert.NoError(t, err)
	assert.NotEmpty(t, taskID)
}

func TestVMService_ResumeVM_Success(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	mockHTTP := &mockHttpService{
		postFunc: func(uri, payload string, headers map[string]string, cookies []*http.Cookie) (string, error) {
			return `{"data": "UPID:pve1:00001234:12345678:5F123456:qmresume:100:user@pam:"}`, nil
		},
	}

	mockSession := &mockSessionService{
		readSessionFileFunc: func() (services.SessionData, error) {
			return getValidSessionData(), nil
		},
	}

	vmService := services.NewVMServiceWithDeps(logger, true, mockHTTP, mockSession)
	taskID, err := vmService.ResumeVM("pve1", 100)

	assert.NoError(t, err)
	assert.NotEmpty(t, taskID)
}

func TestVMService_DeleteVM_Success(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	mockHTTP := &mockHttpService{
		deleteFunc: func(url string, headers map[string]string, cookies []*http.Cookie) (string, error) {
			return `{"data": "UPID:pve1:00001234:12345678:5F123456:qmdestroy:100:user@pam:"}`, nil
		},
	}

	mockSession := &mockSessionService{
		readSessionFileFunc: func() (services.SessionData, error) {
			return getValidSessionData(), nil
		},
	}

	vmService := services.NewVMServiceWithDeps(logger, true, mockHTTP, mockSession)
	taskID, err := vmService.DeleteVM("pve1", 100)

	assert.NoError(t, err)
	assert.NotEmpty(t, taskID)
	assert.Contains(t, taskID, "UPID:")
}

func TestVMService_DeleteVM_HttpError(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	mockHTTP := &mockHttpService{
		deleteFunc: func(url string, headers map[string]string, cookies []*http.Cookie) (string, error) {
			return "", assert.AnError
		},
	}

	mockSession := &mockSessionService{
		readSessionFileFunc: func() (services.SessionData, error) {
			return getValidSessionData(), nil
		},
	}

	vmService := services.NewVMServiceWithDeps(logger, true, mockHTTP, mockSession)
	taskID, err := vmService.DeleteVM("pve1", 100)

	assert.Error(t, err)
	assert.Empty(t, taskID)
}

func TestVMService_StartVM_SessionError(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	mockHTTP := &mockHttpService{}
	mockSession := &mockSessionService{
		readSessionFileFunc: func() (services.SessionData, error) {
			return services.SessionData{}, assert.AnError
		},
	}

	vmService := services.NewVMServiceWithDeps(logger, true, mockHTTP, mockSession)
	taskID, err := vmService.StartVM("pve1", 100)

	assert.Error(t, err)
	assert.Empty(t, taskID)
}

func TestVMService_StartVM_HttpError(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	mockHTTP := &mockHttpService{
		postFunc: func(uri, payload string, headers map[string]string, cookies []*http.Cookie) (string, error) {
			return "", assert.AnError
		},
	}

	mockSession := &mockSessionService{
		readSessionFileFunc: func() (services.SessionData, error) {
			return getValidSessionData(), nil
		},
	}

	vmService := services.NewVMServiceWithDeps(logger, true, mockHTTP, mockSession)
	taskID, err := vmService.StartVM("pve1", 100)

	assert.Error(t, err)
	assert.Empty(t, taskID)
}

func TestVMService_GetVMStatus_InvalidJSON(t *testing.T) {
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

	vmService := services.NewVMServiceWithDeps(logger, true, mockHTTP, mockSession)
	status, err := vmService.GetVMStatus("pve1", 100)

	assert.Error(t, err)
	assert.Nil(t, status)
}
