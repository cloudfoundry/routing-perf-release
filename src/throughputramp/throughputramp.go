package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"throughputramp/aggregator"
	"throughputramp/data"
	"time"
)

var (
	numRequests        = flag.Int("n", 1234, "number of requests to send")
	concurrentRequests = flag.Int("c", 1234, "number of requests to send")
)

func main() {
	flag.Parse()
	if flag.NArg() < 1 {
		usageAndExit()
	}

	url := flag.Args()[0]
	args := []string{
		"-n", strconv.Itoa(*numRequests),
		"-c", strconv.Itoa(*concurrentRequests),
		"-o", "csv",
		url,
	}

	heyData, err := exec.Command("hey", args...).Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "hey error: %s\n", err.Error())
		fmt.Fprintf(os.Stderr, "%s\n", string(heyData))
		os.Exit(1)
	}

	dataPoints, err := data.Parse(string(heyData))
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse error: %s\n", err.Error())
		fmt.Fprintf(os.Stderr, "%s\n", string(heyData))
		os.Exit(1)
	}

	buckets := aggregator.NewBuckets(dataPoints, time.Second)
	report, err := json.Marshal(buckets.Summary())
	if err != nil {
		fmt.Fprintf(os.Stderr, "report marshaling error: %s\n", err.Error())
		os.Exit(1)
	}
	fmt.Println(string(report))
}

func usageAndExit() {
	flag.Usage()
	fmt.Fprintf(os.Stderr, "\n")
	os.Exit(1)
}
