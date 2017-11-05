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
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)

	// Semaphore
	sem := make(chan int, connLimit)
	// Receive new connections on a chan.
	conns := acceptConns(srv, sem)

	for {
		select {
		case conn := <-conns:
			go handleConnection(conn, sem)
		case <-sig:
			// Signals generally print their escape sequence to stdout,
			// so add a leading new line.
			fmt.Printf("\nShutting down server.\n")
			os.Exit(0)
		}
	}
}

func acceptConns(srv net.Listener, sem chan<- int) <-chan net.Conn {
	conns := make(chan net.Conn)

	go func() {
		for {
			conn, err := srv.Accept()
			if err != nil {
				fmt.Printf("Error accepting connection: %v\n", err)
				continue
			}
			// Try pushing a marker value on our semaphore,
			// if we haven't hit our connection limit,
			// pass the connection onto the channel,
			// otherwise close it.
			select {
			case sem <- 1:
				conns <- conn
			default:
				fmt.Fprintf(conn, "Server busy.")
				conn.Close()
			}
		}
	}()

	return conns
}

// Handles incoming requests.
func handleConnection(conn net.Conn, sem <-chan int) {
	// Defer all close logic
	defer func() {
		// Since handleConnection is run in a go routine,
		// it manages the closing of our net.Conn.
		conn.Close()
		// Once our connection is closed,
		// we can drain a value from our semaphore
		// to free up a space in the connection limit.
		<-sem
	}()

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
