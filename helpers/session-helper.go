package helpers

import (
	"encoding/json"
	"os"
	"os/user"
	"path/filepath"
)

// WriteSessionFile writes session data to a file in the user's home directory.
//
// Parameters:
//   - data: A map containing session data to be written to the file.
//
// Returns:
//   - An error if the operation fails, or nil if the operation is successful.
//
// Behavior:
//   - The function determines the current user's home directory.
//   - It creates a `.proxmox` directory in the home directory if it does not exist.
//   - It writes the session data to a file named `session` in the `.proxmox` directory.
//   - The session data is encoded in JSON format.
//
// Example:
//   err := WriteSessionFile(map[string]interface{}{"key": "value"})
//   if err != nil {
//       log.Fatal(err)
//   }
func WriteSessionFile(data map[string]interface{}) error {
	usr, err := user.Current()
	if err != nil {
		return err
	}
	dirPath := filepath.Join(usr.HomeDir, ".proxmox")
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return err
		}
	}
	filePath := filepath.Join(dirPath, "session")
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(data)
}

// ReadSessionFile reads session data from a file in the user's home directory.
//
// Returns:
//   - A map containing the session data if the operation is successful.
//   - An error if the operation fails.
//
// Behavior:
//   - The function determines the current user's home directory.
//   - It attempts to open a file named `session` in the `.proxmox` directory within the home directory.
//   - If the file exists, it decodes the JSON content into a map and returns it.
//   - If the file does not exist or cannot be read, an error is returned.
//
// Example:
//   sessionData, err := ReadSessionFile()
//   if err != nil {
//       log.Fatal(err)
//   }
//   fmt.Println("Session Data:", sessionData)
func ReadSessionFile() (map[string]interface{}, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}
	sessionFilePath := filepath.Join(usr.HomeDir, ".proxmox", "session")
	file, err := os.Open(sessionFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var sessionData map[string]interface{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&sessionData); err != nil {
		return nil, err
	}
	return sessionData, nil
}

// UpdateSessionField updates a specific field in the session file with a new value.
//
// Parameters:
//   - field: The name of the field to update.
//   - value: The new value to set for the specified field.
//
// Returns:
//   - An error if the operation fails, or nil if the operation is successful.
func UpdateSessionField(field string, value interface{}) error {
	// Read the existing session data
	sessionData, err := ReadSessionFile()
	if err != nil {
		return err
	}

	sessionData[field] = value
	return WriteSessionFile(sessionData)
}
