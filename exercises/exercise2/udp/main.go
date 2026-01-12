package main

import (
	"fmt"
	"net"
	"time"
)

const (
	DiscoveryPort = "30000"
	ServerPort    = "20000"
	LocalPort     = 20001
	Message       = "Hello from Group 15"
)

func main() {
	conn, err := setupUDPConnection(LocalPort)
	if err != nil {
		fmt.Printf("Failed to setup UDP: %v\n", err)
		return
	}
	defer conn.Close()

	serverIP := discoverServerIP(DiscoveryPort)
	fmt.Printf("Server discovered at: %s\n", serverIP)

	serverAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%s", serverIP, ServerPort))
	if err != nil {
		fmt.Printf("Failed to resolve server address: %v\n", err)
		return
	}

	fmt.Printf("System initialized. Local port: %d, Target: %s\n", LocalPort, serverAddr)

	go receiveLoop(conn)

	sendLoop(conn, serverAddr)
}

func discoverServerIP(port string) string {
	addr, _ := net.ResolveUDPAddr("udp", ":"+port)
	tempConn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Printf("Discovery error: %v\n", err)
		return ""
	}
	defer tempConn.Close()

	buffer := make([]byte, 1024)

	n, remoteAddr, err := tempConn.ReadFromUDP(buffer)
	if err != nil {
		fmt.Printf("Read error during discovery: %v\n", err)
		return ""
	}

	fmt.Printf("Found server! Message: %s\n", string(buffer[:n]))
	return remoteAddr.IP.String()
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
