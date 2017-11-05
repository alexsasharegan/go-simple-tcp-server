package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
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

	fmt.Printf("Started %s server.\nListening on %s\n", srv.Addr().Network(), srv.Addr().String())
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
			go handleConnection(conn, swg)
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
		for {
			conn, err := srv.Accept()
			if err != nil {
				fmt.Printf("Error accepting connection: %v\n", err)
				continue
			}
			netc <- conn
		}
	}()

	return netc
}

// Handles incoming requests.
func handleConnection(conn net.Conn, swg *sync.SizedWaitGroup) {
	defer swg.Done()
	defer conn.Close()

	var (
		str string
		i   int
		err error
		s   = bufio.NewScanner(conn)
	)

	for s.Scan() {
		str = strings.TrimSpace(s.Text())
		if str == "" {
			fmt.Fprintf(conn, "malformed request\n")
			return
		}
		i, err = strconv.Atoi(str)
		if err != nil {
			fmt.Fprintf(conn, "malformed request\n")
			return
		}
		break
	}

	if err := s.Err(); err != nil {
		log.Fatalf("Error reading: %v", err)
	}

	fmt.Fprintf(conn, "%d\n", i)
	err = logToFile(fmt.Sprintf("%d", i))
	if err != nil {
		fmt.Printf("could not log request: %v\n", err)
	}
}

func logToFile(s string) error {
	f, err := os.OpenFile(fmt.Sprintf("logs/data.%d.log", 0), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("could not open log file: %v", err)
	}
	defer f.Close()

	f.WriteString(s + "\n")
	return nil
}
