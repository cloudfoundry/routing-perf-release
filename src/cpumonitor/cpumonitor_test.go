package main_test

import (
	"cpumonitor/stats"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var runCpumonitor = func(port, runInterval, cpuInterval int, perCpu bool) *gexec.Session {
	var session *gexec.Session
	var cpumonitorPath string
	var err error
	cpumonitorPath, err = gexec.Build("cpumonitor", "-race")
	Expect(err).ToNot(HaveOccurred())
	command := exec.Command(cpumonitorPath, "-port", strconv.Itoa(port), "-runInterval", strconv.Itoa(runInterval), "-cpuInterval", strconv.Itoa(cpuInterval), "-perCpu", strconv.FormatBool(perCpu))
	session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ToNot(HaveOccurred())
	return session
}

var _ = Describe("Cpumonitor", func() {
	var session *gexec.Session
	AfterEach(func() {
		session.Kill()
	})
	Context("Argument handling", func() {
		It("errors if no port number is passed", func() {
			session = runCpumonitor(0, 0, 0, false)
			Eventually(session).Should(gexec.Exit(1))
			Expect(session.Err).To(gbytes.Say("-port must be within range 1024-65535"))
		})

		It("errors if port number is not valid", func() {
			session = runCpumonitor(65536, 0, 0, false)
			Eventually(session).Should(gexec.Exit(1))
			Expect(session.Err).To(gbytes.Say("-port must be within range 1024-65535"))
		})

		It("errors if runtInterval is less than 1", func() {
			session = runCpumonitor(6553, 0, 0, false)
			Eventually(session).Should(gexec.Exit(1))
			Expect(session.Err).To(gbytes.Say("-runInterval must be above 1"))
		})
	})
	Context("when port and runInterval is passed", func() {
		BeforeEach(func() {
			session = runCpumonitor(6530, 100, 1, false)
			//wait for server to collect stats
			time.Sleep(time.Second * 1)

			Eventually(session.Err).Should(gbytes.Say("cpumonitor listening on 6530"))
		})

		It("returns an http error if start is called multiple times", func() {
			resp, err := http.Get("http://localhost:6530/start")
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			resp, err = http.Get("http://localhost:6530/start")
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			body, err := ioutil.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(body)).To(ContainSubstring(stats.ErrMultipleCollector.Error()))

		})

		It("returns http error if stop called without calling start", func() {
			var resp *http.Response
			var err error
			resp, err = http.Get("http://localhost:6530/stop")
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			var body []byte
			body, err = ioutil.ReadAll(resp.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(body)).To(ContainSubstring(stats.ErrCollectorNotRunning.Error()))
		})

		It("returns CPU stats", func() {
			resp, err := http.Get("http://localhost:6530/start")
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			resp, err = http.Get("http://localhost:6530/stop")
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			body, err := ioutil.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(body)).To(ContainSubstring("Percentage"))
			Expect(string(body)).To(ContainSubstring("TimeStamp"))
			session.Kill()

		})
	})
})
