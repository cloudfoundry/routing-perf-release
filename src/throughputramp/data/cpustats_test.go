package data_test

import (
	"encoding/json"
	"throughputramp/data"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var singlePercentageCsv = `timeStamp,percentage
2016-12-15 15:00:47.575579693 -0800 PST,12.358514
2016-12-15 15:00:47.672438722 -0800 PST,20.779221`

var multiplePercentageCsv = `timeStamp,percentage,percentage
2016-12-15 15:00:47.575579693 -0800 PST,12.358514,13.358514
2016-12-15 15:00:47.672438722 -0800 PST,20.779221,21.779221`

var _ = Describe("GenerateCpuCSV", func() {
	It("Returns an empty byte slice if empty byte passed", func() {
		emptyByte := data.GenerateCpuCSV([]byte(""))
		Expect(emptyByte).To(BeEmpty())
	})

	It("Returns an empty byte slice if nil byte passed", func() {
		emptyByte := data.GenerateCpuCSV(nil)
		Expect(emptyByte).To(BeNil())
	})

	Context("CSV format single Percentagent", func() {
		It("returns correctly formatted CSV output", func() {
			var timeStamp1, timeStamp2 time.Time
			var err error

			timeStamp1, err = time.Parse(time.RFC3339, "2016-12-15T15:00:47.575579693-08:00")
			Expect(err).ToNot(HaveOccurred())
			timeStamp2, err = time.Parse(time.RFC3339, "2016-12-15T15:00:47.672438722-08:00")
			Expect(err).ToNot(HaveOccurred())
			cpuStats := []data.CPUStat{
				data.CPUStat{TimeStamp: timeStamp1, Percentage: []float64{12.358514295296388}},
				data.CPUStat{TimeStamp: timeStamp2, Percentage: []float64{20.77922077922078}},
			}
			jsonData, err := json.Marshal(cpuStats)
			Expect(err).ToNot(HaveOccurred())
			result := data.GenerateCpuCSV(jsonData)
			Expect(string(result)).To(Equal(singlePercentageCsv))
		})
	})

	Context("CSV format multiple Percentagent", func() {
		It("returns correctly formatted CSV output", func() {
			var timeStamp1, timeStamp2 time.Time
			var err error

			timeStamp1, err = time.Parse(time.RFC3339, "2016-12-15T15:00:47.575579693-08:00")
			Expect(err).ToNot(HaveOccurred())
			timeStamp2, err = time.Parse(time.RFC3339, "2016-12-15T15:00:47.672438722-08:00")
			Expect(err).ToNot(HaveOccurred())
			cpuStats := []data.CPUStat{
				data.CPUStat{TimeStamp: timeStamp1, Percentage: []float64{12.358514295296388, 13.358514295296388}},
				data.CPUStat{TimeStamp: timeStamp2, Percentage: []float64{20.77922077922078, 21.77922077922078}},
			}
			jsonData, err := json.Marshal(cpuStats)
			Expect(err).ToNot(HaveOccurred())
			result := data.GenerateCpuCSV(jsonData)
			Expect(string(result)).To(Equal(multiplePercentageCsv))
		})
	})
})
