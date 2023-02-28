## Simulating NSQ Consumer Error

Let us simulate an NSQ consumer encountering an error (for example, network partition, database access timeout, etc) and see how NSQ works.

### Run NSQ

```./nsqlookupd & ./nsqd --lookupd-tcp-address=127.0.0.1:4160 & ./nsqadmin --lookupd-http-address=127.0.0.1:4161```

### Run NSQ Producer

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

```./producer -msg=20```

Output

```bash
2015/12/02 08:07:18 INF    1 (192.168.56.101:4150) connecting to nsqd
2015/12/02 08:07:19 Time took 279.880274ms
2015/12/02 08:07:19 INF    1 stopping
2015/12/02 08:07:19 INF    1 exiting router
2015/12/02 08:07:19 ERR    1 (192.168.56.101:4150) IO error - EOF
2015/12/02 08:07:19 INF    1 (192.168.56.101:4150) beginning close
2015/12/02 08:07:19 INF    1 (192.168.56.101:4150) readLoop exiting
2015/12/02 08:07:19 INF    1 (192.168.56.101:4150) breaking out of writeLoop
2015/12/02 08:07:19 INF    1 (192.168.56.101:4150) writeLoop exiting                        
2015/12/02 08:07:19 INF    1 (192.168.56.101:4150) finished draining, cleanup exiting
2015/12/02 08:07:19 INF    1 (192.168.56.101:4150) clean close complete
```

### Run NSQ Consumer with Error

This program reads messages received from NSQ producer, counts the number of vowels
in each message and returns (simulates) an error if the number of vowels is even (not odd).

```go
//cerror.go

package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/ilhamsj/go-queue"
	"github.com/nsqio/go-nsq"
	"runtime"
	"strconv"
)

var numbPtr = flag.Int("msg", 100, "number of messages (default: 100)")
var lkp = flag.String("lkp", "", "IP address of nsqlookupd")

func FuncSimulateError(msg *nsq.Message) error {

	str := string(msg.Body)

	//http://play.golang.org/p/MFboCiikYW
	t := 0 //number of vowels in str
	for _, value := range str {
		switch value {
		case 'a', 'e', 'i', 'o', 'u', 'A', 'E', 'I', 'O', 'U':
			t++
		}
	}

	if t%2 == 0 {
		return errors.New(strconv.Itoa(t))
	} else {
		return nil
	}
}

func main() {

	flag.Parse()

	c := queue.NewConsumer("test", "ch")

	c.Set("nsqlookupd", *lkp+":4161")
	c.Set("concurrency", runtime.GOMAXPROCS(runtime.NumCPU()))
	c.Set("max_attempts", 10)
	c.Set("max_in_flight", 150)
	c.Set("default_requeue_delay", "15s")

	c.Start(nsq.HandlerFunc(FuncSimulateError))

	fmt.Scanln()

	c.Stop()
}
```

Run ```./cerror -msg=20```

Output

