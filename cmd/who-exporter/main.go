package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Gauge for the total number of unique logged-in users
	uniqueUsers = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "system_unique_logged_in_users",
			Help: "Number of unique logged-in users on the system",
		},
	)

	// GaugeVec for per-user session counts with IP addresses as labels
	userSessions = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "system_user_sessions",
			Help: "Number of active sessions per user with IP addresses",
		},
		[]string{"username", "ip"},
	)
)

func init() {
	// Register metrics with Prometheus
	prometheus.MustRegister(uniqueUsers, userSessions)
}

func updateLoggedInUsers() {
	// Run the 'who' command
	cmd := exec.Command("who")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Error running 'who' command: %v", err)
		return
	}

	// Parse output
	lines := strings.Split(string(output), "\n")
	userMap := make(map[string]map[string]int) // username -> IP -> session count
	uniqueUserSet := make(map[string]struct{}) // unique usernames

	for _, line := range lines {
		if line == "" {
			continue
		}

		// Split line into fields (username, pts/X, date, time, IP)
		fields := strings.Fields(line)
		if len(fields) < 5 {
			log.Printf("Skipping malformed line: %s", line)
			continue
		}

		username := fields[0]
		ip := strings.Trim(fields[4], "()") // Remove parentheses from IP

		// Track unique users
		uniqueUserSet[username] = struct{}{}

		// Initialize IP map for user if not exists
		if _, exists := userMap[username]; !exists {
			userMap[username] = make(map[string]int)
		}

		// Increment session count for this user-IP pair
		userMap[username][ip]++
	}

	// Update metrics
	uniqueUsers.Set(float64(len(uniqueUserSet)))

	// Reset previous user session metrics
	userSessions.Reset()

	// Set new session counts
	for username, ipMap := range userMap {
		for ip, count := range ipMap {
			userSessions.WithLabelValues(username, ip).Set(float64(count))
		}
	}
}

func main() {
	// Define command-line flags
	host := flag.String("host", "localhost", "Host address to listen on (e.g., 0.0.0.0 or localhost)")
	port := flag.Int("port", 9101, "Port to listen on")
	flag.Parse()

	// Construct listening address
	addr := fmt.Sprintf("%s:%d", *host, *port)

	// Update metrics periodically
	go func() {
		for {
			updateLoggedInUsers()
			time.Sleep(10 * time.Second)
		}
	}()

	// Expose /metrics endpoint
	http.Handle("/metrics", promhttp.Handler())
	log.Printf("Starting who-exporter on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
