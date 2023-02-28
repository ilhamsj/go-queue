// broadcast/consumer2.go

// https://itjumpstart.wordpress.com/category/nsq/

package main

import (
	"flag"
	"fmt"
	"log"
	"runtime"
	"sync/atomic"

	"github.com/ibmendoza/go-lib"
	"github.com/ilhamsj/go-queue"
	"github.com/nsqio/go-nsq"
)

var ops uint64 = 0
var numbPtr = flag.Int("msg", 100, "number of messages (default: 10000)")

func main() {
	ipaddr, _ := lib.GetIPAddress()

	flag.Parse()

	c := queue.NewConsumer("mytopic", "mychannel2")

	c.Set("nsqlookupd", ipaddr+":4161")
	c.Set("concurrency", runtime.GOMAXPROCS(runtime.NumCPU()))
	c.Set("max_attempts", 10)
	c.Set("max_in_flight", 150)
	c.Set("default_requeue_delay", "15s")

	c.Start(nsq.HandlerFunc(func(msg *nsq.Message) error {

		log.Println(string(msg.Body))

		atomic.AddUint64(&ops, 1)

		return nil
	}))

	fmt.Scanln()

	c.Stop()

	fmt.Println(ops)
}
