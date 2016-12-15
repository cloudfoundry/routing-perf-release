package stats

import (
	"log"
	"time"
)

//go:generate counterfeiter -o ../fakes/fake_cpucollector.go . Collector

type Collector interface {
	Run() error
	Result() []Info
}

type Ticker interface {
	Stop()
}

type TickerHarness func(d time.Duration) <-chan time.Time

func DefaultTicker() TickerHarness {
	return func(d time.Duration) <-chan time.Time {
		t := time.NewTicker(d)
		return t.C
	}
}

type Info struct {
	Percentage []float64 `json:"Percentage"`
	TimeStamp  time.Time `json:"TimeStamp"`
}

type CPUCollector struct {
	tickChan    <-chan time.Time
	cpuInterval time.Duration
	perCPU      bool
	cpu         cpu
	data        chan []Info
	stop        chan struct{}
}

func NewCPUCollector(cpu cpu, ticker TickerHarness, runInterval, cpuInterval time.Duration, perCpu bool) Collector {
	c := ticker(runInterval)
	return &CPUCollector{
		tickChan:    c,
		cpuInterval: cpuInterval,
		perCPU:      perCpu,
		cpu:         cpu,
		data:        make(chan []Info),
		stop:        make(chan struct{}),
	}
}

func (c *CPUCollector) Run() error {
	var result []Info
	for {
		select {
		case <-c.tickChan:
			percentage, err := c.cpu.Percent(c.cpuInterval, c.perCPU)
			if err != nil {
				// TODO: we might not want to return if there is an error
				log.Printf("Failed to get CPU Percentage %v", err)
				return err
			}
			result = append(result, Info{
				Percentage: percentage,
				TimeStamp:  time.Now(),
			})
		case <-c.stop:
			c.data <- result
			return nil
		}
	}
	return nil
}

func (c *CPUCollector) Result() []Info {
	c.stop <- struct{}{}
	return <-c.data
}
