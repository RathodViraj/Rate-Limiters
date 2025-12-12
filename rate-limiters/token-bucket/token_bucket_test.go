package tokenbucket_test

import (
	"net/http"
	"testing"
	"time"
)

func TestTBBurst(t *testing.T) {
	t.Parallel()

	homeURL := "http://localhost:8080/home"
	freeURL := "http://localhost:8080/free"

	totalRequests := 12

	client := &http.Client{Timeout: 20 * time.Second}

	homeStatuses := make([]int, 0, totalRequests)
	homeLatencies := make([]time.Duration, 0, totalRequests)
	for i := 0; i < totalRequests; i++ {
		req, _ := http.NewRequest(http.MethodGet, homeURL, nil)
		start := time.Now()
		resp, err := client.Do(req)
		latency := time.Since(start)
		if err != nil {
			t.Fatalf("request %d failed: %v (is the server running on %s?)", i+1, err, homeURL)
		}
		homeStatuses = append(homeStatuses, resp.StatusCode)
		homeLatencies = append(homeLatencies, latency)
		resp.Body.Close()

		if i == 6 {
			time.Sleep(time.Second * 5) // wait for refile
		}
	}

	freeLatencies := make([]time.Duration, 0, totalRequests)
	for i := 0; i < totalRequests; i++ {
		req, _ := http.NewRequest(http.MethodGet, freeURL, nil)
		start := time.Now()
		resp, err := client.Do(req)
		latency := time.Since(start)
		if err != nil {
			t.Fatalf("free request %d failed: %v", i+1, err)
		}
		freeLatencies = append(freeLatencies, latency)
		resp.Body.Close()

		if i == 6 {
			time.Sleep(time.Second * 5)
		}
	}

	// Calculate average latencies
	var homeTotal, freeTotal time.Duration
	for i := 0; i < totalRequests; i++ {
		homeTotal += homeLatencies[i]
		freeTotal += freeLatencies[i]
	}
	homeAvg := homeTotal / time.Duration(totalRequests)
	freeAvg := freeTotal / time.Duration(totalRequests)
	latencyDiff := homeAvg - freeAvg

	t.Logf("burst statuses: %v", homeStatuses)
	t.Logf("Average latency for /home: %v", homeAvg)
	t.Logf("Average latency for /free: %v", freeAvg)
	t.Logf("Average latency difference (overhead): %v", latencyDiff)
}
