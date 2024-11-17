package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net"
	"time"
)

func main() {
	mode := flag.String("mode", "rendezvous", "Mode: rendezvous, udp-client-nat")
	address := flag.String("address", "localhost:8080", "Server address (e.g., localhost:8080)")
	flag.Parse()

	switch *mode {
	case "udp-client-nat":
		runUDPClientNAT(*address)
	default:
		fmt.Println("Unknown mode:", *mode)
	}
}

func runUDPClientNAT(address string) {
	serverAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		fmt.Printf("Error resolving server address: %v\n", err)
		return
	}
	fmt.Printf("Resolved server address: %v\n", serverAddr)

	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		fmt.Printf("Error connecting to rendezvous server: %v\n", err)
		return
	}
	defer conn.Close()
	fmt.Println("Connected to rendezvous server.")

	fmt.Println("Registering with rendezvous server...")
	_, err = conn.Write([]byte("REGISTER"))
	if err != nil {
		fmt.Printf("Error registering with server: %v\n", err)
		return
	}

	buf := make([]byte, 1024)
	n, _, err := conn.ReadFromUDP(buf)
	if err != nil {
		fmt.Printf("Error receiving peer address: %v\n", err)
		return
	}
	peerAddrStr := string(buf[:n])
	if peerAddrStr == "" {
		fmt.Println("Error: received empty peer address")
		return
	}
	fmt.Printf("Received peer address: %s\n", peerAddrStr)

	peerAddr, err := net.ResolveUDPAddr("udp", peerAddrStr)
	if err != nil {
		fmt.Printf("Error resolving peer address: %v\n", err)
		return
	}
	fmt.Printf("Resolved peer address: %v\n", peerAddr)
	clientPort := 55001 + rand.Intn(10000)

	localConn, err := net.ListenUDP("udp", nil)
	if err != nil {
		fmt.Printf("Error creating local UDP socket: %v\n", err)
		return
	}
	defer localConn.Close()
	fmt.Printf("Local UDP socket created on port %d.\n", clientPort)

	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-done:
				fmt.Println("Stopping receive goroutine.")
				return
			default:
				buf := make([]byte, 1024)
				n, addr, err := localConn.ReadFromUDP(buf)
				if err != nil {
					fmt.Printf("Error receiving data: %v\n", err)
					continue
				}
				fmt.Printf("Received data from %s: %s\n", addr, string(buf[:n]))
			}
		}
	}()

	fmt.Println("Attempting NAT hole punching...")
	for i := 0; i < 5; i++ {
		msg := fmt.Sprintf("Hello, peer! Attempt %d", i+1)
		fmt.Printf("Sending message to peer: %s\n", msg)
		_, err := localConn.WriteToUDP([]byte(msg), peerAddr)
		if err != nil {
			fmt.Printf("Error sending message to peer: %v\n", err)
		}
		time.Sleep(1 * time.Second)
	}
	fmt.Println("NAT hole punching completed. Waiting for responses...")
	time.Sleep(30 * time.Second)
	close(done)
}
