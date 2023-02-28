## Broadcast Example

You need to understand the concept of [topics and channels](http://blog.charmes.net/2014/10/first-look-at-nsq.html) before you can appreciate NSQ.

Here is the code for broadcast example. You need to run one instance of consumer1 and consumer2 first before running the 
producer.

To run consumer, type ```./consumer1``` and so on.

To run producer, type ```./producer -msg=3```

You should see consumer1 and consumer2 getting the same 3 messages.

```go
// broadcast/consumer1.go

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

	c := queue.NewConsumer("mytopic", "mychannel1")

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

```go
// broadcast/consumer2.go

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
```


```go
// broadcast/producer.go

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
