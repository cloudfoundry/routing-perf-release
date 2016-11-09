package data_test

import (
	"throughputramp/data"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var input string = `start-time,response-time
2016-11-01T21:04:42.760279114Z,0.0280
2016-11-01T21:04:42.760213269Z,0.0280
2016-11-01T21:04:42.760373651Z,0.0279
2016-11-01T21:04:42.760159771Z,0.0282
2016-11-01T21:04:42.760090065Z,0.0291
2016-11-01T21:04:42.788256168Z,0.0138
2016-11-01T21:04:42.788331398Z,0.0137
2016-11-01T21:04:42.788291332Z,0.0138
2016-11-01T21:04:42.788256153Z,0.0141
2016-11-01T21:04:42.789231777Z,0.0136`

var extraColumnInput string = `start-time,response-time,dummy-header
2016-11-01T21:04:42.760279114Z,0.0280,0.123`

var badHeadersInput string = `dummy-header,dummy-header-2
2016-11-01T21:04:42.760279114Z,0.0280`

var missingColumnInput string = `start-time
2016-11-01T21:04:42.760279114Z`

var badValuesInput string = `start-time,response-time
2016-11-01T21:04:42.760279114Z`

var _ = Describe("Data", func() {
	Describe("Parse", func() {
		It("unmarshals the input correctly", func() {
			dataPoints, err := data.Parse(input)
			Expect(err).NotTo(HaveOccurred())
			Expect(dataPoints).To(ConsistOf(
				data.Point{time.Date(2016, 11, 1, 21, 04, 42, 760279114, time.UTC), time.Duration(28000000)},
				data.Point{time.Date(2016, 11, 1, 21, 04, 42, 760213269, time.UTC), time.Duration(28000000)},
				data.Point{time.Date(2016, 11, 1, 21, 04, 42, 760373651, time.UTC), time.Duration(27900000)},
				data.Point{time.Date(2016, 11, 1, 21, 04, 42, 760159771, time.UTC), time.Duration(28200000)},
				data.Point{time.Date(2016, 11, 1, 21, 04, 42, 760090065, time.UTC), time.Duration(29100000)},
				data.Point{time.Date(2016, 11, 1, 21, 04, 42, 788256168, time.UTC), time.Duration(13800000)},
				data.Point{time.Date(2016, 11, 1, 21, 04, 42, 788331398, time.UTC), time.Duration(13700000)},
				data.Point{time.Date(2016, 11, 1, 21, 04, 42, 788291332, time.UTC), time.Duration(13800000)},
				data.Point{time.Date(2016, 11, 1, 21, 04, 42, 788256153, time.UTC), time.Duration(14100000)},
				data.Point{time.Date(2016, 11, 1, 21, 04, 42, 789231777, time.UTC), time.Duration(13600000)},
			))
		})

		Context("when the input has additional columns", func() {
			It("unmarshals the input correctly", func() {
				dataPoints, err := data.Parse(extraColumnInput)
				Expect(err).NotTo(HaveOccurred())
				Expect(dataPoints).To(ConsistOf(
					data.Point{time.Date(2016, 11, 1, 21, 04, 42, 760279114, time.UTC), time.Duration(28000000)},
				))
			})
		})

		Context("when the input does not have the required headers", func() {
			It("returns an error", func() {
				_, err := data.Parse(badHeadersInput)
				Expect(err).To(MatchError("csv headers not found"))
			})
		})

		Context("when the input is missing a column", func() {
			It("returns an error", func() {
				_, err := data.Parse(missingColumnInput)
				Expect(err).To(MatchError("csv headers not found"))
			})
		})

		Context("when the input is badly formatted", func() {
			It("returns an error", func() {
				_, err := data.Parse(badValuesInput)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when the input is empty", func() {
			It("returns an error", func() {
				_, err := data.Parse("")
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Sort", func() {
		It("sorts the data by start time", func() {
			dataPoints := []data.Point{
				data.Point{time.Unix(1478020922, 270882000), time.Duration(19100000)},
				data.Point{time.Unix(1478020922, 288598000), time.Duration(6300000)},
				data.Point{time.Unix(1478020923, 0), time.Duration(10000000)},
				data.Point{time.Unix(1478020922, 270864000), time.Duration(19900000)},
			}

			data.Sort(dataPoints)

			Expect(dataPoints).To(Equal([]data.Point{
				data.Point{time.Unix(1478020922, 270864000), time.Duration(19900000)},
				data.Point{time.Unix(1478020922, 270882000), time.Duration(19100000)},
				data.Point{time.Unix(1478020922, 288598000), time.Duration(6300000)},
				data.Point{time.Unix(1478020923, 0), time.Duration(10000000)},
			}))
		})
	})
})
