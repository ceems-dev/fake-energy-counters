package main

import (
	"github.com/prometheus/procfs"
)

func updateCounters(newStats, oldStats procfs.Stat) {
	elapsedCPUTime := (getTotalCPUTime(newStats.CPUTotal) - getTotalCPUTime(oldStats.CPUTotal)) / updateInterval.Seconds()

	countersMu.Lock()
	currentPowerUsage = elapsedCPUTime * tdp
	currentEnergyuJ += int64(elapsedCPUTime * tdp * float64(updateInterval.Microseconds()))
	countersMu.Unlock()
}

func getTotalCPUTime(s procfs.CPUStat) float64 {
	return s.User + s.System + s.Steal + s.SoftIRQ + s.Nice + s.IRQ + s.Guest + s.GuestNice
}
