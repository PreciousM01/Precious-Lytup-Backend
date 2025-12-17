package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"sync"
	"testing"
)



func TestDHIConcurrentRequests(t *testing.T) {
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	payload := map[string]any{
		"SrID": "sr05",
		"Seed": map[string]any{"x": 1},
	}

	body, _ := json.Marshal(payload)

	const N = 100
	var wg sync.WaitGroup
	wg.Add(N)

	for i := 0; i < N; i++ {
		go func(i int) {
			defer wg.Done()

			resp, err := client.Post(
				"https://localhost:8443",
				"application/json",
				bytes.NewBuffer(body),
			)
			if err != nil {
				t.Errorf("request %d failed: %v", i, err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("request %d returned %d", i, resp.StatusCode)
			}
		}(i)
	}

	wg.Wait()
}
