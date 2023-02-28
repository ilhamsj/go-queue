### Basic Example of using NSQ

This post was inspired from [Traun Leyden's example](http://tleyden.github.io/blog/2014/11/12/an-example-of-using-nsq-from-go/) but uses Go client library derived from https://github.com/segmentio/go-queue.

You need to understand the concept of [topics and channels](http://blog.charmes.net/2014/10/first-look-at-nsq.html) before you can appreciate NSQ.

Note: You can choose any folder location of your own choosing. For this guide, we assume some defaults:

- ```/home/nsq``` - location of NSQ binaries
- ```/home/nsqapps``` - location of NSQ clients (producers and consumers)
- for single-node NSQ installation, there is a [script](https://github.com/ibmendoza/go-examples/blob/master/nsq/startup.sh) to run nsqlookup, nsqd and nsqadmin for convenience


```./nsqlookupd & ./nsqd --lookupd-tcp-address=127.0.0.1:4160 & ./nsqadmin --lookupd-http-address=127.0.0.1:4161```


### Install NSQ

- Download NSQ from https://github.com/nsqio/nsq/releases to your host computer
- Run Turnkey Linux Core 14 VM under VirtualBox (VM in this guide uses 128MB only)
- Using VirtualBox host-only networking, you should have a default IP address like 192.168.56.101
- Open a Web browser, type https://192.168.56.101:12320 to run Web Shell from TKLinux VM
- Create a directory named ```/home/nsq``` and ```/home/nsqapps``` at TKLinux VM
- From a separate Web browser tab, type https://192.168.56.101:12321 to run Webmin from TKLinux VM
- Using Webmin, upload ```nsq-0.3.6.linux-amd64.go1.5.1.tar.gz``` and extract it to ```/home/nsq```

### Create an NSQ producer

This NSQ producer produces random 320-character messages up to a configurable number (default: 100).


```go 
//NSQ Producer Client
//producer.go

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
        w.Publish("test", []byte(randSeq(320)))
    }

    elapsed := time.Since(start)
    log.Printf("Time took %s", elapsed)

    w.Stop()

    fmt.Scanln()
}
```

### Run NSQ producer

Assuming you place this producer.go under producer folder of your Go workspace and build it using LiteIDE cross-linux64 option 
(remember that NSQ binaries are 64-bit only), the binary will be named producer.
You then need to upload that binary to TKLinux VM under ```/home/nsqapps``` and make it executable by using 
```chmod +x producer``` for example.

To run NSQ producer with 20 messages,

```./producer -msg=20```

Output:

```bash
2015/12/02 04:50:23 INF    1 (192.168.56.101:4150) connecting to nsqd
2015/12/02 04:50:23 Time took 22.152598ms
2015/12/02 04:50:23 INF    1 stopping
2015/12/02 04:50:23 INF    1 exiting router
2015/12/02 04:50:23 ERR    1 (192.168.56.101:4150) IO error - EOF
2015/12/02 04:50:23 INF    1 (192.168.56.101:4150) beginning close
2015/12/02 04:50:23 INF    1 (192.168.56.101:4150) readLoop exiting
2015/12/02 04:50:23 INF    1 (192.168.56.101:4150) breaking out of writeLoop
2015/12/02 04:50:23 INF    1 (192.168.56.101:4150) writeLoop exiting
2015/12/02 04:50:23 INF    1 (192.168.56.101:4150) finished draining, cleanup exiting 
2015/12/02 04:50:23 INF    1 (192.168.56.101:4150) clean close complete
```


### NSQ Admin

![NSQAdmin](https://itjumpstart.files.wordpress.com/2015/12/nsqadmin.png)

Note: With NSQ, you need not concern of message loss if an NSQ consumer has run later than the NSQ producer. However, since NSQ is in-memory 
message queue, in the event of a clean shutdown of nsqd (for example, with a Ctrl-C), those pending messages in memory may be persisted to disk.

See image below after a clean shutdown of nsqd.

![nsqd clean shutdown](https://itjumpstart.files.wordpress.com/2015/12/nsqd-cleanshutdown.png)


### Run NSQ Consumer

```go
package main

import (
	"flag"
	"fmt"
	"github.com/ilhamsj/go-queue"
	"github.com/nsqio/go-nsq"
	"log"
	"runtime"
	"sync/atomic"
	"time"
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
```

Run ```./consumer -msg=20```

Output:

```
2015/12/02 05:31:06 INF    1 [test/ch] querying nsqlookupd http://:4161/lookup?topic=test
2015/12/02 05:31:06 INF    1 [test/ch] (core:4150) connecting to nsqd                       
```

Looking at NSQAdmin, as long as messages are in memory or on disk, an NSQ consumer will 
be able to process those pending messages.

![nsqconsumer](https://itjumpstart.files.wordpress.com/2015/12/nsqconsumer.png)

![nsqconsumer-msgprocessed](https://itjumpstart.files.wordpress.com/2015/12/nsqconsumer-msgprocessed.png)

### Little Experiment

Up to this time, we had run an NSQ producer which produces an arbitrary number of messages (in this case, 20).

```./producer -msg=20```

Then, we process those messages with an NSQ consumer.

```./consumer -msg=20```

Now, try to run the same NSQ producer for the second time.

```./producer -msg=20```

This is what you will see in NSQAdmin.

![nsqproducer-2ndtime](https://itjumpstart.files.wordpress.com/2015/12/nsqproducer-2ndtime.png)

![nsqproducer-2ndprocess](https://itjumpstart.files.wordpress.com/2015/12/nsqproducer-2ndprocess.png)

We have seen that NSQ reported 40 Messages Processed. From NSQ point of view,
we have **enqueued** 40 messages all in all although those 20 messages are yet to be processed by an NSQ consumer (you can see in NSQAdmin that those messages are still in memory).

So running ```./consumer -msg=20``` again will process those 20 pending messages.

![nsqconsumer-2ndprocess](https://itjumpstart.files.wordpress.com/2015/12/nsqconsumer-2ndprocess.png)
