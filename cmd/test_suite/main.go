package main

import (
	"fmt"
	"net/http"
	"io/ioutil"
	"strings"
	"time"
)

func main() {
	baseURL := "http://localhost:8080/api/v1"
	headers := map[string]string{
		"X-App-Key": "netengine_secret_key_123",
		"Content-Type": "application/json",
	}

	tests := []struct {
		Name   string
		Method string
		URL    string
		Body   string
	}{
		{"1. Health Check", "GET", "/router/1/health", ""},
		{"2. Get Users", "GET", "/router/1/users", ""},
		{"3. Isolate User (Sync)", "POST", "/isolate?sync=true", `{"ip": "10.2.3.164", "action": "add", "router_id": 1, "list": "ISOLATED", "comment": "Test Suite"}`},
		{"4. Unisolate User (Async)", "POST", "/isolate", `{"ip": "10.2.3.164", "action": "remove", "router_id": 1, "list": "ISOLATED"}`},
		{"5. Force Sync Secrets", "POST", "/sync/1", ""},
	}

	fmt.Println("=== Starting API Test Suite for Randuagung (Router 1) ===")
	client := &http.Client{Timeout: 35 * time.Second}

	for _, t := range tests {
		fmt.Printf("\nRunning: %s...\n", t.Name)
		req, _ := http.NewRequest(t.Method, baseURL+t.URL, strings.NewReader(t.Body))
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("FAILED: %v\n", err)
			continue
		}
		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Printf("Status: %s\n", resp.Status)
		if len(body) > 300 {
			fmt.Printf("Response: %s... (truncated)\n", string(body)[:300])
		} else {
			fmt.Printf("Response: %s\n", string(body))
		}
	}
}
