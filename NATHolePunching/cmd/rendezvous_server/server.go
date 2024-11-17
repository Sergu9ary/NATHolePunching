package main

import (
	"flag"
	"fmt"
	"net"
	"sync"
)

func main() {
	mode := flag.String("mode", "rendezvous", "Mode: rendezvous, udp-client-nat")
	flag.Parse()

	switch *mode {
	case "rendezvous":
		runRendezvousServer()
	default:
		fmt.Println("Unknown mode:", *mode)
	}
}

var (
	clientList []*net.UDPAddr
	mu         sync.Mutex
)

func runRendezvousServer() {
	addr, err := net.ResolveUDPAddr("udp", ":3478")
	if err != nil {
		fmt.Println("Error resolving server address:", err)
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("Error starting UDP server:", err)
		return
	}
	defer conn.Close()
	fmt.Println("Rendezvous server started at")

	buf := make([]byte, 1024)
	for {
		n, clientAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Println("Error receiving data:", err)
			continue
		}

		message := string(buf[:n])
		fmt.Printf("Received '%s' from %s\n", message, clientAddr)

		if message == "REGISTER" {
			mu.Lock()
			clientList = append(clientList, clientAddr)
			if len(clientList) == 2 {
				fmt.Println("Both clients registered. Sharing addresses.")
				conn.WriteToUDP([]byte(clientList[1].String()), clientList[0])
				conn.WriteToUDP([]byte(clientList[0].String()), clientList[1])
				clientList = nil
			}
			mu.Unlock()
		}
	}
}