```bash
2015/12/02 08:07:40 INF    1 [test/ch] querying nsqlookupd http://:4161/lookup?topic=test                                             
2015/12/02 08:07:40 INF    1 [test/ch] (core:4150) connecting to nsqd                                                                 
2015/12/02 08:07:40 ERR    1 [test/ch] Handler returned error (38) for msg 09534a96354c3066                                           
2015/12/02 08:07:40 ERR    1 [test/ch] Handler returned error (40) for msg 09534a96354c3067                                           
2015/12/02 08:07:40 WRN    1 [test/ch] backing off for 2.0000 seconds (backoff level 1), setting all to RDY 0                         
2015/12/02 08:07:40 ERR    1 [test/ch] Handler returned error (54) for msg 09534a96354c306a                                           
2015/12/02 08:07:40 ERR    1 [test/ch] Handler returned error (46) for msg 09534a96354c306d                                           
2015/12/02 08:07:40 ERR    1 [test/ch] Handler returned error (40) for msg 09534a96354c3073                                           
2015/12/02 08:07:40 ERR    1 [test/ch] Handler returned error (52) for msg 09534a96354c3074                                           
2015/12/02 08:07:42 WRN    1 [test/ch] (core:4150) backoff timeout expired, sending RDY 1                                             
2015/12/02 08:07:55 ERR    1 [test/ch] Handler returned error (38) for msg 09534a96354c3066                                           
2015/12/02 08:07:55 WRN    1 [test/ch] backing off for 4.0000 seconds (backoff level 2), setting all to RDY 0                         
2015/12/02 08:07:59 WRN    1 [test/ch] (core:4150) backoff timeout expired, sending RDY 1                                             
2015/12/02 08:08:00 ERR    1 [test/ch] Handler returned error (40) for msg 09534a96354c3067                                           
2015/12/02 08:08:00 WRN    1 [test/ch] backing off for 8.0000 seconds (backoff level 3), setting all to RDY 0                         
2015/12/02 08:08:08 WRN    1 [test/ch] (core:4150) backoff timeout expired, sending RDY 1                                             
2015/12/02 08:08:08 ERR    1 [test/ch] Handler returned error (54) for msg 09534a96354c306a                                           
2015/12/02 08:08:08 WRN    1 [test/ch] backing off for 16.0000 seconds (backoff level 4), setting all to RDY 0                        
2015/12/02 08:08:24 WRN    1 [test/ch] (core:4150) backoff timeout expired, sending RDY 1                                             
2015/12/02 08:08:24 ERR    1 [test/ch] Handler returned error (46) for msg 09534a96354c306d                                           
2015/12/02 08:08:24 WRN    1 [test/ch] backing off for 32.0000 seconds (backoff level 5), setting all to RDY 0                        
2015/12/02 08:08:42 INF    1 [test/ch] querying nsqlookupd http://:4161/lookup?topic=test                                             
2015/12/02 08:08:56 WRN    1 [test/ch] (core:4150) backoff timeout expired, sending RDY 1                                             
2015/12/02 08:08:56 ERR    1 [test/ch] Handler returned error (40) for msg 09534a96354c3073                                           
2015/12/02 08:08:56 WRN    1 [test/ch] backing off for 64.0000 seconds (backoff level 6), setting all to RDY 0                        
2015/12/02 08:09:42 INF    1 [test/ch] querying nsqlookupd http://:4161/lookup?topic=test                                             
2015/12/02 08:10:00 WRN    1 [test/ch] (core:4150) backoff timeout expired, sending RDY 1                                             
2015/12/02 08:10:00 ERR    1 [test/ch] Handler returned error (52) for msg 09534a96354c3074                                           
2015/12/02 08:10:00 WRN    1 [test/ch] backing off for 64.0000 seconds (backoff level 6), setting all to RDY 0                        
2015/12/02 08:10:42 INF    1 [test/ch] querying nsqlookupd http://:4161/lookup?topic=test  
```

![cerror](https://itjumpstart.files.wordpress.com/2015/12/cerror.png)

Looking at the output of NSQ and NSQAdmin, you can see that there are six instances of errors (that is, those that returned even number of vowels).

By default, NSQ has ```-msg-timeout``` flag of 60 seconds which means it expects the message to be processed successfully within a minute. If not, the message will be requeued . In practice, there are other possible [scenarios](http://word.bitly.com/post/50027069647/building-nsq-client-libraries#message_handling). 

In our example, the purpose is to illustrate how NSQ works when it encounters an error. 

When a certain message has encountered consecutive errors, nsqd will try to requeue the message. In our program, there is a configurable ```default_requeue_delay``` of 15 seconds.

See https://godoc.org/github.com/nsqio/go-nsq for more configuration options.

```bash
2015/12/02 08:07:40 ERR    1 [test/ch] Handler returned error (38) for msg 09534a96354c3066 

...

2015/12/02 08:07:55 ERR    1 [test/ch] Handler returned error (38) for msg 09534a96354c3066
```

### But Wait...There's More

Since NSQ is a complex state machine, for now it's suffice to say that in our simple 
NSQ consumer example, NSQ exhibits requeue with backoff.

That is, when an NSQ consumer encountered consecutive errors, it would slow down consuming messages from nsqd for some increasing duration of time (exponential backoff).

Therefore, it is up to the NSQ consumer how to handle the error. One possible scenario is to delegate the message to a data store for persistence, and signal to nsqd that it is done with the message (return nil error).

This is the case where the circuit breaker pattern becomes invaluable.
