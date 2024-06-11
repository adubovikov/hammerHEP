package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync/atomic"
	"time"

	"github.com/adubovikov/hammerHEP/publish"
)

// Internal Stats
type Statistics struct {
	TotalPackets atomic.Uint64
	timeStart    time.Time
	timeStop     time.Time
}

var hammerStats Statistics

func main() {
	var (
		port        = flag.String("port", "9060", "Destination Port")
		addr        = flag.String("address", "localhost", "Destination Address")
		rate        = flag.Int("rate", 16, "Packets per second")
		generateID  = flag.Bool("replace-callid", false, "Generate a new callID")
		replaceTime = flag.Bool("replace-time", false, "set to current time")
		replaceIP   = flag.Bool("replace-ip", false, "generate fake IP")
		proto       = flag.String("protocol", "HEP", "Possible protocols are HEP,IPFIX, FILE-TXT")
		fileTxt     = flag.String("file", "", "Generate calls from file")
		trans       = flag.String("transport", "TLS", "Possible transports are UDP,TCP,TLS")
	)
	flag.Parse()

	if len(*port) < 1 || len(*addr) < 1 || len(*proto) < 1 || len(*trans) < 1 || *rate < 1 {
		fmt.Println("Invalid flags provided!")
		os.Exit(1)
	}

	replace := publish.ReplaceParams{
		ReplaceCid:  *generateID,
		ReplaceTime: *replaceTime,
		ReplaceIP:   *replaceIP,
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			fmt.Printf("Got %s signal. Aborting...\n", sig)
			hammerStats.timeStop = time.Now()
			fmt.Printf("Period from: %v, to %v\n", hammerStats.timeStart.Local(), hammerStats.timeStop.Local())
			fmt.Printf("Seconds: %d\n", hammerStats.timeStop.Unix()-hammerStats.timeStart.Unix())
			fmt.Printf("Sent packets: %d\n", hammerStats.TotalPackets.Load())
			os.Exit(1)
		}
	}()

	hammerStats.timeStart = time.Now()

	hammer, err := NewHammer(*proto, *addr, *port, *trans, *rate, *fileTxt, replace)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("Hammer down %s at %s over %s with %d pps\n", *proto, *addr+":"+*port, *trans, *rate)
	hammer.start()
}
