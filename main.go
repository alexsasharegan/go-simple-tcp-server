package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

const (
	port        = 3280
	connLimit   = 6
	validLen    = 10
	minValue    = 1000000
	outputIntvl = 5 * time.Second
	logIntvl    = 10 * time.Second
)

func main() {
	// Start up the tcp server.
	srv, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Error listening: %v", err)
	}

	fmt.Printf(
		"Started %s server.\nListening on %s\n",
		srv.Addr().Network(), srv.Addr().String())
	defer srv.Close()

	counter := NewCounter(connLimit)

	// Listen for termination signals.
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)

	// Set up intervals
	go counter.RunOutputInterval(outputIntvl)
	go counter.RunLogInterval(logIntvl)

	// Receive new connections on a chan.
	conns := acceptConns(srv, counter)

	for {
		select {
		case conn := <-conns:
			go handleConnection(conn, counter)
		case <-sig:
			// Add a leading new line since the signal escape sequence prints on stdout.
			fmt.Printf("\nShutting down server.\n")
			counter.Close()
			os.Exit(0)
		}
	}
}

func acceptConns(srv net.Listener, counter *Counter) <-chan net.Conn {
	conns := make(chan net.Conn)

	go func() {
		for {
			conn, err := srv.Accept()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error accepting connection: %v\n", err)
				continue
			}

			select {
			case counter.Sem <- 1:
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
func handleConnection(conn net.Conn, counter *Counter) {
	// Defer all close logic.
	// Using a closure makes it easy to group logic as well as execute serially
	// and avoid the deferred LIFO exec order.
	defer func() {
		// Since handleConnection is run in a go routine,
		// it manages the closing of our net.Conn.
		conn.Close()
		// Once our connection is closed,
		// we can drain a value from our semaphore
		// to free up a space in the connection limit.
		<-counter.Sem
	}()

	r := bufio.NewReader(conn)
	s, err := r.ReadString('\n')
	// Failure to read input is probably my bad. Exit(1)
	if err != nil && err != io.EOF {
		log.Fatalf("Error reading: %v", err)
	}

	// Digit chars as safe for counting via len()
	if len(s) != validLen {
		fmt.Fprintf(conn, "ERR Malformed Request: expected length %d, got %d.\n", validLen, len(s))
		return
	}

	num, err := strconv.Atoi(s)
	if err != nil {
		fmt.Fprintf(conn, "ERR Malformed Request: expected number\n")
		return
	}

	if num < minValue {
		fmt.Fprintf(conn, "ERR Malformed Request: expected number greater than %d\n", minValue)
		return
	}

	/* From here on out, we have a valid input. */

	// Increment total counter safely
	counter.Inc()

	// Echo input back to conn.
	fmt.Fprintf(conn, "%d\n", num)

	// Check if input has been recorded prev.
	if counter.HasUniq(num) {
		return
	}
	// Record the new unique value.
	counter.RecUniq(num)

	// In this case, logging is part of our reqs.
	// We should fail is we didn't get this right.
	if err = counter.WriteInt(num); err != nil {
		log.Fatalf("could not log unique value: %v\n", err)
	}
}
