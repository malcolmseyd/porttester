package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

var b = make([]byte, 4096)
var ping = []byte("ping")

func main() {
	if len(os.Args) < 3 {
		fmt.Printf("usage: %v HOSTNAME PORT\n", os.Args[0])
		os.Exit(1)
	}
	raddr, err := net.ResolveUDPAddr("udp", net.JoinHostPort(os.Args[1], os.Args[2]))
	if err != nil {
		log.Fatalln("address resolutoin error", err)
	}
	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		log.Fatalln("connect error", err)
	}
	conn.Write(ping)

	conn.SetReadDeadline(time.Now().Add(time.Second))
	_, err = conn.Read(b)
	if err != nil {
		if err, ok := err.(net.Error); ok && err.Timeout() {
			fmt.Println("port filtered")
		} else {
			log.Println("error:", err)
		}
	} else {
		fmt.Println("port open")
	}
}
