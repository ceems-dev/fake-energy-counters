package main

import (
	"context"
	"fmt"
	"os"
	"time"
)

func updateIPMIDCMIReading(_ context.Context) error {
	countersMu.RLock()
	powerUsage := currentPowerUsage
	countersMu.RUnlock()

	content := fmt.Sprintf(`Current Power                        : %[1]d Watts
Minimum Power over sampling duration : %[1]d watts
Maximum Power over sampling duration : %[1]d watts
Average Power over sampling duration : %[1]d watts
Time Stamp                           : %s
Statistics reporting time period     : 1473439000 milliseconds
Power Measurement                    : Active
`, int(powerUsage), time.Now())

	err := os.WriteFile(ipmiDcmiFile, []byte(content), 0755)
	if err != nil {
		return fmt.Errorf("failed to write ipmi file %s: %w", ipmiDcmiFile, err)
	}

	return nil
}
