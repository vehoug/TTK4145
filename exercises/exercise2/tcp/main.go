package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

const (
	UDPDiscoveryPort = "30000"
	TCPDelimitedPort = "33546"
	MyCallbackPort   = "40000"
)

func main() {
	fmt.Println("Searching for server via UDP broadcast...")
	serverIP := discoverServerIP(UDPDiscoveryPort)
	fmt.Printf("Server found at: %s\n", serverIP)

	myIP := getMyIP(serverIP)
	fmt.Printf("My IP identified as: %s\n", myIP)

	go startLocalListener(MyCallbackPort)

	// Connect to the Server
	serverAddr := serverIP + ":" + TCPDelimitedPort
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		fmt.Printf("Failed to connect to server: %v\n", err)
		return
	}
	defer conn.Close()

	go handleConnection(conn, "Primary Connection")

	invite := fmt.Sprintf("Connect to: %s:%s\x00", myIP, MyCallbackPort)
	conn.Write([]byte(invite))
	fmt.Printf("Sent callback invitation: %s\n", invite)

	for {
		conn.Write([]byte("Checking in...\x00"))
		time.Sleep(5 * time.Second)
	}
}

func discoverServerIP(port string) string {
	addr, _ := net.ResolveUDPAddr("udp", ":"+port)
	conn, _ := net.ListenUDP("udp", addr)
	defer conn.Close()

	buf := make([]byte, 1024)
	n, remoteAddr, _ := conn.ReadFromUDP(buf)
	fmt.Printf("Received discovery: %s\n", string(buf[:n]))
	return remoteAddr.IP.String()
}

func getMyIP(serverIP string) string {
	// A trick to find our 'source' IP: Dial the server (UDP is connectionless,
	// so this doesn't actually send a packet, it just checks the OS routing table)
	conn, err := net.Dial("udp", serverIP+":1")
	if err != nil {
		return "127.0.0.1"
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

func startLocalListener(port string) {
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Printf("Listener Error: %v\n", err)
		return
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		fmt.Printf("!!! Server established callback connection from %s !!!\n", conn.RemoteAddr())
		go handleConnection(conn, "Callback Connection")
	}
}

func handleConnection(conn net.Conn, label string) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadBytes('\x00')
		if err != nil {
			fmt.Printf("[%s] Terminated.\n", label)
			return
		}
		fmt.Printf("[%s] Received: %s\n", label, strings.Trim(string(msg), "\x00"))
	}
}
