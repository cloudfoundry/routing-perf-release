package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"

	"throughputramp/aggregator"
	"throughputramp/data"
	"throughputramp/plotgen"
	"throughputramp/uploader"
)

var (
	numRequests        = flag.Int("n", 1000, "number of requests to send")
	concurrentRequests = flag.Int("c", 10, "number of requests to send")
	proxy              = flag.String("x", "", "proxy for request")
	lowerThroughput    = flag.Int("lower-throughput", 50, "Starting throughput value")
	upperThroughput    = flag.Int("upper-throughput", 200, "Ending throughput value")
	throughputStep     = flag.Int("throughput-step", 50, "Throughput increase per run")
	s3Endpoint         = flag.String("s3-endpoint", "", "The endpoint for the S3 service to which plots will be uploaded.")
	s3Region           = flag.String("s3-region", "", "The region for the S3 service to which plots will be uploaded. If provided, endpoint is ignored.")
	bucketName         = flag.String("bucket-name", "", "Name of the bucket to which plots will be uploaded.")
	accessKeyID        = flag.String("access-key-id", "", "AccessKeyID for the S3 service.")
	secretAccessKey    = flag.String("secret-access-key", "", "SecretAccessKey for the S3 service.")
)

func main() {
	flag.Parse()
	if flag.NArg() < 1 {
		usageAndExit()
	}

	s3Config := &uploader.Config{
		Endpoint:        *s3Endpoint,
		AwsRegion:       *s3Region,
		BucketName:      *bucketName,
		AccessKeyID:     *accessKeyID,
		SecretAccessKey: *secretAccessKey,
	}
	err := s3Config.Validate()
	if err != nil {
		fmt.Fprintf(os.Stderr, "s3 config error: %s\n", err.Error())
		usageAndExit()
	}

	url := flag.Args()[0]

	start := *lowerThroughput
	end := *upperThroughput
	step := *throughputStep

	var dataPoints []data.Point
	for i := start; i <= end; i += step {
		points, err := runBenchmark(url, *proxy, *numRequests, *concurrentRequests, i)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			os.Exit(1)
		}

		dataPoints = append(dataPoints, points...)
	}

	buckets := aggregator.NewBuckets(dataPoints, time.Second)
	summary := buckets.Summary()

	filename := time.Now().UTC().Format(time.RFC3339)

	loc, err := uploader.Upload(s3Config, bytes.NewBufferString(summary.GenerateCSV()), filename+".csv")
	if err != nil {
		fmt.Fprintf(os.Stderr, "uploading to s3 error: %s\n", err.Error())
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "csv uploaded to %s\n", loc)

	plotFile, err := plotgen.Generate(summary)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate plot: %s", err.Error())
		os.Exit(1)
	}
	defer plotFile.Close()

	loc, err = uploader.Upload(s3Config, plotFile, filename+".png")
	if err != nil {
		fmt.Fprintf(os.Stderr, "uploading to s3 error: %s\n", err.Error())
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "png uploaded to %s\n", loc)
}

func runBenchmark(url, proxy string, numRequests, concurrentRequests, rateLimit int) ([]data.Point, error) {
	fmt.Fprintf(os.Stderr, "Running benchmark with %d requests, %d concurrency, and %d throughput\n", numRequests, concurrentRequests, rateLimit)
	args := []string{
		"-x", proxy,
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
