package main_test

import (
	"os/exec"
	"strconv"

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
	Context("Argument handling", func() {
		It("errors if no port number is passed", func() {
			session := runCpumonitor(0, 0, 0, false)
			Eventually(session).Should(gexec.Exit(1))
			Expect(session.Err).To(gbytes.Say("-port must be provided"))
		})

		It("errors if port number is not valid", func() {
			session := runCpumonitor(65536, 0, 0, false)
			Eventually(session).Should(gexec.Exit(1))
			Expect(session.Err).To(gbytes.Say("-port must be valid"))
		})
	})
})
