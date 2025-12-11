package slidingwindowcounter_test

import (
	"net/http"
	"testing"
	"time"
)

func TestSWCBurstRequest(t *testing.T) {
	t.Parallel()

	url := "http://localhost:8080/home"

	totalRequests := 10
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
}
