package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type chartResponse struct {
	Chart struct {
		Result []struct {
			Meta struct {
				RegularMarketPrice float64 `json:"regularMarketPrice"`
			} `json:"meta"`
		} `json:"result"`
		Error interface{} `json:"error"`
	} `json:"chart"`
}

type convertResponse struct {
	Base      string  `json:"base"`
	Target    string  `json:"target"`
	Amount    float64 `json:"amount"`
	Rate      float64 `json:"rate"`
	Converted float64 `json:"converted"`
	Source    string  `json:"source"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/convert", convertHandler)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	handler := withCORS(mux)

	addr := ":8080"
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}

	log.Printf("currency-converter backend listening on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func convertHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	base := strings.ToUpper(r.URL.Query().Get("base"))
	target := strings.ToUpper(r.URL.Query().Get("target"))
	amountStr := r.URL.Query().Get("amount")

	if base == "" || target == "" {
		http.Error(w, "base and target query parameters are required", http.StatusBadRequest)
		return
	}

	amount := 1.0
	if amountStr != "" {
		parsed, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			http.Error(w, "amount must be a number", http.StatusBadRequest)
			return
		}
		amount = parsed
	}

	rate, err := rateFetcher(base, target)
	if err != nil {
		log.Printf("failed to fetch rate: %v", err)
		http.Error(w, "failed to fetch rate", http.StatusBadGateway)
		return
	}

	resp := convertResponse{
		Base:      base,
		Target:    target,
		Amount:    amount,
		Rate:      rate,
		Converted: rate * amount,
		Source:    "yahoo-finance",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}

var rateFetcher = fetchRate

func fetchRate(base, target string) (float64, error) {
	symbol := base + target + "=X"
	endpoint := fmt.Sprintf("https://query1.finance.yahoo.com/v8/finance/chart/%s?range=1d&interval=1m", symbol)

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return 0, err
	}

	req.Header.Set("User-Agent", "currency-converter-agent/1.0")

	res, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status code %d", res.StatusCode)
	}

	var payload chartResponse
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return 0, err
	}

	if payload.Chart.Error != nil {
		return 0, errors.New("chart api returned an error")
	}

	if len(payload.Chart.Result) == 0 {
		return 0, errors.New("chart api returned no results")
	}

	price := payload.Chart.Result[0].Meta.RegularMarketPrice
	if price == 0 {
		return 0, errors.New("received zero price from api")
	}

	return price, nil
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
