package slidingwindowlog_test

import (
	"net/http"
	"testing"
	"time"
)

func TestSWLRateLimiterBurst(t *testing.T) {
	t.Parallel()

	homeURL := "http://localhost:8080/home"
	freeURL := "http://localhost:8080/free"

	expectedLimit := 6

	totalRequests := 10
	if expectedLimit > 0 {
		totalRequests = expectedLimit + 3
	}

	client := &http.Client{Timeout: 2 * time.Second}

	statuses := make([]int, 0, totalRequests)
	homeLatencies := make([]time.Duration, 0, totalRequests)
	for i := 0; i < totalRequests; i++ {
		req, _ := http.NewRequest(http.MethodGet, homeURL, nil)
		start := time.Now()
		resp, err := client.Do(req)
		latency := time.Since(start)
		if err != nil {
			t.Fatalf("request %d failed: %v (is the server running on %s?)", i+1, err, homeURL)
		}
		statuses = append(statuses, resp.StatusCode)
		homeLatencies = append(homeLatencies, latency)
		resp.Body.Close()
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
	}

	var homeTotal, freeTotal time.Duration
	for i := 0; i < totalRequests; i++ {
		homeTotal += homeLatencies[i]
		freeTotal += freeLatencies[i]
	}
	homeAvg := homeTotal / time.Duration(totalRequests)
	freeAvg := freeTotal / time.Duration(totalRequests)
	latencyDiff := homeAvg - freeAvg

	t.Logf("burst statuses: %v", statuses)
	t.Logf("Average latency for /home: %v", homeAvg)
	t.Logf("Average latency for /free: %v", freeAvg)
	t.Logf("Average latency difference (overhead): %v", latencyDiff)

	limited := count(statuses, http.StatusTooManyRequests)
	ok := count(statuses, http.StatusOK)
	if limited == 0 {
		t.Fatalf("expected at least one 429 response, got none (statuses=%v)", statuses)
	}
	if ok == 0 {
		t.Fatalf("expected at least one 200 response, got none (statuses=%v)", statuses)
	}

	if expectedLimit > 0 {
		if ok > expectedLimit {
			t.Fatalf("successes exceeded expected limit: got %d > limit %d (statuses=%v)", ok, expectedLimit, statuses)
		}
	}
}

func count(arr []int, code int) int {
	c := 0
	for _, v := range arr {
		if v == code {
			c++
		}
	}
	return c
}
