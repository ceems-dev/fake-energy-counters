package main

import (
	"context"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
)

var (
	powerCapDir string

	raplDomains                = 2
	maxPkgEnergyRangeuJ  int64 = 262143328850
	maxDramEnergyRangeuJ int64 = 65712999613
)

func setupRAPLDirectories(_ context.Context) error {
	powerCapDir = filepath.Join(sysfsClassDir, "powercap")

	// First create powercap directory
	err := os.MkdirAll(powerCapDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create powercap dir: %w", err)
	}

	// Make RAPL domain directories
	for dom := range raplDomains {
		domainDir := filepath.Join(powerCapDir, fmt.Sprintf("intel-rapl:%d", dom))
		err := os.MkdirAll(domainDir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create domain dir %s: %w", domainDir, err)
		}

		// Create necessary files
		// name file
		nameFile := filepath.Join(domainDir, "name")
		err = os.WriteFile(nameFile, fmt.Appendf(nil, "package-%d", dom), 0755)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", nameFile, err)
		}

		// Enabled file
		enabledFile := filepath.Join(domainDir, "enabled")
		err = os.WriteFile(enabledFile, fmt.Appendf(nil, "%d", 1), 0755)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", enabledFile, err)
		}

		// max counter file
		counterFile := filepath.Join(domainDir, "energy_uj")
		err = os.WriteFile(counterFile, fmt.Appendf(nil, "%d", 0), 0755)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", counterFile, err)
		}

		// max counter file
		maxCounterFile := filepath.Join(domainDir, "max_energy_range_uj")
		err = os.WriteFile(maxCounterFile, fmt.Appendf(nil, "%d", maxPkgEnergyRangeuJ), 0755)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", maxCounterFile, err)
		}

		// power limit file
		powerLimitFile := filepath.Join(domainDir, "constraint_0_power_limit_uw")
		err = os.WriteFile(powerLimitFile, fmt.Appendf(nil, "%d", int64(tdp*1e6)), 0755)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", powerLimitFile, err)
		}

		// Each domain will have a sub domain which is dram
		subDomainDir := filepath.Join(domainDir, fmt.Sprintf("intel-rapl:%d:0", dom))
		err = os.MkdirAll(subDomainDir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create sub domain dir %s: %w", subDomainDir, err)
		}

		// Need to create symlinks to root
		subDomainSymLink := filepath.Join(powerCapDir, fmt.Sprintf("intel-rapl:%d:0", dom))
		err = os.Symlink(subDomainDir, subDomainSymLink)
		if err != nil {
			return fmt.Errorf("failed to create symlink %s: %w", subDomainSymLink, err)
		}

		// Create necessary files
		// name file
		nameFile = filepath.Join(subDomainDir, "name")
		err = os.WriteFile(nameFile, []byte("dram"), 0755)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", nameFile, err)
		}

		// Enabled file
		enabledFile = filepath.Join(subDomainDir, "enabled")
		err = os.WriteFile(enabledFile, fmt.Appendf(nil, "%d", 1), 0755)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", enabledFile, err)
		}

		// counter file
		counterFile = filepath.Join(subDomainDir, "energy_uj")
		err = os.WriteFile(counterFile, fmt.Appendf(nil, "%d", 0), 0755)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", counterFile, err)
		}

		// max counter file
		maxCounterFile = filepath.Join(subDomainDir, "max_energy_range_uj")
		err = os.WriteFile(maxCounterFile, fmt.Appendf(nil, "%d", maxDramEnergyRangeuJ), 0755)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", maxCounterFile, err)
		}

		// power limit file
		powerLimitFile = filepath.Join(subDomainDir, "constraint_0_power_limit_uw")
		err = os.WriteFile(powerLimitFile, fmt.Appendf(nil, "%d", int64(tdp*1e5)), 0755)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", powerLimitFile, err)
		}
	}

	return nil
}

func updateRAPLCounters(_ context.Context) error {
	pkgSplitRatio := rand.Float64()

	countersMu.RLock()
	energyuJ := currentEnergyuJ
	countersMu.RUnlock()

	pkgCounters := []int64{int64(pkgSplitRatio * float64(energyuJ)), int64((1 - pkgSplitRatio) * float64(energyuJ))}
	for dom, counter := range pkgCounters {
		subDomainSplitRatio := randFloat(0.1, 0.2)

		domainDir := filepath.Join(powerCapDir, fmt.Sprintf("intel-rapl:%d", dom))
		subDomainDir := filepath.Join(domainDir, fmt.Sprintf("intel-rapl:%d:0", dom))

		// domain counter file
		counter := int64(float64(counter) * subDomainSplitRatio)
		if counter > maxPkgEnergyRangeuJ {
			counter = counter - maxPkgEnergyRangeuJ
		}
		counterFile := filepath.Join(domainDir, "energy_uj")
		err := os.WriteFile(counterFile, fmt.Appendf(nil, "%d", counter), 0755)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", counterFile, err)
		}

		// sub domain counter file
		counter = int64(float64(counter) * (1 - subDomainSplitRatio))
		if counter > maxDramEnergyRangeuJ {
			counter = counter - maxDramEnergyRangeuJ
		}
		counterFile = filepath.Join(subDomainDir, "energy_uj")
		err = os.WriteFile(counterFile, fmt.Appendf(nil, "%d", counter), 0755)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", counterFile, err)
		}
	}

	return nil
}

func randFloat(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}
