// based on https://github.com/danoctavian/c10k-bench/blob/master/go-bencher/tcp_bencher.go
package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

var (
	outNum uint64
	inNum  uint64
	stop   uint64
)

var letterRunes = []rune("0123456789")

// genIntString generates a string of ints for the given size
func genIntString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func main() {
	rand.Seed(time.Now().Unix())
	pins1 := make([]string, 1000000)
	for i := 0; i < 1000000; i++ {
		pins1[i] = genIntString(10) + "\n"
	}
	go func() {
		time.Sleep(time.Second * 10)
		atomic.StoreUint64(&stop, 1)
	}()

	var wg sync.WaitGroup
	for i := 0; i < 6; i++ {
		time.Sleep(500 * time.Millisecond)
		wg.Add(1)
		go func(waitGroup *sync.WaitGroup) {
			defer wg.Done()
			if conn, err := net.Dial("tcp", "localhost:3280"); err == nil {
				for _, i := range pins1 {
					_, err := conn.Write([]byte(i))
					atomic.AddUint64(&outNum, 1)
					if err != nil {
						log.Println(err)
						break
					}
					atomic.AddUint64(&inNum, 1)
				}
			} else {
				log.Println(err)
			}
		}(&wg)
	}
	wg.Wait()

	fmt.Println("Benchmarking:", "localhost:3280")
	fmt.Println(6, "clients, running", 10, "bytes,", 10, "sec.")
	fmt.Println("Speed:", outNum/uint64(10), "request/sec,", inNum/uint64(10), "response/sec")
	fmt.Println("Requests:", outNum)
	fmt.Println("Responses:", inNum)
}
