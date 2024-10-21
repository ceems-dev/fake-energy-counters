package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/mahendrapaipuri/fake-energy-counters/pkg/helpers"
	"github.com/prometheus/procfs"
)

var (
	fakeEnergyCounters = kingpin.New("fake-energy-counters", "A command-line application to provide fake IPMI DCMI power readings for testing.")
	tdp                = fakeEnergyCounters.Flag("power.tdp", "Reference TDP to use for estimating energy counters").Default("180").Float64()
	procfsPath         = fakeEnergyCounters.Flag("path.procfs", "procfs mountpoint.").Default("/proc").String()
)

type Reading struct {
	Inst, Min, Max, Avg float64
}

func DCMIReading(procfsPath string, tdp float64) (Reading, error) {
	fs, err := procfs.NewFS(procfsPath)
	if err != nil {
		return Reading{}, fmt.Errorf("failed to open procfs: %w", err)
	}

	cpuStats, err := fs.Stat()
	if err != nil {
		return Reading{}, err
	}

	time.Sleep(500 * time.Millisecond)

	newStats, err := fs.Stat()
	if err != nil {
		return Reading{}, err
	}

	powerReading := helpers.GetPowerReading(tdp, cpuStats, newStats, 500*time.Millisecond)

	return Reading{Inst: powerReading}, nil
}

func PrintReading() {
	// Update reading
	var reading Reading
	var err error
	if reading, err = DCMIReading(*procfsPath, *tdp); err != nil {
		log.Fatalf("failed to get DCMI power readings: %s", err)
	}

	fmt.Printf(`Current Power                        : %d Watts
Minimum Power over sampling duration : %d watts
Maximum Power over sampling duration : %d watts
Average Power over sampling duration : %d watts
Time Stamp                           : %s
Statistics reporting time period     : 1473439000 milliseconds
Power Measurement                    : Active`, int(reading.Inst), int(reading.Min), int(reading.Max), int(reading.Avg), time.Now())
	fmt.Printf("\n")
}

// Main entry point for `fake_energy_counters` app.
func main() {
	// Parse CLI opts.
	_, err := fakeEnergyCounters.Parse(os.Args[1:])
	if err != nil {
		log.Fatalf("failed to parse CLI args: %s", err)
	}

	PrintReading()
}
