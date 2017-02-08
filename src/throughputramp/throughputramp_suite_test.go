package main_test

import (
	"os/exec"
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/tedsuo/ifrit/ginkgomon"

	"testing"
)

var (
	binPath string
)

func NewThroughputRamp(throughputRampPath string, args Args) *ginkgomon.Runner {
	return ginkgomon.New(ginkgomon.Config{
		Name:    "throughputramp",
		Command: exec.Command(throughputRampPath, args.ArgSlice()...),
	})
}

type Args struct {
	NumRequests      int
	RateLimit        int
	StartConcurrency int
	EndConcurrency   int
	ConcurrencyStep  int
	Proxy            string
	URL              string
	Endpoint         string
	BucketName       string
	AccessKeyID      string
	SecretAccessKey  string
	CPUMonitorURL    string
	localCSV         string
}

func (args Args) ArgSlice() []string {
	argSlice := []string{
		"-n", strconv.Itoa(args.NumRequests),
		"-q", strconv.Itoa(args.RateLimit),
		"-x", args.Proxy,
		"-lower-concurrency", strconv.Itoa(args.StartConcurrency),
		"-upper-concurrency", strconv.Itoa(args.EndConcurrency),
		"-concurrency-step", strconv.Itoa(args.ConcurrencyStep),
		"-s3-endpoint", args.Endpoint,
		"-bucket-name", args.BucketName,
		"-access-key-id", args.AccessKeyID,
		"-secret-access-key", args.SecretAccessKey,
		"-cpumonitor-url", args.CPUMonitorURL,
		"-local-csv", args.localCSV,
	}

	argSlice = append(argSlice, args.URL)
	return argSlice
}

func TestThroughputramp(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Throughputramp Suite")
}

var _ = BeforeSuite(func() {
	var err error
	binPath, err = gexec.Build("throughputramp", "-race")
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})
