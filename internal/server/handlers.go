package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/rxxuzi/sirius/internal/client"
)

func handleInfo(w http.ResponseWriter, r *http.Request) {
	client.SSHMutex.Lock()
	defer client.SSHMutex.Unlock()

	if client.SSHClient == nil {
		http.Error(w, "SSH connection not established", http.StatusServiceUnavailable)
		return
	}

	session, err := client.SSHClient.NewSession()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create SSH session: %v", err), http.StatusInternalServerError)
		return
	}
	defer session.Close()

	output, err := session.CombinedOutput("uname -a && df -h && free -m")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to execute command: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	if _, err := w.Write(output); err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

func handleConnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var config client.SSHConfig
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&config); err != nil {
		log.Printf("Error decoding JSON: %v", err)
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	if config.Port <= 0 || config.Port > 65535 {
		http.Error(w, "Invalid port number. Port must be between 1 and 65535", http.StatusBadRequest)
		return
	}

	if config.Host == "" || config.User == "" {
		http.Error(w, "Host and User are required fields", http.StatusBadRequest)
		return
	}

	if err := client.ConnectSSH(&config); err != nil {
		log.Printf("Failed to connect to SSH server: %v", err)
		http.Error(w, fmt.Sprintf("Failed to connect: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Connected successfully")
}
