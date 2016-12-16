package aggregator

import (
	"bytes"
	"strconv"
	"throughputramp/data"
	"time"
)

type Aggregator struct {
	buckets  [][]*data.Point
	interval time.Duration
}

type Point struct {
	Throughput float64
	Latency    time.Duration
}

func (p *Point) string() string {
	return strconv.FormatFloat(p.Throughput, 'f', -1, 64) + "," + strconv.FormatFloat(p.Latency.Seconds(), 'f', 6, 64)
}

type Report []Point

func (r Report) GenerateCSV() []byte {
	buf := bytes.NewBuffer(nil)
	buf.WriteString("throughput,latency")
	for _, p := range r {
		buf.WriteByte('\n')
		buf.WriteString(p.string())
	}
	return buf.Bytes()
}

func New(dataPoints []*data.Point, interval time.Duration) *Aggregator {
	if len(dataPoints) == 0 {
		return &Aggregator{}
	}

	data.Sort(dataPoints)
	buckets := [][]*data.Point{}

	startTime := dataPoints[0].StartTime
	currentBucketTime := startTime
	nextBucketTime := startTime.Add(interval)
	bucketIndex := 0
	buckets = append(buckets, nil)

	for _, dp := range dataPoints {
		for dp.StartTime.After(nextBucketTime) {
			currentBucketTime = nextBucketTime
			bucketIndex += 1
			buckets = append(buckets, nil)
			nextBucketTime = currentBucketTime.Add(interval)
		}
		buckets[bucketIndex] = append(buckets[bucketIndex], dp)
	}
	return &Aggregator{buckets: buckets, interval: interval}
}

func (a *Aggregator) Data() Report {
	var report Report

	for _, points := range a.buckets {
		for _, dataPoint := range points {
			report = append(report, Point{
				Throughput: float64(len(points)) / a.interval.Seconds(),
				Latency:    dataPoint.ResponseTime,
			})
		}
	}
	return report
}
