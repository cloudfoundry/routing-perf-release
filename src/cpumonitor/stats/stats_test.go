package stats_test

import (
	"cpumonitor/fakes"
	"cpumonitor/stats"
	"errors"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Collector", func() {
	//TODO: move to handler test
	//	Context("Start", func() {
	//		It("returns 200 StatusOK", func() {
	//			cpustat := new(fakes.FakeCPUStats)
	//			statsHandler := stats.NewStatHandler(cpustat)
	//			testServer := httptest.NewServer(http.HandlerFunc(statsHandler.Start))
	//			resp, err := http.Get(testServer.URL)
	//			Expect(err).ToNot(HaveOccurred())
	//			Expect(resp.StatusCode).To(Equal(http.StatusOK))
	//		})
	//
	//		It("returns non 200 if error occurs", func() {
	//			cpustat := new(fakes.FakeCPUStats)
	//			cpustat.RunReturns(errors.New("error"))
	//			statsHandler := stats.NewStatHandler(cpustat)
	//			testServer := httptest.NewServer(http.HandlerFunc(statsHandler.Start))
	//			resp, err := http.Get(testServer.URL)
	//
	//			Expect(err).ToNot(HaveOccurred())
	//			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
	//		})
	//	})
	Describe("Run", func() {
		It("returns an error if cpu percent errors", func() {
			fakeCpu := new(fakes.FakeCpu)
			expectedErr := errors.New("something bad happened")
			fakeCpu.PercentReturns(nil, expectedErr)
			collector := stats.NewCollector(fakeCpu, time.Second, time.Second)

			Eventually(func() error { return collector.Run() }()).Should(Equal(expectedErr))
		})

		It("uses specified cpuInterval", func() {
			fakeCpu := new(fakes.FakeCpu)
			cpuInterval := time.Second
			collector := stats.NewCollector(fakeCpu, time.Millisecond, cpuInterval)
			go collector.Run()

			time.Sleep(time.Second)
			interval, _ := fakeCpu.PercentArgsForCall(0)
			Expect(interval).To(Equal(cpuInterval))
		})

		It("uses specified interval", func() {
			fakeCpu := new(fakes.FakeCpu)
			cpuInterval := time.Second
			collector := stats.NewCollector(fakeCpu, time.Millisecond*100, cpuInterval)
			go collector.Run()

			time.Sleep(time.Second)
			Expect(fakeCpu.PercentCallCount()).To(BeNumerically(">=", 9))
		})
	})

	Describe("Collect", func() {
		It("returns a slice of Info", func() {

			fakeCpu := new(fakes.FakeCpu)
			stat := []float64{1.1, 2.2}
			fakeCpu.PercentReturns(stat, nil)
			collector := stats.NewCollector(fakeCpu, time.Millisecond*100, time.Second)

			done := make(chan struct{})
			go func() {
				collector.Run()
				close(done)
			}()

			time.Sleep(time.Second)
			result := collector.Result()
			<-done
			Expect(len(result)).ToNot(Equal(0))
			for i := 0; i < len(result); i++ {
				Expect(result[i].Percentage).To(Equal(stat))
			}
		})
	})
})
