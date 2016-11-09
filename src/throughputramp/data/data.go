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

type points []Point

// Parse will take in an input CSV string and return a slice of data points
func Parse(input string) ([]Point, error) {
	if input == "" {
		return nil, errors.New("empty input")
	}
	r := csv.NewReader(strings.NewReader(input))
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	header := records[0]
	if len(header) < 2 || (header[0] != "start-time" && header[1] != "response-time") {
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

// Sort will sort the records in place by StartTime
func Sort(records []Point) {
	sort.Sort(points(records))
}

// Len is the number of elements in the collection.
func (p points) Len() int {
	return len(p)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (p points) Less(i, j int) bool {
	return p[j].StartTime.After(p[i].StartTime)
}

// Swap swaps the elements with indexes i and j.
func (p points) Swap(i, j int) {
	t := p[i]
	p[i] = p[j]
	p[j] = t
}
