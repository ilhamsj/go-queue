package main

import (
	"flag"
	"fmt"
	"log"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/ilhamsj/go-queue"
	"github.com/nsqio/go-nsq"
)

var start = time.Now()
var ops uint64 = 0
var numbPtr = flag.Int("msg", 100, "number of messages (default: 10000)")
var ipnsqlookupd = flag.String("ipnsqlookupd", "", "IP address of ipnsqlookupd")

func main() {

	flag.Parse()

	c := queue.NewConsumer("test", "ch")

	c.Set("nsqlookupd", ":4161")
	c.Set("concurrency", runtime.GOMAXPROCS(runtime.NumCPU()))
	c.Set("max_attempts", 10)
	c.Set("max_in_flight", 150)
	c.Set("default_requeue_delay", "15s")

	c.Start(nsq.HandlerFunc(func(msg *nsq.Message) error {

		log.Println(string(msg.Body))

		atomic.AddUint64(&ops, 1)
		if ops == uint64(*numbPtr) {
			elapsed := time.Since(start)
			log.Printf("Time took %s", elapsed)
		}

		return nil
	}))

	fmt.Scanln()

	c.Stop()
}
