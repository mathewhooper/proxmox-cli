package tests

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"proxmox-cli/services"

	"github.com/sirupsen/logrus"
)

func TestHttpService_Get_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hello world"))
	}))
	defer ts.Close()

	logger := logrus.New()
	httpService := services.NewHttpService(logger, false)
	resp, err := httpService.Get(ts.URL, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "hello world" {
		t.Errorf("expected body 'hello world', got '%s'", string(body))
	}
}

func TestHttpService_Get_Error(t *testing.T) {
	logger := logrus.New()
	httpService := services.NewHttpService(logger, false)
	_, err := httpService.Get("http://invalid.invalid", nil, nil)
	if err == nil {
		t.Error("expected error for invalid host")
	}
}

func TestHttpService_Post_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != "foo=bar" {
			t.Errorf("expected body 'foo=bar', got '%s'", string(body))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer ts.Close()

	logger := logrus.New()
	httpService := services.NewHttpService(logger, false)
	resp, err := httpService.Post(ts.URL, "foo=bar", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "ok" {
		t.Errorf("expected response 'ok', got '%s'", resp)
	}
}

func TestHttpService_Post_Error(t *testing.T) {
	logger := logrus.New()
	httpService := services.NewHttpService(logger, false)
	_, err := httpService.Post("http://invalid.invalid", "foo=bar", nil, nil)
	if err == nil {
		t.Error("expected error for invalid host")
	}
}

func TestHttpService_HeadersAndCookies(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Test-Header") != "test-value" {
			t.Errorf("expected header X-Test-Header to be 'test-value'")
		}
		cookie, err := r.Cookie("testcookie")
		if err != nil || cookie.Value != "cookieval" {
			t.Errorf("expected cookie 'testcookie' to be 'cookieval'")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer ts.Close()

	logger := logrus.New()
	httpService := services.NewHttpService(logger, false)
	headers := map[string]string{"X-Test-Header": "test-value"}
	cookies := []*http.Cookie{{Name: "testcookie", Value: "cookieval"}}
	resp, err := httpService.Post(ts.URL, "", headers, cookies)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "ok" {
		t.Errorf("expected response 'ok', got '%s'", resp)
	}
}

func TestHttpService_Put_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != "test=update" {
			t.Errorf("expected body 'test=update', got '%s'", string(body))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("updated"))
	}))
	defer ts.Close()

	logger := logrus.New()
	httpService := services.NewHttpService(logger, false)
	resp, err := httpService.Put(ts.URL, "test=update", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "updated" {
		t.Errorf("expected response 'updated', got '%s'", resp)
	}
}

func TestHttpService_Put_Error(t *testing.T) {
	logger := logrus.New()
	httpService := services.NewHttpService(logger, false)
	_, err := httpService.Put("http://invalid.invalid", "test=data", nil, nil)
	if err == nil {
		t.Error("expected error for invalid host")
	}
}

func TestHttpService_Put_WithHeadersAndCookies(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.Header.Get("X-Test-Header") != "test-value" {
			t.Errorf("expected header X-Test-Header to be 'test-value'")
		}
		cookie, err := r.Cookie("testcookie")
		if err != nil || cookie.Value != "cookieval" {
			t.Errorf("expected cookie 'testcookie' to be 'cookieval'")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("updated"))
	}))
	defer ts.Close()

	logger := logrus.New()
	httpService := services.NewHttpService(logger, false)
	headers := map[string]string{"X-Test-Header": "test-value"}
	cookies := []*http.Cookie{{Name: "testcookie", Value: "cookieval"}}
	resp, err := httpService.Put(ts.URL, "test=update", headers, cookies)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "updated" {
		t.Errorf("expected response 'updated', got '%s'", resp)
	}
}

func TestHttpService_Delete_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("deleted"))
	}))
	defer ts.Close()

	logger := logrus.New()
	httpService := services.NewHttpService(logger, false)
	resp, err := httpService.Delete(ts.URL, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "deleted" {
		t.Errorf("expected response 'deleted', got '%s'", resp)
	}
}

func TestHttpService_Delete_Error(t *testing.T) {
	logger := logrus.New()
	httpService := services.NewHttpService(logger, false)
	_, err := httpService.Delete("http://invalid.invalid", nil, nil)
	if err == nil {
		t.Error("expected error for invalid host")
	}
}

func TestHttpService_Delete_WithHeadersAndCookies(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.Header.Get("CSRFPreventionToken") != "csrf-token" {
			t.Errorf("expected header CSRFPreventionToken to be 'csrf-token'")
		}
		cookie, err := r.Cookie("authcookie")
		if err != nil || cookie.Value != "authvalue" {
			t.Errorf("expected cookie 'authcookie' to be 'authvalue'")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("deleted"))
	}))
	defer ts.Close()

	logger := logrus.New()
	httpService := services.NewHttpService(logger, false)
	headers := map[string]string{"CSRFPreventionToken": "csrf-token"}
	cookies := []*http.Cookie{{Name: "authcookie", Value: "authvalue"}}
	resp, err := httpService.Delete(ts.URL, headers, cookies)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "deleted" {
		t.Errorf("expected response 'deleted', got '%s'", resp)
	}
}
