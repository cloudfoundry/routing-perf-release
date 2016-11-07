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
	numRequests        = flag.Int("n", 1000, "number of requests to send")
	concurrentRequests = flag.Int("c", 10, "number of requests to send")
	lowerThroughput    = flag.Int("lower-throughput", 50, "Starting throughput value")
	upperThroughput    = flag.Int("upper-throughput", 200, "Ending throughput value")
	throughputStep     = flag.Int("throughput-step", 50, "Throughput increase per run")
)

func main() {
	flag.Parse()
	if flag.NArg() < 1 {
		usageAndExit()
	}

	url := flag.Args()[0]

	start := *lowerThroughput
	end := *upperThroughput
	step := *throughputStep

	var dataPoints []data.Point
	for i := start; i <= end; i += step {
		points, err := runBenchmark(url, *numRequests, *concurrentRequests, i)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			os.Exit(1)
		}

		dataPoints = append(dataPoints, points...)
	}

	buckets := aggregator.NewBuckets(dataPoints, time.Second)
	report, err := json.Marshal(buckets.Summary())
	if err != nil {
		fmt.Fprintf(os.Stderr, "report marshaling error: %s\n", err.Error())
		os.Exit(1)
	}
	fmt.Println(string(report))
}

func runBenchmark(url string, numRequests, concurrentRequests, rateLimit int) ([]data.Point, error) {
	fmt.Fprintf(os.Stderr, "Running benchmark with %d requests, %d concurrency, and %d throughput\n", numRequests, concurrentRequests, rateLimit)
	args := []string{
		"-n", strconv.Itoa(numRequests),
		"-c", strconv.Itoa(concurrentRequests),
		"-q", strconv.Itoa(rateLimit),
		"-o", "csv",
		url,
	}

	heyData, err := exec.Command("hey", args...).Output()
	if err != nil {
		return nil, fmt.Errorf("hey error: %s\nData:\n%s\n", err.Error(), string(heyData))
	}

	dataPoints, err := data.Parse(string(heyData))
	if err != nil {
		return nil, fmt.Errorf("parse error: %s\nData:\n%s\n", err.Error(), string(heyData))
	}

	return dataPoints, nil
}

func usageAndExit() {
	flag.Usage()
	fmt.Fprintf(os.Stderr, "\n")
	os.Exit(1)
}
