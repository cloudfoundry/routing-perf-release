package main_test

import (
	"io/ioutil"
	"net/http"
	"os/exec"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"
)

var _ = Describe("Throughputramp", func() {
	var (
		runner       *ginkgomon.Runner
		process      ifrit.Process
		testServer   *ghttp.Server
		testS3Server *ghttp.Server
		bodyChan     chan []byte
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

			bodyChan = make(chan []byte, 2)
			bucketName := "blah-bucket"

			testS3Server = ghttp.NewServer()

			testHandlers := []http.HandlerFunc{
				ghttp.VerifyHeaderKV("X-Amz-Acl", "public-read"),
				func(rw http.ResponseWriter, req *http.Request) {
					defer GinkgoRecover()
					defer req.Body.Close()
					bodyBytes, err := ioutil.ReadAll(req.Body)
					Expect(err).ToNot(HaveOccurred())
					bodyChan <- bodyBytes
				},
				ghttp.RespondWith(http.StatusOK, nil),
			}
			testS3Server.AppendHandlers(
				ghttp.CombineHandlers(testHandlers...),
				ghttp.CombineHandlers(append([]http.HandlerFunc{ghttp.VerifyContentType("image/png")}, testHandlers...)...),
			)

			runner = NewThroughputRamp(binPath, Args{
				NumRequests:      12,
				RateLimit:        100,
				StartConcurrency: 2,
				EndConcurrency:   4,
				ConcurrencyStep:  2,
				Proxy:            testServer.URL(),
				URL:              url,
				BucketName:       bucketName,
				Endpoint:         testS3Server.URL(),
				AccessKeyID:      "ABCD",
				SecretAccessKey:  "ABCD",
			})
		})

		JustBeforeEach(func() {
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

		It("sends the csv and plot to the s3 bucket", func() {
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
