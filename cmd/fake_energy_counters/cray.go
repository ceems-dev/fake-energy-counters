package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

var (
	crayPMCDir      string
	numAccelerators = 4
	accelPowerUsage = 100
)

func setupCrayPMCDirectories(_ context.Context) error {
	crayPMCDir = filepath.Join(sysfsPath, "cray", "pm_counters")

	// First create pmc directory
	err := os.MkdirAll(crayPMCDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory %s: %w", crayPMCDir, err)
	}

	// Create temp, accel files
	for cpu := range 2 {
		fp := filepath.Join(crayPMCDir, fmt.Sprintf("cpu%d_temp", cpu))

		err := os.WriteFile(fp, []byte("48 C 1734094386460680 us"), 0755)
		if err != nil {
			return fmt.Errorf("failed to write file %s: %w", fp, err)
		}
	}

	for accel := range numAccelerators {
		fp := filepath.Join(crayPMCDir, fmt.Sprintf("accel%d_power", accel))

		err := os.WriteFile(fp, fmt.Appendf(nil, "%d W 1734094386309238 us", accelPowerUsage), 0755)
		if err != nil {
			return fmt.Errorf("failed to write file %s: %w", fp, err)
		}

		fp = filepath.Join(crayPMCDir, fmt.Sprintf("accel%d_power_cap", accel))

		err = os.WriteFile(fp, []byte("0 W"), 0755)
		if err != nil {
			return fmt.Errorf("failed to write file %s: %w", fp, err)
		}
	}

	return nil
}

func updateCrayPMC(_ context.Context) error {
	countersMu.RLock()
	powerUsage := currentPowerUsage
	countersMu.RUnlock()

	dramSplitRatio := randFloat(0.1, 0.2)

	// Assuming 10% power is for other peripherals
	cpuPower := int64(0.9 * powerUsage * (1 - dramSplitRatio))
	dramPower := int64(0.9 * powerUsage * dramSplitRatio)

	// We assumed that accelerators are constantly taking 400 W
	powerUsage += float64(numAccelerators * accelPowerUsage)

	// Write total power file
	fp := filepath.Join(crayPMCDir, "power")

	err := os.WriteFile(fp, fmt.Appendf(nil, "%d W 1734094386309238 us", int64(powerUsage)), 0755)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", fp, err)
	}

	// Write total cpu power file
	fp = filepath.Join(crayPMCDir, "cpu_power")

	err = os.WriteFile(fp, fmt.Appendf(nil, "%d W 1734094386309238 us", int64(cpuPower)), 0755)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", fp, err)
	}

	// Write total power file
	fp = filepath.Join(crayPMCDir, "memory_power")

	err = os.WriteFile(fp, fmt.Appendf(nil, "%d W 1734094386309238 us", int64(dramPower)), 0755)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", fp, err)
	}

	return nil
}
