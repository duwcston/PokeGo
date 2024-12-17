package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

const (
	HOST_C = "localhost"
	PORT_C = "8080"
	// TYPE = "udp"
)

func main() {
	serverAddr := HOST_C + ":" + PORT_C

	// Resolve server address
	s, err := net.ResolveUDPAddr("udp4", serverAddr)
	if err != nil {
		fmt.Println("Error resolving address: ", err)
		return
	}

	// Dial UDP connection
	conn, err := net.DialUDP("udp4", nil, s)
	if err != nil {
		fmt.Println("Error connecting to server: ", err)
		return
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your name: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	// Register the username with the server
	_, err = conn.Write([]byte("LOGIN " + username))
	if err != nil {
		fmt.Println("Error sending username to server: ", err)
		return
	}

	fmt.Printf("Connected to UDP server at %s\n", conn.RemoteAddr().String())

	// Start a goroutine to listen for incoming messages from the server
	go func() {
		for {
			buffer := make([]byte, 1024)
			n, _, err := conn.ReadFromUDP(buffer)
			if err != nil {
				fmt.Println("Error reading from server: ", err)
				return
			}
			fmt.Printf("\nMessage: %s\n>> ", string(buffer[:n]))
		}
	}()

	// Handle client input
	for {
		fmt.Print(">> ")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		IsSended := 0
		// Check for exit command
		if text == "STOP" {
			IsSended, err = conn.Write([]byte("LOGOUT" + username))
			if err != nil {
				fmt.Println("Error sending message: ", err)
			}
			fmt.Println("Exiting UDP client!")
			return
		}

		// Handle private and broadcast messaging
		if strings.HasPrefix(text, "@") {
			IsSended, err = conn.Write([]byte("BATTLE " + text + " want to battle"))
			if err != nil {
				fmt.Println("Error sending message:", err)
				continue
			}
		}
		if IsSended == 0 {
			_, err = conn.Write([]byte(text))
			if err != nil {
				fmt.Println("Error sending message:", err)
				continue
			}
		}

	}
}
