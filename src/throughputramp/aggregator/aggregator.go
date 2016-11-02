package aggregator

import (
	"encoding/json"
	"throughputramp/data"
	"time"
)

type Buckets struct {
	Value    map[time.Time][]data.Point
	interval time.Duration
}

type Point struct {
	Throughput float64
	Latency    time.Duration
}

var _ json.Marshaler = Point{}

func (p Point) MarshalJSON() ([]byte, error) {
	return json.Marshal([]float64{p.Throughput, p.Latency.Seconds()})
}

type Report []Point

func NewBuckets(dataPoints []data.Point, interval time.Duration) *Buckets {
	data.SortByStartTime(dataPoints)
	dataBuckets := make(map[time.Time][]data.Point)

	startTime := dataPoints[0].StartTime
	currentBucketTime := startTime
	nextBucketTime := startTime.Add(interval)
	for _, dp := range dataPoints {
		for dp.StartTime.After(nextBucketTime) {
			currentBucketTime = nextBucketTime
			nextBucketTime = currentBucketTime.Add(interval)
		}
		dataBuckets[currentBucketTime] = append(dataBuckets[currentBucketTime], dp)

	}
	return &Buckets{Value: dataBuckets, interval: interval}
}

func (b Buckets) Summary() Report {
	var report Report
	for _, points := range b.Value {
		for _, dataPoint := range points {
			report = append(report, Point{
				Throughput: float64(len(points)) / b.interval.Seconds(),
				Latency:    dataPoint.ResponseTime,
			})
		}
	}
	return report
}
