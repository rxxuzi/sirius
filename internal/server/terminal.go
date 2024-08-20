package server

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/rxxuzi/sirius/internal/client"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"net/http"
	"sync"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // test.
	},
}

func handleTerminal(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		http.Error(w, "WebSocket upgrade failed", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	session, err := createSSHSession()
	if err != nil {
		sendErrorMessage(conn, fmt.Sprintf("Failed to create SSH session: %v", err))
		return
	}
	defer session.Close()

	stdin, stdout, stderr, err := getSessionPipes(session)
	if err != nil {
		sendErrorMessage(conn, fmt.Sprintf("Failed to get session pipes: %v", err))
		return
	}

	if err := session.Shell(); err != nil {
		sendErrorMessage(conn, fmt.Sprintf("Failed to start shell: %v", err))
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go handleWebSocketToSSH(&wg, conn, stdin)
	go handleSSHToWebSocket(&wg, conn, io.MultiReader(stdout, stderr))

	wg.Wait()
}

func createSSHSession() (*ssh.Session, error) {
	client.SSHMutex.Lock()
	defer client.SSHMutex.Unlock()

	if client.SSHClient == nil {
		return nil, fmt.Errorf("SSH connection not established")
	}

	session, err := client.SSHClient.NewSession()
	if err != nil {
		return nil, err
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	if err := session.RequestPty("xterm", 40, 80, modes); err != nil {
		session.Close()
		return nil, err
	}

	return session, nil
}

func getSessionPipes(session *ssh.Session) (io.WriteCloser, io.Reader, io.Reader, error) {
	stdin, err := session.StdinPipe()
	if err != nil {
		return nil, nil, nil, err
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		return nil, nil, nil, err
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		return nil, nil, nil, err
	}

	return stdin, stdout, stderr, nil
}

func handleWebSocketToSSH(wg *sync.WaitGroup, conn *websocket.Conn, stdin io.WriteCloser) {
	defer wg.Done()
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading WebSocket message: %v", err)
			return
		}
		if _, err := stdin.Write(msg); err != nil {
			log.Printf("Error writing to SSH stdin: %v", err)
			return
		}
	}
}

func handleSSHToWebSocket(wg *sync.WaitGroup, conn *websocket.Conn, reader io.Reader) {
	defer wg.Done()
	buf := make([]byte, 32*1024)
	for {
		n, err := reader.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Printf("Error reading from SSH: %v", err)
			}
			return
		}
		if n > 0 {
			err = conn.WriteMessage(websocket.TextMessage, buf[:n])
			if err != nil {
				log.Printf("Error writing to WebSocket: %v", err)
				return
			}
		}
	}
}

func sendErrorMessage(conn *websocket.Conn, message string) {
	log.Print(message)
	conn.WriteMessage(websocket.TextMessage, []byte(message))
}
