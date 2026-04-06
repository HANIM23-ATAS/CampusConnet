package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type EventParticipant struct {
	UserID    int    `json:"userId"`
	EventID   int    `json:"eventId"`
	Action    string `json:"action"` 
	Timestamp string `json:"timestamp"`
}

var (
	participants      []EventParticipant
	mu                sync.Mutex
	webhookSecret     = os.Getenv("WEBHOOK_SECRET")
	expectedApiKey    = os.Getenv("API_KEY")
	limiter           = rate.NewLimiter(1, 60) // 60 per minute is 1 per sec with burst 60. Wait, actually rate.Limit for 60/min is: 
	// limiter = rate.NewLimiter(rate.Limit(60.0/60.0), 60) but let's just make it rate.Limit(1) and burst 60
)

func init() {
	limiter = rate.NewLimiter(rate.Limit(1), 60)
}

type ProblemDetails struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail"`
}

func writeProblem(w http.ResponseWriter, status int, title, detail string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ProblemDetails{
		Type:   "about:blank",
		Title:  title,
		Status: status,
		Detail: detail,
	})
}

func middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			writeProblem(w, http.StatusTooManyRequests, "Too Many Requests", "Rate limit exceeded (60 req/min).")
			return
		}

		apiKey := r.Header.Get("X-API-Key")
		if apiKey != expectedApiKey {
			writeProblem(w, http.StatusUnauthorized, "Unauthorized", "Invalid or missing X-API-Key.")
			return
		}

		next(w, r)
	}
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeProblem(w, http.StatusMethodNotAllowed, "Method Not Allowed", "Use POST.")
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "Could not read body.")
		return
	}

	signature := r.Header.Get("X-Signature")
	mac := hmac.New(sha256.New, []byte(webhookSecret))
	mac.Write(body)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expectedMAC)) {
		writeProblem(w, http.StatusUnauthorized, "Unauthorized", "Invalid HMAC signature.")
		return
	}

	var event EventParticipant
	if err := json.Unmarshal(body, &event); err != nil {
		writeProblem(w, http.StatusBadRequest, "Bad Request", "Invalid JSON.")
		return
	}

	mu.Lock()
	participants = append(participants, event)
	mu.Unlock()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Webhook received"))
}

func handleTopEvents(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	scores := make(map[int]int)
	for _, p := range participants {
		if p.Action == "join" {
			scores[p.EventID]++
		} else if p.Action == "leave" {
			scores[p.EventID]--
		}
	}

	type kv struct {
		EventID int `json:"eventId"`
		Count   int `json:"count"`
	}
	var ss []kv
	for k, v := range scores {
		ss = append(ss, kv{k, v})
	}

	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Count > ss[j].Count
	})

	if len(ss) > 5 {
		ss = ss[:5]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"topEvents": ss})
}

func handleWeeklyStats(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	now := time.Now()
	var weekly []EventParticipant

	for _, p := range participants {
		t, err := time.Parse(time.RFC3339, p.Timestamp)
		if err == nil && now.Sub(t).Hours() <= 24*7 {
			weekly = append(weekly, p)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"weeklyEvents": weekly})
}

func main() {
	if webhookSecret == "" {
		webhookSecret = "my_webhook_secret"
	}
	if expectedApiKey == "" {
		expectedApiKey = "secure_api_key_123"
	}

	http.HandleFunc("/webhook", handleWebhook)
	http.HandleFunc("/analytics/top-events", middleware(handleTopEvents))
	http.HandleFunc("/analytics/stats", middleware(handleWeeklyStats))

	fmt.Println("Go Service running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
