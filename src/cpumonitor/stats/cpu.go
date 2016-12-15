package stats

import (
	"time"

	CPUUtil "github.com/shirou/gopsutil/cpu"
)

//go:generate counterfeiter -o ../fakes/fake_cpu.go . cpu

type cpu interface {
	Percent(time.Duration, bool) ([]float64, error)
}

type CPUOps struct{}

func (c *CPUOps) Percent(interval time.Duration, perCpu bool) ([]float64, error) {
	return CPUUtil.Percent(interval, perCpu)
}
