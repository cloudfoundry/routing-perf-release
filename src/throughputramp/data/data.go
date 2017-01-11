package data

import (
	"bytes"
	"encoding/csv"
	"errors"
	"strconv"
	"strings"
	"time"
)

type Point struct {
	StartTime    time.Time
	ResponseTime time.Duration
}

type Points []*Point

// Parse will take in an input CSV string and return a slice of data points
func Parse(input string) ([]*Point, error) {
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

func (p *Point) string() string {
	return p.StartTime.Format(time.RFC3339Nano) +
		"," +
		strconv.FormatFloat(p.ResponseTime.Seconds(), 'f', 6, 64)
}

func (p Points) GenerateCSV() []byte {
	buf := bytes.NewBuffer(nil)
	buf.WriteString("start-time,response-time")
	for _, pt := range p {
		buf.WriteByte('\n')
		buf.WriteString(pt.string())
	}
	return buf.Bytes()
}

func fillDataPoints(records [][]string) ([]*Point, error) {
	var dataPoints []*Point
	for _, record := range records[1:] {
		startTime, err := time.Parse(time.RFC3339Nano, record[0])
		if err != nil {
			return nil, err
		}
		responseTime, err := time.ParseDuration(record[1] + "s")
		if err != nil {
			return nil, err
		}
		dataPoints = append(dataPoints, &Point{StartTime: startTime, ResponseTime: responseTime})
	}
	return dataPoints, nil
}
