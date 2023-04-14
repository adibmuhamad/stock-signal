package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const (
	yahooFinanceAPI = "https://query1.finance.yahoo.com/v8/finance/chart/%s?interval=1d&range=2y"
)

type StockData struct {
	Close []float64 `json:"close"`
}

type Signal struct {
	Symbol       string  `json:"symbol"`
	Action       string  `json:"action"`
	Target       float64 `json:"target"`
	CurrentPrice float64 `json:"current_price"`
}

type FibonacciRetracements struct {
	Level0   float64 `json:"level_0"`
	Level236 float64 `json:"level_23.6"`
	Level382 float64 `json:"level_38.2"`
	Level500 float64 `json:"level_50"`
	Level618 float64 `json:"level_61.8"`
	Level764 float64 `json:"level_76.4"`
	Level100 float64 `json:"level_100"`
}

var upgrader = websocket.Upgrader{}

func main() {
	http.HandleFunc("/stock", stockHandler)
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func stockHandler(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading connection: %v", err)
		return
	}
	defer conn.Close()

	symbols := strings.Split(r.URL.Query().Get("symbols"), ",")
	if len(symbols) == 0 || symbols[0] == "" {
		conn.WriteMessage(websocket.TextMessage, []byte("Error: Stock symbol required."))
		return
	}

	tickerIntervalStr := r.URL.Query().Get("ticker")
		if tickerIntervalStr == "" {
			conn.WriteMessage(websocket.TextMessage, []byte("Error: missing ticker parameter."))
			return
		}

		tickerInterval, err := strconv.Atoi(tickerIntervalStr)
		if err != nil || tickerInterval <= 0 {
			conn.WriteMessage(websocket.TextMessage, []byte("Error: Invalid ticker parameter value."))
			return
		}

	conn.SetCloseHandler(func(code int, text string) error {
		message := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
		return conn.WriteControl(websocket.CloseMessage, message, time.Now().Add(time.Second))
	})

	ticker := time.NewTicker(time.Duration(tickerInterval) * time.Second)
	defer ticker.Stop()

	stop := make(chan struct{})
	go func() {
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				for _, symbol := range symbols {
					signal, err := getSignal(symbol)
					if err != nil {
						log.Printf("Error getting signal for %s: %v", symbol, err)
						continue
					}
					if err := conn.WriteJSON(signal); err != nil {
						log.Printf("Error sending signal: %v", err)
						return
					}
				}

			}
		}
	}()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			close(stop)
			break
		}
	}
}

func getSignal(symbol string) (*Signal, error) {
	stockData, err := fetchStockData(symbol)
	if err != nil {
		return nil, err
	}
	sma50, sma200 := calculateSMA(stockData)
	fibonacci := calculateFibonacci(stockData.Close)
	action, target := simpleStrategy(sma50, sma200, fibonacci)
	currentPrice := stockData.Close[len(stockData.Close)-1]
	return &Signal{Symbol: symbol, Action: action, Target: target, CurrentPrice: currentPrice}, nil
}

func fetchStockData(symbol string) (*StockData, error) {
	resp, err := http.Get(fmt.Sprintf(yahooFinanceAPI, symbol))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&data)

	closePrices := data["chart"].(map[string]interface{})["result"].([]interface{})[0].(map[string]interface{})["indicators"].(map[string]interface{})["quote"].([]interface{})[0].(map[string]interface{})["close"].([]interface{})
	stockData := &StockData{}
	for _, v := range closePrices {
		if v != nil {
			stockData.Close = append(stockData.Close, v.(float64))
		}
	}

	return stockData, nil
}

func calculateSMA(stockData *StockData) (float64, float64) {
	// Calculate the 50-day and 200-day simple moving averages (SMA)
	sma50 := 0.0
	sma200 := 0.0

	for i := len(stockData.Close) - 50; i < len(stockData.Close); i++ {
		sma50 += stockData.Close[i]
	}
	sma50 /= 50

	for i := len(stockData.Close) - 200; i < len(stockData.Close); i++ {
		sma200 += stockData.Close[i]
	}
	sma200 /= 200

	return sma50, sma200
}

func calculateFibonacci(closePrices []float64) FibonacciRetracements {
	high := closePrices[0]
	low := closePrices[0]

	for _, price := range closePrices {
		if price > high {
			high = price
		}
		if price < low {
			low = price
		}
	}

	diff := high - low
	return FibonacciRetracements{
		Level0:   low,
		Level236: low + diff*0.236,
		Level382: low + diff*0.382,
		Level500: low + diff*0.500,
		Level618: low + diff*0.618,
		Level764: low + diff*0.764,
		Level100: high,
	}
}

func simpleStrategy(sma50, sma200 float64, fibonacci FibonacciRetracements) (string, float64) {
	action := "hold"
	currentPrice := sma50
	target := currentPrice

	// Check if the current price is near one of the Fibonacci retracement levels
	threshold := 0.01 // 1% threshold to consider the price near a retracement level
	nearLevel := false
	for _, level := range []float64{fibonacci.Level236, fibonacci.Level382, fibonacci.Level500, fibonacci.Level618, fibonacci.Level764} {
		if math.Abs(currentPrice-level)/level <= threshold {
			nearLevel = true
			break
		}
	}

	// Basic price prediction 5 minutes before the current time (unrealistic and for demonstration purposes only)
	predictionFactor := 5.0 / (24 * 60) // Assuming 5 minutes in terms of trading days
	predictedPrice := currentPrice
	if sma50 > sma200 {
		predictedPrice *= (1 + predictionFactor)
	} else if sma50 < sma200 {
		predictedPrice *= (1 - predictionFactor)
	}

	// Simple trading strategy based on the predicted price and Fibonacci retracement levels
	if predictedPrice > currentPrice && nearLevel {
		action = "buy"
		target = predictedPrice * 1.05 // 5% profit target
	} else if predictedPrice < currentPrice && nearLevel {
		action = "sell"
		target = predictedPrice * 0.95 // 5% stop loss
	}

	return action, target
}
