package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

//go:embed assets/*
var assetsFS embed.FS

func redfishServer(ctx context.Context) {
	// Registering our handler functions, and creating paths.
	redfishMux := http.NewServeMux()
	redfishMux.HandleFunc("/redfish/v1/", serviceRootHandler)
	redfishMux.HandleFunc("/redfish/v1/Chassis", chassisRootHandler)
	redfishMux.HandleFunc("/redfish/v1/Chassis/{chassisID}", chassisHandler)
	redfishMux.HandleFunc("/redfish/v1/Chassis/{chassisID}/Power", chassisPowerHandler)

	log.Println("Started Redfish on port", redfishPort)

	// Start server
	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", redfishPort),
		ReadHeaderTimeout: 3 * time.Second,
		Handler:           redfishMux,
	}

	defer func() {
		err := server.Shutdown(ctx)
		if err != nil {
			log.Println("Failed to shutdown Redfish server", err)
		}
	}()

	go func() {
		// Spinning up the server.
		err := server.ListenAndServe()
		if err != nil {
			log.Println(err)
		}
	}()

	<-ctx.Done()
}

// serviceRootHandler handles root of redfish API.
func serviceRootHandler(w http.ResponseWriter, r *http.Request) {
	data, err := assetsFS.ReadFile("assets/redfish/service_root.json")
	if err == nil {
		w.Write(data) //nolint:errcheck

		return
	}

	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("KO")) //nolint:errcheck
}

// chassisRootHandler handles chassis collections of redfish API.
func chassisRootHandler(w http.ResponseWriter, r *http.Request) {
	data, err := assetsFS.ReadFile("assets/redfish/chassis_collection.json")
	if err == nil {
		w.Write(data) //nolint:errcheck

		return
	}

	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("KO")) //nolint:errcheck
}

// chassisHandler handles a given chassis of redfish API.
func chassisHandler(w http.ResponseWriter, r *http.Request) {
	chassisID := strings.ReplaceAll(strings.ToLower(r.PathValue("chassisID")), "-", "_")

	data, err := assetsFS.ReadFile(fmt.Sprintf("assets/redfish/%s.json", chassisID))
	if err == nil {
		w.Write(data) //nolint:errcheck

		return
	}

	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("KO")) //nolint:errcheck
}

// chassisPowerHandler handles chassis power of redfish API.
func chassisPowerHandler(w http.ResponseWriter, r *http.Request) {
	chassisID := strings.ReplaceAll(strings.ToLower(r.PathValue("chassisID")), "-", "_")

	data, err := assetsFS.ReadFile(fmt.Sprintf("assets/redfish/%s_power.json", chassisID))
	if err == nil {
		countersMu.RLock()
		powerUsage := currentPowerUsage
		countersMu.RUnlock()

		data = []byte(strings.ReplaceAll(string(data), `"PowerConsumedWatts": 344`, fmt.Sprintf(`"PowerConsumedWatts": %d`, int64(powerUsage))))
		w.Write(data) //nolint:errcheck

		return
	}

	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("KO")) //nolint:errcheck
}
