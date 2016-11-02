package data

import (
	"encoding/csv"
	"errors"
	"sort"
	"strings"
	"time"
)

type Point struct {
	StartTime    time.Time
	ResponseTime time.Duration
}

type pointsSortableByStartTime []Point
type pointsSortableByResponseTime []Point

// Parse will take in an input CSV string and return a slice of data points
func Parse(input string) ([]Point, error) {
	r := csv.NewReader(strings.NewReader(input))
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	header := records[0]
	if header[0] != "start-time" && header[1] != "response-time" {
		return nil, errors.New("csv headers not found")
	}

	return fillDataPoints(records)
}

func fillDataPoints(records [][]string) ([]Point, error) {
	var dataPoints []Point
	for _, record := range records[1:] {
		startTime, err := time.Parse(time.RFC3339Nano, record[0])
		if err != nil {
			return nil, err
		}
		responseTime, err := time.ParseDuration(record[1] + "s")
		if err != nil {
			return nil, err
		}
		dataPoints = append(dataPoints, Point{StartTime: startTime, ResponseTime: responseTime})
	}
	return dataPoints, nil
}

// SortByStartTime will sort the records in place by StartTime
func SortByStartTime(records []Point) {
	sort.Sort(pointsSortableByStartTime(records))
}

// Len is the number of elements in the collection.
func (p pointsSortableByStartTime) Len() int {
	return len(p)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (p pointsSortableByStartTime) Less(i, j int) bool {
	return p[j].StartTime.After(p[i].StartTime)
}

// Swap swaps the elements with indexes i and j.
func (p pointsSortableByStartTime) Swap(i, j int) {
	t := p[i]
	p[i] = p[j]
	p[j] = t
}

// SortByResponseTime will sort the records in place by ResponseTime
func SortByResponseTime(records []Point) {
	sort.Sort(pointsSortableByResponseTime(records))
}

// Len is the number of elements in the collection.
func (p pointsSortableByResponseTime) Len() int {
	return len(p)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (p pointsSortableByResponseTime) Less(i, j int) bool {
	return p[i].ResponseTime < p[j].ResponseTime
}

// Swap swaps the elements with indexes i and j.
func (p pointsSortableByResponseTime) Swap(i, j int) {
	t := p[i]
	p[i] = p[j]
	p[j] = t
}
