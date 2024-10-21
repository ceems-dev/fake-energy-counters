package helpers

import (
	"time"

	"github.com/prometheus/procfs"
)

func GetPowerReading(tdp float64, oldStats, newStats procfs.Stat, interval time.Duration) float64 {
	elapsedCPUTime := (getTotalCPUTime(newStats.CPUTotal) - getTotalCPUTime(oldStats.CPUTotal)) / interval.Seconds()

	return elapsedCPUTime * tdp
}

func getTotalCPUTime(s procfs.CPUStat) float64 {
	return s.User + s.System + s.Steal + s.SoftIRQ + s.Nice + s.IRQ + s.Guest + s.GuestNice
}
