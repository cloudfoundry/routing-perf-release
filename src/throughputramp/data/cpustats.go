package data

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type cpuStat struct {
	TimeStamp  time.Time `json:"TimeStamp"`
	Percentage []float64 `json:"Percentage"`
}

func (i *cpuStat) percentageString() string {
	var results []string
	for _, d := range i.Percentage {
		results = append(results, strconv.FormatFloat(d, 'f', 6, 64))
	}
	return strings.Join(results, ",")
}

func (i *cpuStat) string() string {
	return fmt.Sprintf("%s,%v", i.TimeStamp, i.percentageString())
}

func GenerateCpuCSV(body []byte) ([]byte, error) {
	if body == nil || len(body) == 0 {
		return nil, errors.New("empty/nil body")
	}

	var results []cpuStat
	err := json.Unmarshal(body, &results)
	if err != nil {
		return nil, fmt.Errorf("marshaling data: %s", err)
	}
	buf := bytes.NewBuffer(nil)

	buf.WriteString("timeStamp" + strings.Repeat(",percentage", len(results[0].Percentage)))
	for _, p := range results {
		buf.WriteByte('\n')
		buf.WriteString(p.string())
	}
	return buf.Bytes(), nil
}
