package main

import (
	"fmt"
	_log "log"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// cleaner logger for user facing stuff, this is a client after all
var log = _log.New(os.Stderr, "", 0)

var b = make([]byte, 4096)
var ping = []byte("ping")

var host string

var singlePattern = regexp.MustCompile(`^\d+$`)
var listPattern = regexp.MustCompile(`^\d+(,\d+)+$`)
var rangePattern = regexp.MustCompile(`^\d+-\d+$`)

var openPorts = make(map[int]struct{})
var openPortsMutex = sync.Mutex{}

func main() {
	if len(os.Args) < 3 {
		log.Fatalf("usage: %v HOSTNAME PORT\n", os.Args[0])
	}
	host = os.Args[1]
	portsUnparsed := os.Args[2]

	ports := parsePorts(portsUnparsed)

	// TODO semaphore, the system doesn't like having very many connections at once
	// https://pkg.go.dev/golang.org/x/sync/semaphore
	wg := &sync.WaitGroup{}
	for _, p := range ports {
		wg.Add(1)
		go func(port int) {
			checkPort(port)
			wg.Done()
		}(p)
	}
	wg.Wait()

	fmt.Println("open ports:")
	printOpen(ports)
	fmt.Println("filtered ports:")
	printFiltered(ports)
}

func checkPort(port int) {
	raddr, err := net.ResolveUDPAddr("udp", net.JoinHostPort(host, strconv.Itoa(port)))
	if err != nil {
		log.Println("address resolution error:", err)
		return
	}
	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		log.Println("connect error:", err)
		return
	}
	conn.Write(ping)

	conn.SetReadDeadline(time.Now().Add(time.Second * 5))
	_, err = conn.Read(b)
	if err != nil {
		if err, ok := err.(net.Error); ok && err.Timeout() {
			// fmt.Println("port", port, "filtered")
		} else {
			log.Printf("error on port %v: %v\n", port, err)
		}
	} else {
		// fmt.Println("port", port, "open")
		openPortsMutex.Lock()
		openPorts[port] = struct{}{}
		openPortsMutex.Unlock()
	}
}

func parsePorts(s string) []int {
	if singlePattern.MatchString(s) {
		return parseSinglePort(s)
	} else if listPattern.MatchString(s) {
		return parseListPort(s)
	} else if rangePattern.MatchString(s) {
		return parseRangePort(s)
	} else {
		log.Fatalln("unknown port format")
		return nil // unreachable
	}
}

func parseSinglePort(s string) []int {
	port, err := strconv.Atoi(s)
	if err != nil {
		log.Fatalln(err)
	}
	return []int{port}
}

func parseListPort(s string) []int {
	portsStr := strings.Split(s, ",")
	ports := make([]int, len(portsStr))
	for i := range ports {
		port, err := strconv.Atoi(portsStr[i])
		if err != nil {
			log.Fatalln(err)
		}
		ports[i] = port
	}
	return ports
}

func parseRangePort(s string) []int {
	portsStr := strings.Split(s, "-")
	if len(portsStr) != 2 {
		log.Fatalln("invalid port range syntax")
	}
	from, err := strconv.Atoi(portsStr[0])
	if err != nil {
		log.Fatalln(err)
	}
	to, err := strconv.Atoi(portsStr[1])
	if err != nil {
		log.Fatalln(err)
	}
	// swap if decreasing range
	if to < from {
		tmp := to
		to = from
		from = tmp
	}
	ports := make([]int, to-from+1)
	port := from
	for i := range ports {
		ports[i] = port
		port++
	}
	return ports
}

func printFiltered(ports []int) {
	for _, port := range ports {
		if _, ok := openPorts[port]; !ok {
			fmt.Println(port)
		}
	}
}

func printOpen(ports []int) {
	for _, port := range ports {
		if _, ok := openPorts[port]; ok {
			fmt.Println(port)
		}
	}
}
