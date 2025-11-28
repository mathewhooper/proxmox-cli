package services

import (
	"crypto/tls"
	"io"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

// HTTPServiceInterface allows mocking of HTTP operations for AuthService.
type HTTPServiceInterface interface {
	Post(url string, payload string, headers map[string]string, cookies []*http.Cookie) (string, error)
	Get(url string, headers map[string]string, cookies []*http.Cookie) (*http.Response, error)
	Put(url string, payload string, headers map[string]string, cookies []*http.Cookie) (string, error)
	Delete(url string, headers map[string]string, cookies []*http.Cookie) (string, error)
}

// HttpService provides HTTP client functionality with optional SSL trust and logging.
type HttpService struct {
	HTTPClient *http.Client
	trust      bool
	logger     *logrus.Logger
}

// URLEncodedHeader is a reusable header map for URL-encoded POST requests.
var URLEncodedHeader = map[string]string{
	"Content-Type": "application/x-www-form-urlencoded; charset=UTF-8",
}

// NewHttpService creates and returns a new HttpService instance.
// logger: the logger to use for HTTP operations.
// trust: if true, disables SSL certificate verification.
func NewHttpService(logger *logrus.Logger, trust bool) *HttpService {
	transport := &http.Transport{}

	if trust {
		//nolint:gosec // G402: InsecureSkipVerify is intentional when user sets --trust flag
		transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // #nosec G402 -- User explicitly requested with --trust flag
		}
	}

	return &HttpService{
		HTTPClient: &http.Client{Transport: transport},
		trust:      trust,
		logger:     logger,
	}
}

// Get sends an HTTP GET request to the specified URL with optional headers and cookies.
// Returns the HTTP response or an error.
func (s *HttpService) Get(url string, headers map[string]string, cookies []*http.Cookie) (*http.Response, error) {
	req, err := s.createRequest("GET", url, nil, headers, cookies)
	if err != nil {
		s.logger.Error("Error creating GET request:", err)
		return nil, err
	}

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		s.logger.Error("Error executing GET request:", err)
		return nil, err
	}

	return resp, nil
}

// Post sends an HTTP POST request to the specified URL with the given payload, headers, and cookies.
// Returns the response body as a string or an error.
func (s *HttpService) Post(url string, payload string, headers map[string]string, cookies []*http.Cookie) (string, error) {
	req, err := s.createRequest("POST", url, strings.NewReader(payload), headers, cookies)
	if err != nil {
		s.logger.Error("Error creating POST request:", err)
		return "", err
	}

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		s.logger.Error("Error executing POST request:", err)
		return "", err
	}
	//nolint:errcheck // Best effort close in defer
	defer func() { _ = resp.Body.Close() }()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Error("Error reading POST response body:", err)
		return "", err
	}

	return string(bodyBytes), nil
}

// Put sends an HTTP PUT request to the specified URL with the given payload, headers, and cookies.
// Returns the response body as a string or an error.
func (s *HttpService) Put(url string, payload string, headers map[string]string, cookies []*http.Cookie) (string, error) {
	req, err := s.createRequest("PUT", url, strings.NewReader(payload), headers, cookies)
	if err != nil {
		s.logger.Error("Error creating PUT request:", err)
		return "", err
	}

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		s.logger.Error("Error executing PUT request:", err)
		return "", err
	}
	//nolint:errcheck // Best effort close in defer
	defer func() { _ = resp.Body.Close() }()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Error("Error reading PUT response body:", err)
		return "", err
	}

	return string(bodyBytes), nil
}

// Delete sends an HTTP DELETE request to the specified URL with optional headers and cookies.
// Returns the response body as a string or an error.
func (s *HttpService) Delete(url string, headers map[string]string, cookies []*http.Cookie) (string, error) {
	req, err := s.createRequest("DELETE", url, nil, headers, cookies)
	if err != nil {
		s.logger.Error("Error creating DELETE request:", err)
		return "", err
	}

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		s.logger.Error("Error executing DELETE request:", err)
		return "", err
	}
	//nolint:errcheck // Best effort close in defer
	defer func() { _ = resp.Body.Close() }()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Error("Error reading DELETE response body:", err)
		return "", err
	}

	return string(bodyBytes), nil
}

// createRequest constructs an HTTP request with the specified method, URL, payload, headers, and cookies.
// Returns the constructed *http.Request or an error.
func (s *HttpService) createRequest(method, url string, payload io.Reader, headers map[string]string, cookies []*http.Cookie) (*http.Request, error) {
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return nil, err
	}

	s.logger.Infof("HTTP %s Request: %s", method, url)

	if headers != nil {
		s.logger.Info("Headers:")

		for key, value := range headers {
			req.Header.Set(key, value)
			s.logger.Infof("  %s: %s", key, value)
		}
	}

	if cookies != nil {
		s.logger.Info("Cookies:")

		for _, cookie := range cookies {
			req.AddCookie(cookie)
			s.logger.Infof("  %s: %s", cookie.Name, cookie.Value)
		}
	}

	return req, nil
}
