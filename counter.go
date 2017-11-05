package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
)

// Counter is a container for tracking and managing the runtime counters.
type Counter struct {
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

// FlushClose writes the log contents to disk, closes file, and clears counters.
func (c *Counter) FlushClose() (err error) {
	err = c.Log.w.Flush()
	if err != nil {
		return fmt.Errorf("could not flush log to disk: %v", err)
	}

	// Clear interval counter after the log has been flushed.
	c.IntvlCnt = 0

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
