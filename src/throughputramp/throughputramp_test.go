package main_test

import (
	"net/http"
	"net/http/httptest"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"
)

var _ = Describe("Throughputramp", func() {
	var (
		runner     *ginkgomon.Runner
		process    ifrit.Process
		testServer *httptest.Server
	)

	Context("when correct arguments are used", func() {
		BeforeEach(func() {
			testServer = httptest.NewServer(
				http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
					res.WriteHeader(http.StatusOK)
				}),
			)

			runner = NewThroughputRamp(binPath, Args{
				NumRequests:        1,
				ConcurrentRequests: 1,
				URL:                testServer.URL,
			})
		})

		JustBeforeEach(func() {
			process = ginkgomon.Invoke(runner)
		})

		AfterEach(func() {
			ginkgomon.Interrupt(process)
			testServer.Close()
		})

		It("prints throughput/latency data points", func() {
			<-process.Wait()
			Expect(runner.ExitCode()).To(Equal(0))
			Expect(runner.Buffer()).To(gbytes.Say(`\[\[.*\]\]`))
		})
	})

	Context("when incorrect arguments are passed in", func() {
		BeforeEach(func() {
			runner = NewThroughputRamp(binPath, Args{})
			runner.Command = exec.Command(binPath)
		})

		It("exits 1 with usage", func() {
			process := ifrit.Background(runner)
			<-process.Wait()
			Expect(runner.ExitCode()).To(Equal(1))
		})
	})
})
