package aggregator_test

import (
	"throughputramp/aggregator"
	"throughputramp/data"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var dataPoints []data.Point = []data.Point{
	data.Point{time.Date(2016, 11, 1, 21, 04, 42, 0, time.UTC), time.Duration(28000000)},
	data.Point{time.Date(2016, 11, 1, 21, 04, 42, 760279114, time.UTC), time.Duration(28000000)},
	data.Point{time.Date(2016, 11, 1, 21, 04, 43, 760213269, time.UTC), time.Duration(28000000)},
	data.Point{time.Date(2016, 11, 1, 21, 04, 42, 760373651, time.UTC), time.Duration(27900000)},
	data.Point{time.Date(2016, 11, 1, 21, 04, 43, 760159771, time.UTC), time.Duration(28200000)},
	data.Point{time.Date(2016, 11, 1, 21, 04, 44, 760090065, time.UTC), time.Duration(29100000)},
	data.Point{time.Date(2016, 11, 1, 21, 04, 44, 788256168, time.UTC), time.Duration(13800000)},
	data.Point{time.Date(2016, 11, 1, 21, 04, 46, 788331398, time.UTC), time.Duration(13700000)},
	data.Point{time.Date(2016, 11, 1, 21, 04, 45, 788291332, time.UTC), time.Duration(13800000)},
	data.Point{time.Date(2016, 11, 1, 21, 04, 45, 788256153, time.UTC), time.Duration(14100000)},
	data.Point{time.Date(2016, 11, 1, 21, 04, 46, 789231777, time.UTC), time.Duration(13600000)},
}

var _ = Describe("Aggregator", func() {
	Describe("NewBuckets", func() {
		It("puts datapoints into interval buckets", func() {
			dataBuckets := aggregator.NewBuckets(dataPoints, time.Second)
			Expect(dataBuckets.Value).To(ConsistOf(
				ConsistOf(
					data.Point{time.Date(2016, 11, 1, 21, 04, 42, 0, time.UTC), time.Duration(28000000)},
					data.Point{time.Date(2016, 11, 1, 21, 04, 42, 760279114, time.UTC), time.Duration(28000000)},
					data.Point{time.Date(2016, 11, 1, 21, 04, 42, 760373651, time.UTC), time.Duration(27900000)},
				),
				ConsistOf(
					data.Point{time.Date(2016, 11, 1, 21, 04, 43, 760213269, time.UTC), time.Duration(28000000)},
					data.Point{time.Date(2016, 11, 1, 21, 04, 43, 760159771, time.UTC), time.Duration(28200000)},
				),
				ConsistOf(
					data.Point{time.Date(2016, 11, 1, 21, 04, 44, 760090065, time.UTC), time.Duration(29100000)},
					data.Point{time.Date(2016, 11, 1, 21, 04, 44, 788256168, time.UTC), time.Duration(13800000)},
				),
				ConsistOf(
					data.Point{time.Date(2016, 11, 1, 21, 04, 45, 788291332, time.UTC), time.Duration(13800000)},
					data.Point{time.Date(2016, 11, 1, 21, 04, 45, 788256153, time.UTC), time.Duration(14100000)},
				),
				ConsistOf(
					data.Point{time.Date(2016, 11, 1, 21, 04, 46, 788331398, time.UTC), time.Duration(13700000)},
					data.Point{time.Date(2016, 11, 1, 21, 04, 46, 789231777, time.UTC), time.Duration(13600000)},
				),
			))
		})

		Context("when no data is passed in", func() {
			It("returns an empty bucket", func() {
				dataBuckets := aggregator.NewBuckets([]data.Point{}, time.Second)
				Expect(dataBuckets.Value).To(BeEmpty())
				Expect(dataBuckets.Summary()).To(BeEmpty())
			})
		})
	})

	Describe("Summary", func() {
		It("returns data that can be graphed in a throughput vs latency plot", func() {
			dataBuckets := aggregator.NewBuckets(dataPoints, time.Second)
			report := dataBuckets.Summary()
			Expect(report).To(ConsistOf(
				aggregator.Point{Throughput: 3, Latency: time.Duration(28000000)},
				aggregator.Point{Throughput: 3, Latency: time.Duration(28000000)},
				aggregator.Point{Throughput: 3, Latency: time.Duration(27900000)},
				aggregator.Point{Throughput: 2, Latency: time.Duration(28000000)},
				aggregator.Point{Throughput: 2, Latency: time.Duration(28200000)},
				aggregator.Point{Throughput: 2, Latency: time.Duration(29100000)},
				aggregator.Point{Throughput: 2, Latency: time.Duration(13800000)},
				aggregator.Point{Throughput: 2, Latency: time.Duration(13800000)},
				aggregator.Point{Throughput: 2, Latency: time.Duration(14100000)},
				aggregator.Point{Throughput: 2, Latency: time.Duration(13700000)},
				aggregator.Point{Throughput: 2, Latency: time.Duration(13600000)},
			))
		})
	})
})
