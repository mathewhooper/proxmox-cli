package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/sirupsen/logrus"
)

// AuthService handles authentication and session validation.
type AuthService struct {
	Logger         *logrus.Logger
	Trust          bool
	HttpService    HttpServiceInterface
	SessionService SessionServiceInterface
}

// NewAuthService creates an AuthService with real HTTP and Session services (for production use).
func NewAuthService(logger *logrus.Logger, trust bool) *AuthService {
	return &AuthService{
		Logger:         logger,
		Trust:          trust,
		HttpService:    NewHttpService(logger, trust),
		SessionService: mustNewSessionService(logger),
	}
}

// NewAuthServiceWithDeps allows injecting mocks for testing.
func NewAuthServiceWithDeps(logger *logrus.Logger, trust bool, httpService HttpServiceInterface, sessionService SessionServiceInterface) *AuthService {
	return &AuthService{
		Logger:         logger,
		Trust:          trust,
		HttpService:    httpService,
		SessionService: sessionService,
	}
}

// mustNewSessionService is a helper for production constructor.
func mustNewSessionService(logger *logrus.Logger) SessionServiceInterface {
	ss, err := NewSessionService(logger)
	if err != nil {
		panic(err)
	}
	return ss
}

func (a *AuthService) LoginToProxmox(server string, port int, httpScheme, username, password string) error {
	uri := fmt.Sprintf("%s://%s:%d/api2/json/access/ticket", httpScheme, server, port)
	payload := fmt.Sprintf("username=%s&password=%s&realm=pam&new-format=1", username, password)
	headers := UrlEncodedHeader

	body, err := a.HttpService.Post(uri, payload, headers, nil)
	if err != nil {
		a.Logger.Error("Error logging in: ", err)
		return err
	}

	a.Logger.Info("Response: ", body)

	var resp SessionDataResponse
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		a.Logger.Error("Error parsing response JSON: ", err)
		return err
	}

	sessionData := SessionData{
		Server:     server,
		Port:       port,
		HttpScheme: httpScheme,
		Response:   resp,
	}

	if err := a.SessionService.WriteSessionFile(sessionData); err != nil {
		a.Logger.Error("Error writing session data to file: ", err)
		return err
	}
	a.Logger.Info("Authenticated!")
	return nil
}

func (a *AuthService) ValidateSession() bool {
	sessionData, err := a.SessionService.ReadSessionFile()
	if err != nil {
		a.Logger.Error("Error reading session file: ", err)
		return false
	}

	uri := fmt.Sprintf("%s://%s:%d/api2/json/access/ticket", sessionData.HttpScheme, sessionData.Server, sessionData.Port)

	payload := fmt.Sprintf("username=%s&password=%s", sessionData.Response.Data.Username, url.QueryEscape(sessionData.Response.Data.Ticket))
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

	body, err := a.HttpService.Post(uri, payload, headers, cookies)
	if err != nil {
		a.Logger.Error("Error validating session: ", err)
		return false
	}

	if err := a.SessionService.UpdateSessionField("response", body); err != nil {
		a.Logger.Error("Error updating session file: ", err)
		return false
	}

	return true
}
