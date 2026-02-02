package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"time"
)

const (
	port             = ":20000"
	broadcastAddress = "255.255.255.255" + port
	timeout          = 2 * time.Second
)

func main() {
	localAddr, _ := net.ResolveUDPAddr("udp", port)

	conn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		fmt.Printf("Error binding to port %s: %v\n", port, err)
	}

	fmt.Println("-- Started as BACKUP. Waiting for primary... --")

	var count = 0
	buffer := make([]byte, 1024)

	for {
		conn.SetReadDeadline(time.Now().Add(timeout))
		n, _, err := conn.ReadFromUDP(buffer)

		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				fmt.Println("Timeout! Primary is dead. Taking over...")
				break
			}
			fmt.Println("Read error: ", err)
		} else {
			receivedData := string(buffer[:n])
			receivedCount, _ := strconv.Atoi(receivedData)
			count = receivedCount
		}
	}

	conn.Close()

	spawnBackup()

	fmt.Println("-- I am now the PRIMARY --")

	dstAddr, _ := net.ResolveUDPAddr("udp", broadcastAddress)

	sender, _ := net.DialUDP("udp", nil, dstAddr)
	defer sender.Close()

	for {
		count++
		fmt.Printf("%d\n", count)

		msg := strconv.Itoa(count)
		_, err := sender.Write([]byte(msg))
		if err != nil {
			fmt.Printf("Error broadcasting: %v\n", err)
		}

		time.Sleep(1 * time.Second)
	}
}

func spawnBackup() {
	executable, err := os.Executable()
	if err != nil {
		fmt.Println("Error getting executable path:", err)
		return
	}

	cmd := exec.Command("gnome-terminal", "--", executable)

	if err := cmd.Start(); err != nil {
		fmt.Printf("Failed to spawn backup: %v\n", err)
	} else {
		fmt.Println("Backup process spawned.")
	}
}
