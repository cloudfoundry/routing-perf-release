package stats

import "time"

type Info struct {
	Percentage []float64
	TimeStamp  time.Time
}

//go:generate counterfeiter -o ../fakes/fake_cpu.go . cpu

type cpu interface {
	Percent(time.Duration, bool) ([]float64, error)
}

type Collector struct {
	interval    time.Duration
	cpuInterval time.Duration
	cpu         cpu
	data        chan []Info
	stop        chan struct{}
}

func NewCollector(c cpu, interval, cpuInterval time.Duration) *Collector {
	return &Collector{
		interval:    interval,
		cpuInterval: cpuInterval,
		cpu:         c,
		data:        make(chan []Info),
		stop:        make(chan struct{}),
	}
}

func (c *Collector) Run() error {
	ticker := time.NewTicker(c.interval)
	var result []Info
	for {
		select {
		case <-ticker.C:
			percentage, err := c.cpu.Percent(c.cpuInterval, false)
			if err != nil {
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

func (c *Collector) Result() []Info {
	c.stop <- struct{}{}
	return <-c.data
}
