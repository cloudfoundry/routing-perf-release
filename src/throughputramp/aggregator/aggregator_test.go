package aggregator_test

import (
	"throughputramp/aggregator"
	"throughputramp/data"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var dataPoints = []*data.Point{
	&data.Point{time.Date(2016, 11, 1, 21, 04, 42, 0, time.UTC), time.Duration(28000000)},
	&data.Point{time.Date(2016, 11, 1, 21, 04, 42, 760279114, time.UTC), time.Duration(28000000)},
	&data.Point{time.Date(2016, 11, 1, 21, 04, 43, 760213269, time.UTC), time.Duration(28000000)},
	&data.Point{time.Date(2016, 11, 1, 21, 04, 42, 760373651, time.UTC), time.Duration(27900000)},
	&data.Point{time.Date(2016, 11, 1, 21, 04, 43, 760159771, time.UTC), time.Duration(28200000)},
	&data.Point{time.Date(2016, 11, 1, 21, 04, 44, 760090065, time.UTC), time.Duration(29100000)},
	&data.Point{time.Date(2016, 11, 1, 21, 04, 44, 788256168, time.UTC), time.Duration(13800000)},
	&data.Point{time.Date(2016, 11, 1, 21, 04, 46, 788331398, time.UTC), time.Duration(13700000)},
	&data.Point{time.Date(2016, 11, 1, 21, 04, 45, 788291332, time.UTC), time.Duration(13800000)},
	&data.Point{time.Date(2016, 11, 1, 21, 04, 45, 788256153, time.UTC), time.Duration(14100000)},
	&data.Point{time.Date(2016, 11, 1, 21, 04, 46, 789231777, time.UTC), time.Duration(13600000)},
}

var _ = Describe("Aggregator", func() {
	Describe("Data", func() {
		It("returns data ordered by time that can be graphed in a throughput vs latency plot", func() {
			ag := aggregator.New(dataPoints, time.Second)
			report := ag.Data()

			expectedReport := aggregator.Report{
				aggregator.Point{Throughput: 3, Latency: time.Duration(28000000)},
				aggregator.Point{Throughput: 3, Latency: time.Duration(28000000)},
				aggregator.Point{Throughput: 3, Latency: time.Duration(27900000)},
				aggregator.Point{Throughput: 2, Latency: time.Duration(28200000)},
				aggregator.Point{Throughput: 2, Latency: time.Duration(28000000)},
				aggregator.Point{Throughput: 2, Latency: time.Duration(29100000)},
				aggregator.Point{Throughput: 2, Latency: time.Duration(13800000)},
				aggregator.Point{Throughput: 2, Latency: time.Duration(14100000)},
				aggregator.Point{Throughput: 2, Latency: time.Duration(13800000)},
				aggregator.Point{Throughput: 2, Latency: time.Duration(13700000)},
				aggregator.Point{Throughput: 2, Latency: time.Duration(13600000)},
			}

			Expect(report).To(HaveLen(len(expectedReport)))
			for i := range report {
				Expect(report[i]).To(Equal(expectedReport[i]))
			}
		})
	})

	Describe("GenerateCSV", func() {
		It("returns the data in a CSV format", func() {
			report := aggregator.Report{
				aggregator.Point{Throughput: 1, Latency: time.Duration(10000000)},
				aggregator.Point{Throughput: 2, Latency: time.Duration(20000000)},
				aggregator.Point{Throughput: 3, Latency: time.Duration(30000000)},
				aggregator.Point{Throughput: 4, Latency: time.Duration(40000000)},
				aggregator.Point{Throughput: 5, Latency: time.Duration(50000000)},
			}
			csv := gbytes.BufferWithBytes(report.GenerateCSV())

			Expect(csv).To(gbytes.Say(`throughput,latency\n`))
			Expect(csv).To(gbytes.Say(`(?m:^1,0\.010*$)`))
			Expect(csv).To(gbytes.Say(`(?m:^2,0\.020*$)`))
			Expect(csv).To(gbytes.Say(`(?m:^3,0\.030*$)`))
			Expect(csv).To(gbytes.Say(`(?m:^4,0\.040*$)`))
			Expect(csv).To(gbytes.Say(`(?m:^5,0\.050*$)`))
		})
	})
})
