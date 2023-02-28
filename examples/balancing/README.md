## Balancing Example

You need to understand the concept of [topics and channels](http://blog.charmes.net/2014/10/first-look-at-nsq.html) before you can appreciate NSQ.

Here is the code for balancing example. You need to run three instances of consumer first before running the producer.

```go
// balancing/consumer.go

package main

import (
	"flag"
	"fmt"
	"github.com/ibmendoza/go-lib"
	"github.com/ilhamsj/go-queue"
	"github.com/nsqio/go-nsq"
	"log"
	"runtime"
	"sync/atomic"
)

var ops uint64 = 0
var numbPtr = flag.Int("msg", 100, "number of messages (default: 10000)")

func main() {
	ipaddr, _ := lib.GetIPAddress()

	flag.Parse()

	c := queue.NewConsumer("mytopic", "mychannel")

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
```

To run consumer, open three instances of Web Shell. Then on each instance, type

```./consumer -msg=20```

```go
// balancing/producer.go

package main

import (
	"flag"
	"fmt"
	"github.com/ibmendoza/go-lib"
	"github.com/nsqio/go-nsq"
	"log"
	"math/rand"
	"time"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ!@#$%^&*()1234567890")
var numbPtr = flag.Int("msg", 100, "number of messages (default: 100)")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func main() {
	config := nsq.NewConfig()

	ipaddr, _ := lib.GetIPAddress()

	w, err := nsq.NewProducer(ipaddr+":4150", config)

	if err != nil {
		log.Fatal("Could not connect")
	}

	flag.Parse()

	start := time.Now()

	for i := 1; i <= *numbPtr; i++ {
		w.Publish("mytopic", []byte(randSeq(320)))
	}

	elapsed := time.Since(start)
	log.Printf("Time took %s", elapsed)

	w.Stop()

	fmt.Scanln()
}
```

To run producer, open another instance of Web Shell, and type

```./producer -msg=20```

You should see that 2 instances of consumer got 7 messages and one consumer got 6 messages.
