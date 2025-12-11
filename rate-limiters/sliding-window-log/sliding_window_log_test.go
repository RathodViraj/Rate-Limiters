package slidingwindowlog_test

import (
	"net/http"
	"testing"
	"time"
)

// TestSWLRateLimiterBurst sends a burst of requests to a locally running
// server instance with rate-limiter middleware enabled and verifies that
// some requests are limited (429). It targets http://localhost:8080/home by default.
func TestSWLRateLimiterBurst(t *testing.T) {
	t.Parallel()

	url := "http://localhost:8080/home"

	expectedLimit := 6

	totalRequests := 10
	if expectedLimit > 0 {
		totalRequests = expectedLimit + 3
	}

	client := &http.Client{Timeout: 2 * time.Second}

	statuses := make([]int, 0, totalRequests)
	for i := 0; i < totalRequests; i++ {
		req, _ := http.NewRequest(http.MethodGet, url, nil)
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("request %d failed: %v (is the server running on %s?)", i+1, err, url)
		}
		statuses = append(statuses, resp.StatusCode)
		resp.Body.Close()
	}

	t.Logf("burst statuses: %v", statuses)

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
