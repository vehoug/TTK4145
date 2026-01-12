package main

import (
	"fmt"
	"net"
	"time"
)

const (
	ServerIP   = "10.0.0.17"
	ServerPort = "20000"
	LocalPort  = 20001
	Message    = "Hello from Group 15"
)

func main() {
	conn, err := setupUDPConnection(LocalPort)
	if err != nil {
		fmt.Printf("Failed to setup UDP: %v\n", err)
		return
	}
	defer conn.Close()

	serverAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%s", ServerIP, ServerPort))
	if err != nil {
		fmt.Printf("Failed to resolve server address: %v\n", err)
		return
	}

	fmt.Printf("System initialized. Local port: %d, Target: %s\n", LocalPort, serverAddr)

	go receiveLoop(conn)

	sendLoop(conn, serverAddr)
}

func setupUDPConnection(port int) (*net.UDPConn, error) {
	localAddr := &net.UDPAddr{
		IP:   net.ParseIP("0.0.0.0"),
		Port: port,
	}
	return net.ListenUDP("udp", localAddr)
}

func receiveLoop(conn *net.UDPConn) {
	buffer := make([]byte, 1024)
	for {
		n, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Printf("Read error: %v\n", err)
			continue
		}
		fmt.Printf("[%s] Received: %s\n", remoteAddr, string(buffer[:n]))
	}
}

func sendLoop(conn *net.UDPConn, target *net.UDPAddr) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	payload := []byte(Message)

	for range ticker.C {
		_, err := conn.WriteToUDP(payload, target)
		if err != nil {
			fmt.Printf("Write error: %v\n", err)
			continue
		}
		fmt.Println("Message sent...")
	}
}
