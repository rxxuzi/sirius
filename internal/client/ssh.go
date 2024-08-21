package client

import (
	"fmt"
	"strings"
	"sync"

	"golang.org/x/crypto/ssh"
)

var (
	SSHClient *ssh.Client
	SSHMutex  sync.Mutex
)

func ConnectSSH(config *SSHConfig) error {
	SSHMutex.Lock()
	defer SSHMutex.Unlock()

	if SSHClient != nil {
		SSHClient.Close()
	}

	sshConfig := &ssh.ClientConfig{
		User: config.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(config.Pass),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return fmt.Errorf("SSH connection failed: %v", err)
	}

	SSHClient = client
	return nil
}

func IsConnected() bool {
	SSHMutex.Lock()
	defer SSHMutex.Unlock()
	return SSHClient != nil
}

func Exec(cmd string) (string, error) {
	session, err := SSHClient.NewSession()
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
