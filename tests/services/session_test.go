package tests

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"proxmox-cli/services"

	"github.com/sirupsen/logrus"
)

// Remove mockUser struct and related code

// Remove monkeyPatchUserCurrent and all its usages

func TestSessionService_WriteAndReadSessionFile(t *testing.T) {
	dir, err := os.MkdirTemp("", "proxmox-test")
	if err != nil {
		t.Fatal(err)
	}
	os.Setenv("HOME", dir)
	defer os.RemoveAll(dir)
	logger := logrus.New()
	ss, err := services.NewSessionService(logger)
	if err != nil {
		t.Fatalf("NewSessionService failed: %v", err)
	}
	sd := services.SessionData{
		Server:     "localhost",
		Port:       8006,
		HttpScheme: "https",
		Response: services.SessionDataResponse{
			Data: struct {
				Username            string `json:"username"`
				Ticket              string `json:"ticket"`
				CSRFPreventionToken string `json:"CSRFPreventionToken"`
			}{
				Username:            "user",
				Ticket:              "ticket",
				CSRFPreventionToken: "csrf",
			},
		},
	}
	err = ss.WriteSessionFile(sd)
	if err != nil {
		t.Fatalf("WriteSessionFile failed: %v", err)
	}
	read, err := ss.ReadSessionFile()
	if err != nil {
		t.Fatalf("ReadSessionFile failed: %v", err)
	}
	if !reflect.DeepEqual(read.Server, sd.Server) || read.Port != sd.Port || read.HttpScheme != sd.HttpScheme {
		t.Errorf("ReadSessionFile returned wrong data: %+v", read)
	}
}

func TestSessionService_ReadSessionFile_Invalid(t *testing.T) {
	dir, err := os.MkdirTemp("", "proxmox-test")
	if err != nil {
		t.Fatal(err)
	}
	os.Setenv("HOME", dir)
	defer os.RemoveAll(dir)
	logger := logrus.New()
	ss, err := services.NewSessionService(logger)
	if err != nil {
		t.Fatalf("NewSessionService failed: %v", err)
	}
	filePath := filepath.Join(dir, ".proxmox", "session")
	os.MkdirAll(filepath.Dir(filePath), 0755)
	err = os.WriteFile(filePath, []byte(`{"foo": "bar"}`), 0644)
	if err != nil {
		t.Fatalf("failed to write invalid session file: %v", err)
	}
	read, err := ss.ReadSessionFile()
	content, _ := os.ReadFile(filePath)
	t.Logf("Session file content: %s", content)
	t.Logf("Read session: %+v, err: %v", read, err)
	if err == nil {
		t.Error("expected error for invalid session data")
	}
}

func TestSessionService_ReadSessionFile_Missing(t *testing.T) {
	dir, err := os.MkdirTemp("", "proxmox-test")
	if err != nil {
		t.Fatal(err)
	}
	os.Setenv("HOME", dir)
	defer os.RemoveAll(dir)
	logger := logrus.New()
	ss, err := services.NewSessionService(logger)
	if err != nil {
		t.Fatalf("NewSessionService failed: %v", err)
	}
	// Remove the session file if it exists
	filePath := filepath.Join(dir, ".proxmox", "session")
	os.Remove(filePath)
	_, err = ss.ReadSessionFile()
	if err == nil {
		t.Error("expected error for missing session file")
	}
}

func TestSessionService_UpdateSessionField(t *testing.T) {
	dir, err := os.MkdirTemp("", "proxmox-test")
	if err != nil {
		t.Fatal(err)
	}
	os.Setenv("HOME", dir)
	defer os.RemoveAll(dir)
	logger := logrus.New()
	ss, err := services.NewSessionService(logger)
	if err != nil {
		t.Fatalf("NewSessionService failed: %v", err)
	}
	sd := services.SessionData{
		Server:     "localhost",
		Port:       8006,
		HttpScheme: "https",
		Response: services.SessionDataResponse{
			Data: struct {
				Username            string `json:"username"`
				Ticket              string `json:"ticket"`
				CSRFPreventionToken string `json:"CSRFPreventionToken"`
			}{
				Username:            "user",
				Ticket:              "ticket",
				CSRFPreventionToken: "csrf",
			},
		},
	}
	err = ss.WriteSessionFile(sd)
	if err != nil {
		t.Fatalf("WriteSessionFile failed: %v", err)
	}
	err = ss.UpdateSessionField("server", "newhost")
	if err != nil {
		t.Errorf("UpdateSessionField failed: %v", err)
	}
	read, err := ss.ReadSessionFile()
	if err != nil {
		t.Fatalf("ReadSessionFile failed: %v", err)
	}
	if read.Server != "newhost" {
		t.Errorf("expected server to be 'newhost', got '%s'", read.Server)
	}
}
