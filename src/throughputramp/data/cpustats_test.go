package data_test

import (
	"throughputramp/data"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var singlePercentageCSV = `timestamp,percentage
2016-12-15T23:00:47.575579693Z,12.358514
2016-12-15T23:00:47.672438722Z,20.779221`

var singlePercentageJSON = `[
{"TimeStamp":"2016-12-15T15:00:47.575579693-08:00","Percentage":[12.358514295296388]},
{"TimeStamp":"2016-12-15T15:00:47.672438722-08:00","Percentage":[20.77922077922078]}
]`

var multiplePercentageCSV = `timestamp,percentage,percentage
2016-12-15T23:00:47.575579693Z,12.358514,13.358514
2016-12-15T23:00:47.672438722Z,20.779221,21.779221`

var multiplePercentageJSON = `[
{"TimeStamp":"2016-12-15T15:00:47.575579693-08:00","Percentage":[12.358514295296388,13.358514295296388]},
{"TimeStamp":"2016-12-15T15:00:47.672438722-08:00","Percentage":[20.77922077922078,21.77922077922078]}
]`

var _ = Describe("GenerateCpuCSV", func() {
	It("returns an error and  empty byte slice if empty byte passed", func() {
		emptyByte, err := data.GenerateCpuCSV([]byte(""))
		Expect(err).To(HaveOccurred())
		Expect(emptyByte).To(BeEmpty())
	})

	It("returns an error and empty byte slice if nil byte passed", func() {
		emptyByte, err := data.GenerateCpuCSV([]byte(nil))
		Expect(err).To(HaveOccurred())
		Expect(emptyByte).To(BeNil())
	})

	It("returns an error if bad data passed", func() {
		badData := `timestamp, timestamp`
		emptyByte, err := data.GenerateCpuCSV([]byte(badData))
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("marshaling data"))
		Expect(emptyByte).To(BeNil())
	})

	Context("when formatting CSV with single percentages", func() {
		It("returns correctly formatted CSV output", func() {
			result, err := data.GenerateCpuCSV([]byte(singlePercentageJSON))
			Expect(err).ToNot(HaveOccurred())
			Expect(string(result)).To(Equal(singlePercentageCSV))
		})
	})

	Context("when formatting CSV with multiple percentages", func() {
		It("returns correctly formatted CSV output", func() {
			result, err := data.GenerateCpuCSV([]byte(multiplePercentageJSON))
			Expect(err).ToNot(HaveOccurred())
			Expect(string(result)).To(Equal(multiplePercentageCSV))
		})
	})
})
