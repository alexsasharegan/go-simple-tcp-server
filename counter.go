package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

// Counter is a container for tracking and managing the runtime counters.
type Counter struct {
	mu sync.RWMutex
	// Uniq is a map of the unique numbers received during uptime.
	Uniq map[int]bool
	// Cnt valid numbers received during uptime.
	Cnt int
	// IntvlCnt is the total valid numbers received during output interval.
	IntvlCnt int
	Log      *struct {
		// Cnt is the log rotation counter.
		Cnt int
		// w is a buffered writer to the current log entry
		w *bufio.Writer
		f io.Closer
	}
	intvl *struct {
		output  chan bool
		logging chan bool
	}
	// Sem is a semaphore to do request limiting.
	Sem chan int
}

var logFmt = "logs/data.%d.log"

// NewCounter constructs a new Counter.
func NewCounter(connLimit int) *Counter {
	f := openLogFile(fmt.Sprintf(logFmt, 0))
	return &Counter{
		Uniq: make(map[int]bool),
		Sem:  make(chan int, connLimit),
		Log: &struct {
			Cnt int
			w   *bufio.Writer
			f   io.Closer
		}{
			w: bufio.NewWriter(f),
			f: f,
		},
		intvl: &struct {
			output  chan bool
			logging chan bool
		}{
			output:  make(chan bool),
			logging: make(chan bool),
		},
	}
}

func openLogFile(name string) *os.File {
	f, err := os.OpenFile(
		name,
		// We only need to write to the log,
		// but we need to create if file does not exist,
		// or else truncate it if we're re-opening it on a new run.
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		0666)

	if err != nil {
		log.Fatalf("could not open log file: %v", err)
	}

	return f
}

// FlushClose writes the log contents to disk and closes the file.
func (c *Counter) FlushClose() (err error) {
	err = c.Log.w.Flush()
	if err != nil {
		return fmt.Errorf("could not flush log to disk: %v", err)
	}

	err = c.Log.f.Close()
	if err != nil {
		return fmt.Errorf("could not close log file: %v", err)
	}

	return
}

// FlushRotate writes the log contents to disk, closes, and rotates the log file.
func (c *Counter) FlushRotate() (err error) {
	err = c.FlushClose()
	if err != nil {
		return
	}

	c.Log.Cnt++
	f := openLogFile(fmt.Sprintf(logFmt, c.Log.Cnt))

	c.Log.f = f
	c.Log.w = bufio.NewWriter(f)

	return
}

// WriteInt writes a new unique value to the buffered writer.
func (c *Counter) WriteInt(i int) (err error) {
	_, err = c.Log.w.WriteString(fmt.Sprintf("%d\n", i))
	return
}

func (c *Counter) outputCounters() {
	c.mu.Lock()

	fmt.Printf(
		"Count unique: %d\n"+
			"Count total: %d\n"+
			"Count last: %d\n",
		len(c.Uniq),
		c.Cnt,
		c.IntvlCnt)
	c.IntvlCnt = 0

	c.mu.Unlock()
}

// RunOutputInterval outputs the counters on an interval.
// It takes a nil channel that the caller will close to stop execution.
// Must be run on go routine.
func (c *Counter) RunOutputInterval(intvl time.Duration) {
	for {
		select {
		case <-time.After(intvl):
			c.outputCounters()
		case <-c.intvl.output:
			return
		}
	}
}

// StopOutputIntvl exits the output interval by closing it's underlying nil channel.
func (c *Counter) StopOutputIntvl() {
	close(c.intvl.output)
}

// RunLogInterval outputs the counters on an interval.
// It takes a nil channel that the caller will close to stop execution.
// Must be run on go routine.
func (c *Counter) RunLogInterval(intvl time.Duration) {
	var err error
	for {
		select {
		case <-time.After(intvl):
			err = c.FlushRotate()
			if err != nil {
				log.Fatalf("could not flush and rotate logs: %v", err)
			}
		case <-c.intvl.logging:
			err = c.FlushClose()
			if err != nil {
				fmt.Fprintf(os.Stderr, "error flushing log to disk: %v", err)
			}
			return
		}
	}
}

// StopLogIntvl exits the output interval by closing it's underlying nil channel.
func (c *Counter) StopLogIntvl() {
	close(c.intvl.logging)
}

// Inc increments the counters in a thread safe way.
func (c *Counter) Inc() {
	c.mu.Lock()
	c.Cnt++
	c.IntvlCnt++
	c.mu.Unlock()
}

// RecUniq adds a unique int to the map in a thread safe way.
func (c *Counter) RecUniq(num int) {
	c.mu.Lock()
	c.Uniq[num] = true
	c.mu.Unlock()
}

// HasUniq checks if an int has been recorded in a thread safe way.
func (c *Counter) HasUniq(num int) (b bool) {
	c.mu.RLock()
	b = c.Uniq[num]
	c.mu.RUnlock()
	return
}
