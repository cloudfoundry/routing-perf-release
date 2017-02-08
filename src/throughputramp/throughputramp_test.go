package main_test

import (
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/ghttp"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"
)

var cpuMonitorData = `
[{"Percentage":[12.358514295296388, 19.1234123],"Timestamp":"2016-12-15T15:00:47.575579693-08:00"},
{"Percentage":[20.77922077922078, 22.23345],"Timestamp":"2016-12-15T15:00:47.672438722-08:00"}]
`

var _ = Describe("Throughputramp", func() {
	var (
		runner          *ginkgomon.Runner
		process         ifrit.Process
		testServer      *ghttp.Server
		testS3Server    *ghttp.Server
		bodyChan        chan []byte
		runnerArgs      Args
		bodyTestHandler http.HandlerFunc
	)

	Context("when correct arguments are used", func() {
		BeforeEach(func() {
			url := "http://example.com"
			testServer = ghttp.NewUnstartedServer()
			handler := ghttp.CombineHandlers(
				func(rw http.ResponseWriter, req *http.Request) {
					Expect(req.Host).To(Equal(strings.TrimPrefix(url, "http://")))
				},
				ghttp.RespondWith(http.StatusOK, nil),
			)
			testServer.AppendHandlers(handler)
			testServer.AllowUnhandledRequests = true
			testServer.Start()

			bodyChan = make(chan []byte, 3)

			testS3Server = ghttp.NewServer()

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
		})

		It("ramps up throughput over multiple tests", func() {
			Eventually(process.Wait(), "5s").Should(Receive())
			Expect(runner.ExitCode()).To(Equal(0))
			Expect(testServer.ReceivedRequests()).To(HaveLen(24))
		})

		Context("when local-csv is specified", func() {
			var dir string
			BeforeEach(func() {
				var err error
				dir, err = ioutil.TempDir("", "test")
				Expect(err).NotTo(HaveOccurred())
				runnerArgs.localCSV = dir
				cpumonitorServer := ghttp.NewServer()

				header := make(http.Header)
				header.Add("Content-Type", "application/json")

				cpumonitorServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/stop"),
						ghttp.RespondWith(http.StatusOK, cpuMonitorData, header),
					),
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
				testS3Server.AppendHandlers(
					bodyTestHandler,
				)
			})
			It("stores the csv locally", func() {

				checkFiles := func() int {
					files, err := ioutil.ReadDir(dir)
					Expect(err).ToNot(HaveOccurred())
					fileCount := 0
					for _, file := range files {
						if strings.Contains(file.Name(), "csv") {
							fileCount++
						}
					}
					return fileCount
				}
				Eventually(checkFiles).Should(Equal(2))
				Expect(os.RemoveAll(dir)).To(Succeed())
			})
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
						ghttp.VerifyRequest("GET", "/stop"),
						ghttp.RespondWith(http.StatusOK, cpuMonitorData, header),
					),
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
				testS3Server.AppendHandlers(
					bodyTestHandler,
				)
			})

			AfterEach(func() {
				cpumonitorServer.Close()
			})

			It("calls cpumonitor server start & stop endpoints", func() {
				Eventually(process.Wait(), "5s").Should(Receive())
				//stop, start, stop
				Expect(cpumonitorServer.ReceivedRequests()).To(HaveLen(3))
			})

			It("uploads the csv of cpuStats to s3 bucket", func() {
				Eventually(process.Wait(), "5s").Should(Receive())
				Expect(runner.ExitCode()).To(Equal(0))

				Eventually(bodyChan).Should(Receive())

				var cpuCsvBytes []byte
				Eventually(bodyChan).Should(Receive(&cpuCsvBytes))
				Expect(cpuCsvBytes).ToNot(BeEmpty())
				Expect(string(cpuCsvBytes)).To(ContainSubstring("timestamp,percentage,percentage\n"))
			})
		})

		It("uploads the csv to the s3 bucket", func() {
			Eventually(process.Wait(), "5s").Should(Receive())
			Expect(runner.ExitCode()).To(Equal(0))

			var csvBytes []byte
			Eventually(bodyChan).Should(Receive(&csvBytes))
			Expect(csvBytes).ToNot(BeEmpty())
			b := gbytes.BufferWithBytes(csvBytes)
			Expect(b).To(gbytes.Say(`start-time,response-time\n`))
			// Make sure the second csv header appears as well
			Expect(b).To(gbytes.Say(`\nstart-time,response-time\n`))
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
