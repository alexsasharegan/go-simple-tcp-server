package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/alexsasharegan/go-simple-tcp-server/sync"
)

const (
	host      = "localhost"
	port      = "3280"
	connLimit = 6
)

func main() {
	handleSignalClose()
	l, err := net.Listen("tcp", host+":"+port)
	if err != nil {
		log.Fatalf("Error listening: %v", err)
	}

	fmt.Println("Listening on " + host + ":" + port)
	defer l.Close()

	swg := sync.NewSizedWaitGroup(connLimit)

	for {
		swg.Add(1)
		conn, err := l.Accept()
		if err != nil {
			log.Fatalf("Error accepting connection: %v", err)
		}

		go func() {
			defer swg.Done()
			handleRequest(conn)
		}()
	}
}

// Handles incoming requests.
func handleRequest(conn net.Conn) {
	defer conn.Close()
	// Make a buffer to hold incoming data.
	buf := make([]byte, 1024)
	// Read the incoming connection into the buffer.
	_, err := conn.Read(buf)
	if err != nil {
		log.Fatalf("Error reading: %v", err)
	}
	if err != nil {
		log.Fatalf("could not parse request: %v", err)
	}
	conn.Write([]byte(fmt.Sprintf("Message received: %s", string(buf))))
	err = logToFile(buf)
	if err != nil {
		log.Fatalf("could not log request: %v", err)
	}
}

func logToFile(b []byte) error {
	f, err := os.OpenFile(fmt.Sprintf("data.%d.log", 0), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("could not open log file: %v", err)
	}
	defer f.Close()

	f.Write(b)
	return nil
}

// handleSignalClose sets up a signal listener for shutdown signals.
// It returns a bool channel to communicate shutdown and allow the caller to perform tear down.
func handleSignalClose() <-chan bool {
	sigc := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigc, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-sigc
		fmt.Println("graceful shutdown")
		done <- true
		os.Exit(0)
	}()

	return done
}
