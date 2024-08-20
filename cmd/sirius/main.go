package main

import (
	"context"
	"fmt"
	"github.com/rxxuzi/sirius/internal/client"
	"github.com/rxxuzi/sirius/internal/server"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	var configFile string
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}

	if configFile != "" {
		config, err := client.LoadSSHConfig(configFile)
		if err != nil {
			fmt.Printf("Failed to load config: %v\n", err)
		} else {
			err = client.ConnectSSH(config)
			if err != nil {
				fmt.Printf("Failed to connect to SSH: %v\n", err)
			} else {
				fmt.Printf("Successfully connected to SSH server: %s\n", config.Host)
			}
		}
	}

	srv := server.NewServer()

	// サーバーの起動情報を表示
	listener, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		fmt.Printf("Failed to listen: %v\n", err)
		os.Exit(1)
	}

	port := listener.Addr().(*net.TCPAddr).Port

	fmt.Println("Sirius is running on:")
	fmt.Printf("  http://localhost:%d\n", port)
	fmt.Println("Press Ctrl+C to stop the server")

	go func() {
		if err := srv.Serve(listener); err != nil {
			fmt.Printf("Server error: %v\n", err)
		}
	}()

	// シグナルハンドリングの設定
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop

	fmt.Println("\nShutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		fmt.Printf("Server forced to shutdown: %v\n", err)
		os.Exit(1)
	}

	if client.SSHClient != nil {
		client.SSHClient.Close()
		fmt.Println("SSH connection closed")
	}

	fmt.Println("Server stopped gracefully")
}
