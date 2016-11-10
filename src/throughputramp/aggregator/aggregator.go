package aggregator

import (
	"bytes"
	"strconv"
	"throughputramp/data"
	"time"
)

type Buckets struct {
	Value    map[time.Time][]*data.Point
	interval time.Duration
}

type Point struct {
	Throughput float64
	Latency    time.Duration
}

func (p *Point) String() string {
	return strconv.FormatFloat(p.Throughput, 'f', -1, 64) + "," + strconv.FormatFloat(p.Latency.Seconds(), 'f', 6, 64)
}

type Report []Point

func (r Report) GenerateCSV() []byte {
	buf := bytes.NewBuffer(nil)
	buf.WriteString("throughput,latency")
	for _, p := range r {
		buf.WriteByte('\n')
		buf.WriteString(p.String())
	}
	return buf.Bytes()
}

func NewBuckets(dataPoints []*data.Point, interval time.Duration) *Buckets {
	if len(dataPoints) == 0 {
		return &Buckets{}
	}

	data.Sort(dataPoints)
	dataBuckets := make(map[time.Time][]*data.Point)

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
