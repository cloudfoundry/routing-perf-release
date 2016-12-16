package main_test

import (
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"
)

const comparisonCSV = `
throughput,latency
100, 0.01
200, 0.02
300, 0.03
400, 0.04
500, 0.05
600, 0.06
`

var cpuMonitorData = `
[{"Percentage":[12.358514295296388, 19.1234123],"TimeStamp":"2016-12-15T15:00:47.575579693-08:00"},
{"Percentage":[20.77922077922078, 22.23345],"TimeStamp":"2016-12-15T15:00:47.672438722-08:00"}]
`

var _ = Describe("Throughputramp", func() {
	var (
		runner             *ginkgomon.Runner
		process            ifrit.Process
		testServer         *ghttp.Server
		testS3Server       *ghttp.Server
		comparisonFilePath string
		bodyChan           chan []byte
		runnerArgs         Args
		bodyTestHandler    http.HandlerFunc
	)

	Context("when correct arguments are used", func() {
		BeforeEach(func() {
			url := "http://example.com"
			testServer = ghttp.NewServer()
			handler := ghttp.CombineHandlers(
				func(rw http.ResponseWriter, req *http.Request) {
					Expect(req.Host).To(Equal(strings.TrimPrefix(url, "http://")))
				},
				ghttp.RespondWith(http.StatusOK, nil),
			)
			testServer.AppendHandlers(handler)
			testServer.AllowUnhandledRequests = true

			bodyChan = make(chan []byte, 3)

			testS3Server = ghttp.NewServer()

			comparisonFile, err := ioutil.TempFile("", "comparison.csv")
			Expect(err).ToNot(HaveOccurred())

			comparisonFile.WriteString(comparisonCSV)
			comparisonFilePath = comparisonFile.Name()

			bodyTestHandler = ghttp.CombineHandlers(
				ghttp.VerifyHeaderKV("X-Amz-Acl", "public-read"),
				func(rw http.ResponseWriter, req *http.Request) {
					defer GinkgoRecover()
					defer req.Body.Close()
					bodyBytes, err := ioutil.ReadAll(req.Body)
					Expect(err).ToNot(HaveOccurred())
					bodyChan <- bodyBytes
				},
				ghttp.RespondWith(http.StatusOK, nil),
			)
			testS3Server.AppendHandlers(
				bodyTestHandler,
				ghttp.CombineHandlers(
					ghttp.VerifyContentType("image/png"),
					bodyTestHandler,
				),
			)

			runnerArgs = Args{
				NumRequests:      12,
				RateLimit:        100,
				StartConcurrency: 2,
				EndConcurrency:   4,
				ConcurrencyStep:  2,
				Proxy:            testServer.URL(),
				URL:              url,
				BucketName:       "blah-bucket",
				Endpoint:         testS3Server.URL(),
				AccessKeyID:      "ABCD",
				SecretAccessKey:  "ABCD",
				ComparisonFile:   comparisonFilePath,
			}
		})

		JustBeforeEach(func() {
			runner = NewThroughputRamp(binPath, runnerArgs)
			process = ginkgomon.Invoke(runner)
		})

		AfterEach(func() {
			ginkgomon.Interrupt(process)
			testServer.Close()
			testS3Server.Close()
			close(bodyChan)
			err := os.Remove(comparisonFilePath)
			Expect(err).ToNot(HaveOccurred())
		})

		It("ramps up throughput over multiple tests", func() {
			Eventually(process.Wait(), "5s").Should(Receive())
			Expect(runner.ExitCode()).To(Equal(0))
			Expect(testServer.ReceivedRequests()).To(HaveLen(24))
		})

		Context("when cpu monitor server is configured", func() {
			var (
				cpumonitorServer *ghttp.Server
			)
			BeforeEach(func() {
				cpumonitorServer = ghttp.NewServer()

				header := make(http.Header)
				header.Add("Content-Type", "application/json")

				cpumonitorServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/start"),
						ghttp.RespondWith(http.StatusOK, nil),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/stop"),
						ghttp.RespondWith(http.StatusOK, cpuMonitorData, header),
					),
				)

				runnerArgs.CPUMonitorURL = cpumonitorServer.URL()
				testS3Server.SetHandler(1, bodyTestHandler)
				testS3Server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyContentType("image/png"),
						bodyTestHandler,
					),
				)
			})
			AfterEach(func() {
				cpumonitorServer.Close()
			})
			It("calls cpumonitor server start & stop endpoints", func() {
				Eventually(process.Wait(), "5s").Should(Receive())
				Expect(cpumonitorServer.ReceivedRequests()).To(HaveLen(2))
			})

			It("uploads the csv of cpuStats to s3 bucket", func() {
				Eventually(process.Wait(), "5s").Should(Receive())
				Expect(runner.ExitCode()).To(Equal(0))

				var cpuCsvBytes []byte
				Eventually(bodyChan).Should(Receive(&cpuCsvBytes))
				Expect(cpuCsvBytes).ToNot(BeEmpty())
				Expect(string(cpuCsvBytes)).To(ContainSubstring("timeStamp,percentage,percentage\n"))
			})
		})

		It("uploads the csv and plot to the s3 bucket", func() {
			Eventually(process.Wait(), "5s").Should(Receive())
			Expect(runner.ExitCode()).To(Equal(0))

			var csvBytes []byte
			Eventually(bodyChan).Should(Receive(&csvBytes))
			Expect(csvBytes).ToNot(BeEmpty())
			Expect(string(csvBytes)).To(ContainSubstring("throughput,latency\n"))

			var pngBytes []byte
			Eventually(bodyChan).Should(Receive(&pngBytes))
			Expect(pngBytes).ToNot(BeEmpty())
			Expect(http.DetectContentType(pngBytes)).To(Equal("image/png"))
		})

		Context("but with no comparison data argument", func() {
			BeforeEach(func() {
				runnerArgs.ComparisonFile = ""
				runner = NewThroughputRamp(binPath, runnerArgs)
			})

			It("does not fail", func() {
				Eventually(process.Wait(), "5s").Should(Receive())
				Expect(runner.ExitCode()).To(Equal(0))
			})
		})

		Context("but with an incorrect comparison data argument", func() {
			BeforeEach(func() {
				runnerArgs.ComparisonFile = "/does/not/exist"
				runner = NewThroughputRamp(binPath, runnerArgs)
			})

			It("exits 1 with an error", func() {
				Eventually(process.Wait(), "5s").Should(Receive())
				Expect(runner.ExitCode()).To(Equal(1))
			})
		})
	})

	Context("when incorrect arguments are passed in", func() {
		BeforeEach(func() {
			runner = NewThroughputRamp(binPath, Args{})
			runner.Command = exec.Command(binPath)
		})

		It("exits 1 with usage", func() {
			process := ifrit.Background(runner)
			Eventually(process.Wait()).Should(Receive())
			Expect(runner.ExitCode()).To(Equal(1))
		})
	})

	Context("when the s3 config is not valid", func() {
		BeforeEach(func() {
			runner = NewThroughputRamp(binPath, Args{})
			runner.Command = exec.Command(binPath, "http://example.com")
		})

		It("exits 1 with usage", func() {
			process := ifrit.Background(runner)
			Eventually(process.Wait()).Should(Receive())
			Expect(runner.ExitCode()).To(Equal(1))
		})
	})
})
