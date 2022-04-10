package main

import (
	"bytes"
	"log"
	"net"
)

var pong = []byte("pong")
var ping = []byte("ping")

func main() {
	for i := range intRange(0, 1<<16) {
		go serve(i)
	}
	select {}
}

func serve(port int) {
	b := make([]byte, 512) // safe size https://stackoverflow.com/a/1099359
	lis, err := net.ListenUDP("udp", &net.UDPAddr{Port: port})
	if err != nil {
		log.Println("listen error:", err)
		return
	}
	for {
		n, addr, err := lis.ReadFromUDP(b)
		if err != nil {
			log.Println("udp read error", err)
		}
		if !bytes.Equal(b[:n], ping) {
			continue
		}
		lis.WriteToUDP(pong, addr)
	}
}

func intRange(from, to int) <-chan int {
	c := make(chan int)
	go func() {
		for i := from; i < to; i++ {
			c <- i
		}
	}()
	return c
}
