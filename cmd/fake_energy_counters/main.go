package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/alecthomas/kingpin/v2"
)

var (
	tdp                              float64
	updateInterval                   time.Duration
	procfsPath                       string
	sysfsPath                        string
	redfishPort                      int
	ipmiDcmi, rapl, crayPMC, redfish bool
)

var (
	fakeEnergyCounters = kingpin.New("fake-energy-counters", "A command-line application to provide fake IPMI DCMI power readings for testing.")
)

var (
	sysfsClassDir     string
	ipmiDcmiFile      string
	currentPowerUsage float64
	currentEnergyuJ   int64
	countersMu        sync.RWMutex
)

// Main entry point for `fake_energy_counters` app.
func main() {
	fakeEnergyCounters.Flag("power.tdp", "Reference TDP to use for estimating energy counters").Default("180").Float64Var(&tdp)
	fakeEnergyCounters.Flag("power.update-interval", "Interval to update the power and energy counters").Default("5s").DurationVar(&updateInterval)
	fakeEnergyCounters.Flag("path.procfs", "procfs mountpoint.").Hidden().Default("/proc").StringVar(&procfsPath)
	fakeEnergyCounters.Flag("path.sysfs", "sysfs mountpoint.").Default("/opt/sys").StringVar(&sysfsPath)
	fakeEnergyCounters.Flag("counters.ipmi", "Enable IPMI DCMI command.").Default("false").BoolVar(&ipmiDcmi)
	fakeEnergyCounters.Flag("counters.rapl", "Enable RAPL counters.").Default("false").BoolVar(&rapl)
	fakeEnergyCounters.Flag("counters.cray", "Enable Cray PM counters.").Default("false").BoolVar(&crayPMC)
	fakeEnergyCounters.Flag("counters.redfish", "Enable Redfish server.").Default("false").BoolVar(&redfish)
	fakeEnergyCounters.Flag("counters.redfish.port", "Redfish port number.").Default("5000").IntVar(&redfishPort)

	// Parse CLI opts.
	_, err := fakeEnergyCounters.Parse(os.Args[1:])
	if err != nil {
		log.Fatalf("failed to parse CLI args: %s", err)
	}

	// Get sysfs class director
	sysfsClassDir = filepath.Join(sysfsPath, "class")

	// Make sysfs directory
	err = os.MkdirAll(sysfsClassDir, 0755)
	if err != nil {
		log.Fatalf("failed to make sysfs directory: %s", err)
	}

	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if rapl {
		err := setupRAPLDirectories(ctx)
		if err != nil {
			log.Fatalf("failed to setup rapl directories: %s", err)
		}

		log.Println("RAPL counters will be emulated at", powerCapDir)
	}

	if crayPMC {
		err := setupCrayPMCDirectories(ctx)
		if err != nil {
			log.Fatalf("failed to setup cray pmc directories: %s", err)
		}

		log.Println("Cray PM counters will be emulated at", crayPMCDir)
	}

	if ipmiDcmi {
		ipmiDcmiFile = filepath.Join(sysfsPath, "ipmi")

		log.Println("IPMI DCMI output will be emulated at file", ipmiDcmiFile)
	}

	wg := sync.WaitGroup{}

	wg.Add(1)

	// Setup power and energy counters
	go func() {
		defer wg.Done()

		log.Println("Starting energy counters")

		err = startEnergyCounters(ctx)
		if err != nil {
			log.Fatalf("failed to setup energy counters: %s", err)
		}
	}()

	wg.Add(1)

	// Setup power and energy counters
	go func() {
		defer wg.Done()

		log.Println("Starting redfish emulator")

		redfishServer(ctx)
	}()

	// Listen for the interrupt signal.
	<-ctx.Done()

	wg.Wait()

	// Restore default behavior on the interrupt signal and notify user of shutdown.
	stop()
	log.Println("Shutting down gracefully, press Ctrl+C again to force")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Println("Emulator exiting")
	log.Println("See you next time!!")
}
