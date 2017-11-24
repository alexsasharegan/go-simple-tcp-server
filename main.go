package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

const (
	port      = 3280
	connLimit = 6
	validLen  = 10
	minValue  = 1000000
	outIntvl  = 5 * time.Second
	logIntvl  = 10 * time.Second
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
	go counter.RunOutputInterval(outIntvl)
	go counter.RunLogInterval(logIntvl)

	// Receive new connections on an unbuffered channel.
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

// acceptConns uses the semaphore channel on the counter to rate limit.
// New connections get sent on the returned channel.
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
// Input is parsed and written to log if unique.
// Handles closing of the connection.
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

	scanner := bufio.NewScanner(conn)
	var s string
	for scanner.Scan() {
		s = scanner.Text()

		// Malformed Request: invalid length
		// Digit chars are safe for counting via len()
		if len(s) != validLen {
			continue
		}

		num, err := strconv.Atoi(s)
		// Malformed Request: not a number
		if err != nil {
			continue
		}

		// Malformed Request: less than minimum
		if num < minValue {
			continue
		}

		/* From here on out, we have a valid input. */
		// Safely increment total counter.
		counter.Inc()

		// Check if input has been recorded previously.
		if counter.HasValue(num) {
			continue
		}

		// Record the new unique value.
		// In this case, logging is part of our reqs.
		// We should fail is we didn't get this right.
		if err = counter.RecordUniq(num); err != nil {
			log.Fatalf("could not log unique value: %v\n", err)
		}
	}

	// If a failure to read input occurs,
	// it's probably my bad.
	// Fail and figure it out if so!
	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading: %v", err)
	}
}
