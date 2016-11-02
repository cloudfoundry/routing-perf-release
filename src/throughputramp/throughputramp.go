package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"throughputramp/aggregator"
	"throughputramp/data"
	"time"
)

func main() {
	heyData, err := exec.Command("hey", os.Args[1:]...).Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "hey error: %s\n", err.Error())
		fmt.Fprintf(os.Stderr, "%s\n", string(heyData))
		os.Exit(1)
	}

	fmt.Println(string(heyData))
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
