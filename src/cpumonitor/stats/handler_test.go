package stats_test

import (
	"cpumonitor/fakes"
	"cpumonitor/stats"
	"encoding/json"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Handler", func() {
	var (
		cpuCollector *fakes.FakeCollector
		statsHandler *stats.Handler
	)

	BeforeEach(func() {
		cpuCollector = new(fakes.FakeCollector)
		statsHandler = stats.NewStatHandler(cpuCollector)
	})

	Describe("Start", func() {
		It("returns 200 status if CPUCollector started successfully", func() {
			testServer := httptest.NewServer(http.HandlerFunc(statsHandler.Start))
			resp, err := http.Get(testServer.URL)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(responseString(resp.Body)).To(ContainSubstring("Collecting CPU stats"))
		})

		It("does not start CPUCollector if already started", func() {
			testServer := httptest.NewServer(http.HandlerFunc(statsHandler.Start))
			resp, err := http.Get(testServer.URL)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			resp, err = http.Get(testServer.URL)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))

			Expect(responseString(resp.Body)).To(ContainSubstring(stats.ErrMultipleCollector.Error()))

			Expect(cpuCollector.RunCallCount()).To(Equal(1))
		})

		//It("returns non 200 if CPUCollector returns an error ", func() {
		//	cpuCollector := new(fakes.FakeCollector)
		//	cpuCollector.RunReturns(errors.New("bad stuff"))
		//	statsHandler := stats.NewStatHandler(cpuCollector)
		//	testServer := httptest.NewServer(http.HandlerFunc(statsHandler.Start))
		//	resp, err := http.Get(testServer.URL)

		//	Expect(err).ToNot(HaveOccurred())
		//	Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		//})
	})

	Describe("Stop", func() {
		It("returns 200 status when CPUCollector is started", func() {
			testServer := httptest.NewServer(http.HandlerFunc(statsHandler.Start))
			http.Get(testServer.URL)

			testServer = httptest.NewServer(http.HandlerFunc(statsHandler.Stop))
			resp, err := http.Get(testServer.URL)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			Expect(cpuCollector.ResultCallCount()).To(Equal(1))
		})

		It("returns json CPU data when CPUCollector is started", func() {
			testServer := httptest.NewServer(http.HandlerFunc(statsHandler.Start))
			http.Get(testServer.URL)

			testData := testData(5)
			cpuCollector.ResultReturns(testData)

			testServer = httptest.NewServer(http.HandlerFunc(statsHandler.Stop))
			resp, err := http.Get(testServer.URL)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(resp.Header.Get("Content-Type")).To(Equal("application/json"))
			Expect(testData).To(Equal(unmarshallResponse(resp.Body)))
		})

		It("does not call result if CPUCollector is not running", func() {
			testServer := httptest.NewServer(http.HandlerFunc(statsHandler.Stop))
			resp, err := http.Get(testServer.URL)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			Expect(responseString(resp.Body)).To(ContainSubstring(stats.ErrCollectorNotRunning.Error()))
			Expect(cpuCollector.ResultCallCount()).To(Equal(0))
		})
	})
})

func responseString(body io.ReadCloser) string {
	responseContent, err := ioutil.ReadAll(body)
	Expect(err).ToNot(HaveOccurred())
	return string(responseContent)
}

func unmarshallResponse(body io.ReadCloser) []stats.Info {
	var results []stats.Info
	content, err := ioutil.ReadAll(body)
	Expect(err).ToNot(HaveOccurred())
	err = json.Unmarshal(content, &results)
	Expect(err).ToNot(HaveOccurred())
	return results
}

func testData(entry int) []stats.Info {
	seeded := rand.NewSource(time.Now().UnixNano())
	r := rand.New(seeded)
	var results []stats.Info
	for i := 0; i <= entry; i++ {
		results = append(results, stats.Info{
			Percentage: []float64{r.Float64()},
			TimeStamp:  time.Now(),
		})
	}
	return results
}
