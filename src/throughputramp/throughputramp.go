package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"throughputramp/data"
	"throughputramp/uploader"
)

var (
	numRequests      = flag.Int("n", 1000, "number of requests to send")
	proxy            = flag.String("x", "", "proxy for request")
	interval         = flag.Int("i", 1, "interval in seconds to average throughput")
	threadRateLimit  = flag.Int("q", 0, "thread rate limit")
	lowerConcurrency = flag.Int("lower-concurrency", 1, "Starting concurrency value")
	upperConcurrency = flag.Int("upper-concurrency", 30, "Ending concurrency value")
	concurrencyStep  = flag.Int("concurrency-step", 1, "Concurrency increase per run")
	s3Endpoint       = flag.String("s3-endpoint", "", "The endpoint for the S3 service to which plots will be uploaded.")
	s3Region         = flag.String("s3-region", "", "The region for the S3 service to which plots will be uploaded. If provided, endpoint is ignored.")
	bucketName       = flag.String("bucket-name", "", "Name of the bucket to which plots will be uploaded.")
	accessKeyID      = flag.String("access-key-id", "", "AccessKeyID for the S3 service.")
	secretAccessKey  = flag.String("secret-access-key", "", "SecretAccessKey for the S3 service.")
	cpuMonitorURL    = flag.String("cpumonitor-url", "", "Endpoint for monitoring CPU metrics")
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
		fmt.Fprintf(os.Stderr, "s3 config error: %s\n", err)
		usageAndExit()
	}

	url := flag.Args()[0]

	cpumonitorURL := strings.TrimPrefix(*cpuMonitorURL, "http://")

	runBenchmark(url,
		*proxy,
		cpumonitorURL,
		*numRequests,
		*lowerConcurrency,
		*upperConcurrency,
		*concurrencyStep,
		*threadRateLimit,
		s3Config)

}

func uploadCSV(s3config *uploader.Config, csvData io.Reader, cpuCsv []byte) {
	timeString := time.Now().UTC().Format(time.RFC3339)

	loc, err := uploader.Upload(s3config, csvData, timeString+".csv")
	if err != nil {
		fmt.Fprintf(os.Stderr, "uploading to s3 error: %s\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stdout, "csv uploaded to %s\n", loc)

	if len(cpuCsv) != 0 {
		cpuFilename := fmt.Sprintf("cpuStats-%s.csv", timeString)

		loc, err = uploader.Upload(s3config, bytes.NewBuffer(cpuCsv), cpuFilename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "uploading to s3 error: %s\n", err)
		}
		fmt.Fprintf(os.Stdout, "cpu csv uploaded to %s\n", loc)
	}
}

func runBenchmark(url,
	proxy,
	cpumonitorURL string,
	numRequests,
	lowerConcurrency,
	upperConcurrency,
	concurrencyStep,
	threshold int,
	uploaderConfig *uploader.Config) {

	if cpumonitorURL != "" {
		if err := startCPUMonitor(cpumonitorURL); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
	}

	benchmarkData := new(bytes.Buffer)
	for i := lowerConcurrency; i <= upperConcurrency; i += concurrencyStep {
		heyData, benchmarkErr := run(url, proxy, numRequests, i, threshold)
		if benchmarkErr != nil {
			fmt.Fprintf(os.Stderr, "%s\n", benchmarkErr)
			os.Exit(1)
		}

		_, writeErr := benchmarkData.Write(heyData)
		if benchmarkErr != nil {
			fmt.Fprintf(os.Stderr, "Buffer error: %s\n", writeErr)
			os.Exit(1)
		}
	}

	var cpuCsv []byte
	if cpumonitorURL != "" {
		var err error
		cpuCsv, err = stopCPUMonitor(cpumonitorURL, uploaderConfig)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
	}

	uploadCSV(uploaderConfig, benchmarkData, cpuCsv)
}

func run(url, proxy string, numRequests, concurrentRequests, rateLimit int) ([]byte, error) {
	fmt.Fprintf(os.Stdout, "Running benchmark with %d requests, %d concurrency, and %d rate limit\n", numRequests, concurrentRequests, rateLimit)
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
		return nil, fmt.Errorf("hey error: %s\nData:\n%s", err, string(heyData))
	}
	return heyData, nil
}

func usageAndExit() {
	flag.Usage()
	fmt.Fprintf(os.Stderr, "\n")
	os.Exit(1)
}

func startCPUMonitor(url string) error {
	startURL := fmt.Sprintf("http://%s/start", url)
	resp, err := http.Get(startURL)
	if err != nil {
		return fmt.Errorf("calling cpumonitor stop %s", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received resp %d", resp.StatusCode)
	}

	return nil
}

func stopCPUMonitor(url string, s3config *uploader.Config) ([]byte, error) {
	startURL := fmt.Sprintf("http://%s/stop", url)
	resp, err := http.Get(startURL)
	if err != nil {
		return nil, fmt.Errorf("calling cpumonitor stop %s", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received resp %d", resp.StatusCode)
	}

	rawData, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	csvData, err := data.GenerateCpuCSV(rawData)
	if err != nil {
		return nil, fmt.Errorf("GeneratateCpuCSV %d", resp.StatusCode)
	}

	return csvData, nil
}
