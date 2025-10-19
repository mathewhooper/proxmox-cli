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

func TestStorageService_ListStorage_Success(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	mockHTTP := &mockHttpService{
		getFunc: func(url string, headers map[string]string, cookies []*http.Cookie) (*http.Response, error) {
			body := `{"data": [
				{
					"storage": "local",
					"type": "dir",
					"content": "vztmpl,iso,backup",
					"shared": 0,
					"active": 1,
					"enabled": 1
				},
				{
					"storage": "local-lvm",
					"type": "lvmthin",
					"content": "images,rootdir",
					"shared": 0,
					"active": 1,
					"enabled": 1
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

	storageService := services.NewStorageServiceWithDeps(logger, true, mockHTTP, mockSession)
	storages, err := storageService.ListStorage()

	assert.NoError(t, err)
	assert.Len(t, storages, 2)
	assert.Equal(t, "local", storages[0].Storage)
	assert.Equal(t, "dir", storages[0].Type)
	assert.Equal(t, 0, storages[0].Shared)
	assert.Equal(t, 1, storages[0].Active)
	assert.Equal(t, "local-lvm", storages[1].Storage)
	assert.Equal(t, "lvmthin", storages[1].Type)
}

func TestStorageService_ListStorage_SessionError(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	mockHTTP := &mockHttpService{}
	mockSession := &mockSessionService{
		readSessionFileFunc: func() (services.SessionData, error) {
			return services.SessionData{}, assert.AnError
		},
	}

	storageService := services.NewStorageServiceWithDeps(logger, true, mockHTTP, mockSession)
	storages, err := storageService.ListStorage()

	assert.Error(t, err)
	assert.Nil(t, storages)
}

func TestStorageService_ListStorage_HttpError(t *testing.T) {
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

	storageService := services.NewStorageServiceWithDeps(logger, true, mockHTTP, mockSession)
	storages, err := storageService.ListStorage()

	assert.Error(t, err)
	assert.Nil(t, storages)
}

func TestStorageService_ListStorage_InvalidJSON(t *testing.T) {
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

	storageService := services.NewStorageServiceWithDeps(logger, true, mockHTTP, mockSession)
	storages, err := storageService.ListStorage()

	assert.Error(t, err)
	assert.Nil(t, storages)
}

func TestStorageService_ListStorageContent_Success(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	mockHTTP := &mockHttpService{
		getFunc: func(url string, headers map[string]string, cookies []*http.Cookie) (*http.Response, error) {
			body := `{"data": [
				{
					"volid": "local:iso/debian-11.0.0-amd64-netinst.iso",
					"format": "iso",
					"size": 377487360,
					"ctime": 1625097600
				},
				{
					"volid": "local:vztmpl/debian-11-standard_11.0-1_amd64.tar.gz",
					"format": "tgz",
					"size": 226492416,
					"ctime": 1625097700
				},
				{
					"volid": "local-lvm:vm-100-disk-0",
					"format": "raw",
					"size": 34359738368,
					"used": 8589934592,
					"vmid": 100,
					"ctime": 1625097800
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

	storageService := services.NewStorageServiceWithDeps(logger, true, mockHTTP, mockSession)
	contents, err := storageService.ListStorageContent("pve1", "local")

	assert.NoError(t, err)
	assert.Len(t, contents, 3)
	assert.Equal(t, "local:iso/debian-11.0.0-amd64-netinst.iso", contents[0].VolID)
	assert.Equal(t, "iso", contents[0].Format)
	assert.Equal(t, int64(377487360), contents[0].Size)
	assert.Equal(t, "local-lvm:vm-100-disk-0", contents[2].VolID)
	assert.Equal(t, 100, contents[2].VMID)
	assert.Equal(t, int64(8589934592), contents[2].Used)
}

func TestStorageService_ListStorageContent_SessionError(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	mockHTTP := &mockHttpService{}
	mockSession := &mockSessionService{
		readSessionFileFunc: func() (services.SessionData, error) {
			return services.SessionData{}, assert.AnError
		},
	}

	storageService := services.NewStorageServiceWithDeps(logger, true, mockHTTP, mockSession)
	contents, err := storageService.ListStorageContent("pve1", "local")

	assert.Error(t, err)
	assert.Nil(t, contents)
}

func TestStorageService_ListStorageContent_HttpError(t *testing.T) {
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

	storageService := services.NewStorageServiceWithDeps(logger, true, mockHTTP, mockSession)
	contents, err := storageService.ListStorageContent("pve1", "local")

	assert.Error(t, err)
	assert.Nil(t, contents)
}

func TestStorageService_ListStorageContent_InvalidJSON(t *testing.T) {
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

	storageService := services.NewStorageServiceWithDeps(logger, true, mockHTTP, mockSession)
	contents, err := storageService.ListStorageContent("pve1", "local")

	assert.Error(t, err)
	assert.Nil(t, contents)
}
