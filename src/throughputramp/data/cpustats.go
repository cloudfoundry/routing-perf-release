package data

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type CPUStat struct {
	TimeStamp  time.Time `json:"TimeStamp"`
	Percentage []float64 `json:"Percentage"`
}

func (i *CPUStat) percentageString() string {
	var results []string
	for _, d := range i.Percentage {
		results = append(results, strconv.FormatFloat(d, 'f', 6, 64))
	}
	return strings.Join(results, ",")
}

func (i *CPUStat) string() string {
	return fmt.Sprintf("%s,%v", i.TimeStamp, i.percentageString())
}

func GenerateCpuCSV(body []byte) []byte {
	if body == nil || len(body) == 0 {
		return nil
	}

	var results []CPUStat
	err := json.Unmarshal(body, &results)
	if err != nil {
		fmt.Println(err.Error())
	}
	buf := bytes.NewBuffer(nil)

	buf.WriteString("timeStamp" + strings.Repeat(",percentage", len(results[0].Percentage)))
	for _, p := range results {
		buf.WriteByte('\n')
		buf.WriteString(p.string())
	}
	return buf.Bytes()
}
