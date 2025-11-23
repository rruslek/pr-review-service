package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

type Stats struct {
	TotalRequests   int
	SuccessRequests int
	ErrorRequests   int
	MinLatency      time.Duration
	MaxLatency      time.Duration
	AvgLatency      time.Duration
	P50Latency      time.Duration
	P95Latency      time.Duration
	P99Latency      time.Duration
	Latencies       []time.Duration
	mu              sync.Mutex
}

func NewStats() *Stats {
	return &Stats{
		MinLatency: time.Hour,
		Latencies:  make([]time.Duration, 0),
	}
}

func (s *Stats) Add(latency time.Duration, success bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.TotalRequests++
	if success {
		s.SuccessRequests++
	} else {
		s.ErrorRequests++
	}

	s.Latencies = append(s.Latencies, latency)

	if latency < s.MinLatency {
		s.MinLatency = latency
	}
	if latency > s.MaxLatency {
		s.MaxLatency = latency
	}
}

func (s *Stats) Calculate() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.Latencies) == 0 {
		return
	}

	for i := 0; i < len(s.Latencies); i++ {
		for j := i + 1; j < len(s.Latencies); j++ {
			if s.Latencies[i] > s.Latencies[j] {
				s.Latencies[i], s.Latencies[j] = s.Latencies[j], s.Latencies[i]
			}
		}
	}

	var total time.Duration
	for _, lat := range s.Latencies {
		total += lat
	}
	s.AvgLatency = total / time.Duration(len(s.Latencies))

	if len(s.Latencies) > 0 {
		s.P50Latency = s.Latencies[len(s.Latencies)*50/100]
		s.P95Latency = s.Latencies[len(s.Latencies)*95/100]
		s.P99Latency = s.Latencies[len(s.Latencies)*99/100]
	}
}

func (s *Stats) Print(name string) {
	s.Calculate()
	fmt.Printf("\n=== %s ===\n", name)
	fmt.Printf("Total Requests: %d\n", s.TotalRequests)
	fmt.Printf("Success: %d (%.2f%%)\n", s.SuccessRequests, float64(s.SuccessRequests)/float64(s.TotalRequests)*100)
	fmt.Printf("Errors: %d (%.2f%%)\n", s.ErrorRequests, float64(s.ErrorRequests)/float64(s.TotalRequests)*100)
	fmt.Printf("Min Latency: %v\n", s.MinLatency)
	fmt.Printf("Max Latency: %v\n", s.MaxLatency)
	fmt.Printf("Avg Latency: %v\n", s.AvgLatency)
	fmt.Printf("P50 Latency: %v\n", s.P50Latency)
	fmt.Printf("P95 Latency: %v\n", s.P95Latency)
	fmt.Printf("P99 Latency: %v\n", s.P99Latency)
}

func makeRequest(client *http.Client, method, url string, body []byte) (time.Duration, bool) {
	start := time.Now()

	var req *http.Request
	var err error

	if body != nil {
		req, err = http.NewRequest(method, url, bytes.NewBuffer(body))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		return time.Since(start), false
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return time.Since(start), false
	}
	defer resp.Body.Close()

	io.Copy(io.Discard, resp.Body)

	latency := time.Since(start)
	success := resp.StatusCode >= 200 && resp.StatusCode < 300

	return latency, success
}

type BodyGenerator func(requestID int) []byte

func runLoadTest(client *http.Client, method, url string, bodyGenerator BodyGenerator, concurrency, requests int) *Stats {
	stats := NewStats()
	var wg sync.WaitGroup
	var requestCounter int
	var counterMu sync.Mutex

	getNextID := func() int {
		counterMu.Lock()
		defer counterMu.Unlock()
		id := requestCounter
		requestCounter++
		return id
	}

	requestsPerGoroutine := requests / concurrency
	remaining := requests % concurrency

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(requestsCount int) {
			defer wg.Done()
			for j := 0; j < requestsCount; j++ {
				var body []byte
				if bodyGenerator != nil {
					body = bodyGenerator(getNextID())
				}
				latency, success := makeRequest(client, method, url, body)
				stats.Add(latency, success)
			}
		}(requestsPerGoroutine)
	}

	if remaining > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < remaining; j++ {
				var body []byte
				if bodyGenerator != nil {
					body = bodyGenerator(getNextID())
				}
				latency, success := makeRequest(client, method, url, body)
				stats.Add(latency, success)
			}
		}()
	}

	wg.Wait()
	return stats
}

func main() {
	baseURL := "http://localhost:8080"
	if len(os.Args) > 1 {
		baseURL = os.Args[1]
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	fmt.Println("Starting load tests...")
	fmt.Printf("Target: %s\n", baseURL)
	fmt.Printf("Concurrency: 10\n")
	fmt.Printf("Requests per endpoint: 100\n\n")

	fmt.Println("Testing /health...")
	healthStats := runLoadTest(client, "GET", baseURL+"/health", nil, 10, 100)
	healthStats.Print("Health Check")

	fmt.Println("\nTesting /stats...")
	statsStats := runLoadTest(client, "GET", baseURL+"/stats", nil, 10, 100)
	statsStats.Print("Stats")

	fmt.Println("\nTesting /team/add...")
	teamBodyGenerator := func(requestID int) []byte {
		teamData, _ := json.Marshal(map[string]interface{}{
			"team_name": fmt.Sprintf("loadtest-team-%d", requestID),
			"members": []map[string]interface{}{
				{"user_id": fmt.Sprintf("lt-%d-1", requestID), "username": fmt.Sprintf("LoadTest%d-1", requestID), "is_active": true},
				{"user_id": fmt.Sprintf("lt-%d-2", requestID), "username": fmt.Sprintf("LoadTest%d-2", requestID), "is_active": true},
			},
		})
		return teamData
	}
	createTeamStats := runLoadTest(client, "POST", baseURL+"/team/add", teamBodyGenerator, 5, 50)
	createTeamStats.Print("Create Team")

	fmt.Println("\nTesting /team/get...")
	getTeamStats := runLoadTest(client, "GET", baseURL+"/team/get?team_name=loadtest-team-0", nil, 10, 100)
	getTeamStats.Print("Get Team")

	fmt.Println("\nTesting /pullRequest/create...")
	prBodyGenerator := func(requestID int) []byte {
		prData, _ := json.Marshal(map[string]interface{}{
			"pull_request_id":   fmt.Sprintf("pr-loadtest-%d", requestID),
			"pull_request_name": fmt.Sprintf("Load Test PR %d", requestID),
			"author_id":         fmt.Sprintf("lt-%d-1", requestID%50),
		})
		return prData
	}
	createPRStats := runLoadTest(client, "POST", baseURL+"/pullRequest/create", prBodyGenerator, 5, 50)
	createPRStats.Print("Create PR")

	fmt.Println("\n=== Load Test Summary ===")
	fmt.Println("All tests completed!")
}
