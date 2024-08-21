package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/rxxuzi/sirius/internal/client"
)

func handleStaticInfo(w http.ResponseWriter, r *http.Request) {
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

	output, err := session.CombinedOutput("uname -a && echo '\n--- Disk Usage ---' && df -h")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to execute command: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	if _, err := w.Write(output); err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

func handleLiveInfo(w http.ResponseWriter, r *http.Request) {
	client.SSHMutex.Lock()
	defer client.SSHMutex.Unlock()

	if client.SSHClient == nil {
		http.Error(w, "SSH connection not established", http.StatusServiceUnavailable)
		return
	}

	info := make(map[string]string)

	// CPU使用率を取得
	cpuCmd := "top -bn1 | grep \"Cpu(s)\" | awk '{print $2 + $4}'"
	cpuUsage, err := executeSSHCommand(cpuCmd)
	if err != nil {
		log.Printf("Error getting CPU usage: %v", err)
		info["CPU"] = "Error"
	} else {
		info["CPU"] = fmt.Sprintf("%.2f%%", parseFloat(cpuUsage))
	}

	// メモリ使用率を取得
	memCmd := "free | grep Mem | awk '{print $3/$2 * 100.0}'"
	memUsage, err := executeSSHCommand(memCmd)
	if err != nil {
		log.Printf("Error getting memory usage: %v", err)
		info["Memory"] = "Error"
	} else {
		info["Memory"] = fmt.Sprintf("%.2f%%", parseFloat(memUsage))
	}

	// GPU情報を取得（NVIDIA GPUがある場合）
	gpuCmd := "if command -v nvidia-smi &> /dev/null; then nvidia-smi --query-gpu=utilization.gpu --format=csv,noheader,nounits; else echo 'N/A'; fi"
	gpuUsage, err := executeSSHCommand(gpuCmd)
	if err != nil {
		log.Printf("Error getting GPU usage: %v", err)
		info["GPU"] = "Error"
	} else {
		gpuUsage = strings.TrimSpace(gpuUsage)
		if gpuUsage != "N/A" {
			info["GPU"] = fmt.Sprintf("%s%%", gpuUsage)
		} else {
			info["GPU"] = gpuUsage
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

func executeSSHCommand(cmd string) (string, error) {
	session, err := client.SSHClient.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create SSH session: %v", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return "", fmt.Errorf("failed to execute command: %v", err)
	}

	return strings.TrimSpace(string(output)), nil
}

func parseFloat(s string) float64 {
	f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		log.Printf("Error parsing float: %v", err)
		return 0
	}
	return f
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
