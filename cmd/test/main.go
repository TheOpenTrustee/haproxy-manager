package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

const haproxyAPIURL = "http://localhost:5555/v3" // URL to HAProxy Data Plane API (adjust as needed)
const haproxyUser = "admin"                      // Basic Auth username
const haproxyPass = "mypassword"                 // Basic Auth password

type Frontend struct {
	Name           string      `json:"name"`
	Mode           string      `json:"mode"`
	Binds          interface{} `json:"binds"`
	DefaultBackend string      `json:"default_backend"`
}

func main() {
	// List existing frontends
	fmt.Println("Existing frontends:")
	if err := listFrontends(); err != nil {
		log.Fatalf("Error listing frontends: %v", err)
	}

	// Wait for 5 seconds
	fmt.Println("Waiting for 5 seconds...")
	time.Sleep(5 * time.Second)

	// Add a new frontend
	frontendName := "new-frontend"
	bindAddress := "0.0.0.0:8081"
	backendName := "existing-backend" // Replace with a valid backend name
	fmt.Printf("Creating a new frontend: %s\n", frontendName)
	if err := createFrontend(frontendName, bindAddress, backendName); err != nil {
		log.Fatalf("Error creating frontend: %v", err)
	}

	// List frontends again
	fmt.Println("Frontends after adding a new one:")
	if err := listFrontends(); err != nil {
		log.Fatalf("Error listing frontends: %v", err)
	}
}

// listFrontends lists all existing frontends using the HAProxy Data Plane API
func listFrontends() error {
	url := fmt.Sprintf("%s/services/haproxy/configuration/frontends", haproxyAPIURL)
	resp, err := makeRequest("GET", url, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var frontends []Frontend
	if err := json.NewDecoder(resp.Body).Decode(&frontends); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	for _, frontend := range frontends {
		fmt.Printf("Frontend Name: %s, Mode: %s, Bind: %+v, Default Backend: %s\n", frontend.Name, frontend.Mode, frontend.Binds, frontend.DefaultBackend)
	}

	return nil
}

// createFrontend creates a new frontend dynamically using the HAProxy Data Plane API
func createFrontend(name string, bindAddress interface{}, backendName string) error {
	frontend := Frontend{
		Name:           name,
		Mode:           "http", // Mode can be "http" or "tcp"
		Binds:          bindAddress,
		DefaultBackend: backendName,
	}

	url := fmt.Sprintf("%s/api/frontend", haproxyAPIURL)
	data, err := json.Marshal(frontend)
	if err != nil {
		return fmt.Errorf("failed to marshal frontend data: %w", err)
	}

	resp, err := makeRequest("POST", url, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to create frontend, status: %d", resp.StatusCode)
	}

	fmt.Printf("Frontend %s created successfully\n", name)
	return nil
}

// makeRequest performs an HTTP request with the provided method, URL, and data
func makeRequest(method, url string, data []byte) (*http.Response, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+encodeBasicAuth(haproxyUser, haproxyPass)) // Add the Basic Auth header

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// encodeBasicAuth encodes the username and password for Basic Auth in the format "username:password" -> Base64
func encodeBasicAuth(username, password string) string {
	auth := fmt.Sprintf("%s:%s", username, password)
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
