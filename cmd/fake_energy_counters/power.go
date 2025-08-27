package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/prometheus/procfs"
)

func startEnergyCounters(ctx context.Context) error {
	fs, err := procfs.NewFS(procfsPath)
	if err != nil {
		return fmt.Errorf("failed to open procfs: %w", err)
	}

	// Initialize tickers
	energyTicker := time.NewTicker(updateInterval)

	oldStats, err := fs.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat procfs: %w", err)
	}

	for {
		newStats, err := fs.Stat()
		if err != nil {
			return fmt.Errorf("failed to stat procfs: %w", err)
		}

		// Update counters
		updateCounters(newStats, oldStats)

		// Update RAPL counters
		if rapl {
			err := updateRAPLCounters(ctx)
			if err != nil {
				log.Printf("failed to update rapl counters: %s\n", err)
			}
		}

		// Update Cray PM counters
		if crayPMC {
			err := updateCrayPMC(ctx)
			if err != nil {
				log.Printf("failed to update cray pm counters: %s\n", err)
			}
		}

		// Update IPMI DCMI file
		if ipmiDcmi {
			err := updateIPMIDCMIReading(ctx)
			if err != nil {
				log.Printf("failed to update ipmi dcmi counters: %s\n", err)
			}
		}

		oldStats = newStats

		select {
		case <-energyTicker.C:
			continue
		case <-ctx.Done():
			log.Println("Received Interrupt. Stopping energy update")

			return nil
		}
	}
}
