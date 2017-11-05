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
	port      = 3280
	connLimit = 6
)

func main() {
	// Start up the tcp server.
	srv, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Error listening: %v", err)
	}

	fmt.Printf("Listening on localhost:%d\n", port)
	defer srv.Close()

	// Listen for termination signals.
	sigc := make(chan os.Signal)
	signal.Notify(sigc, syscall.SIGTERM, syscall.SIGINT)

	// Sized wait groups block incrementing if their limit is reached.
	swg := sync.NewSizedWaitGroup(connLimit)
	// Communicate new connections on a chan of net.Conn.
	netc := acceptConns(srv)

listen:
	for {
		select {
		case conn := <-netc:
			swg.Add(1)
			go handleRequest(conn, swg)
			break
		case <-sigc:
			// Signals generally print their escape sequence to stdout,
			// so add a leading new line.
			fmt.Printf("\nShutting down server.\n")
			break listen
		}
	}
}

func acceptConns(srv net.Listener) <-chan net.Conn {
	netc := make(chan net.Conn)

	go func() {
		conn, err := srv.Accept()
		if err != nil {
			log.Fatalf("Error accepting connection: %v", err)
		}
		netc <- conn
	}()

	return netc
}

// Handles incoming requests.
func handleRequest(conn net.Conn, swg *sync.SizedWaitGroup) {
	defer conn.Close()
	defer swg.Done()
	// Make a buffer to hold incoming data.
	b := make([]byte, 1024)
	// Read the incoming connection into the buffer.
	_, err := conn.Read(b)
	if err != nil {
		log.Fatalf("Error reading: %v", err)
	}
	if err != nil {
		log.Fatalf("could not parse request: %v", err)
	}
	conn.Write([]byte(fmt.Sprintf("Message received: %s", string(b))))
	err = logToFile(b)
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
