// ================================================
// REST API HANDLERS
// ================================================
// HTTP REST API endpoints for dashboard data
// Cached with Redis for fast response times
// ================================================

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"time"
)

// ================================================
// DATA STRUCTURES
// ================================================

// PortfolioSummary represents portfolio overview
type PortfolioSummary struct {
	Value            float64 `json:"value"`
	DayChange        float64 `json:"dayChange"`
	DayChangePercent float64 `json:"dayChangePercent"`
	Positions        int     `json:"positions"`
	Alerts           int     `json:"alerts"`
	BuyingPower      float64 `json:"buyingPower"`
}

// WatchlistStock represents a stock in watchlist
type WatchlistStock struct {
	Symbol       string           `json:"symbol"`
	Name         string           `json:"name"`
	Price        float64          `json:"price"`
	Change       float64          `json:"change"`
	ChangePercent float64         `json:"changePercent"`
	IntradayData []float64        `json:"intradayData"`
	Metrics      StockMetrics     `json:"metrics"`
}

// StockMetrics represents additional stock metrics
type StockMetrics struct {
	RelativeVolume float64 `json:"relativeVolume"`
	RSI            int     `json:"rsi"`
	OptionsFlow    string  `json:"optionsFlow"` // bullish, bearish, neutral
}

// MarketPulse represents complete market overview
type MarketPulse struct {
	MarketIndices   MarketIndices   `json:"marketIndices"`
	TopGainers      []WatchlistStock `json:"topGainers"`
	TopLosers       []WatchlistStock `json:"topLosers"`
	SectorETFs      []WatchlistStock `json:"sectorETFs"`
	MarketBreadth   MarketBreadth    `json:"marketBreadth"`
	EconomicCalendar []EconomicEvent  `json:"economicCalendar"`
	CriticalAlerts  []CriticalAlert  `json:"criticalAlerts"`
}

// MarketIndices represents major market indices
type MarketIndices struct {
	SP500  IndexData `json:"sp500"`
	NASDAQ IndexData `json:"nasdaq"`
	DOW    IndexData `json:"dow"`
	VIX    IndexData `json:"vix"`
}

// IndexData represents individual index data
type IndexData struct {
	Value         float64 `json:"value"`
	ChangePercent float64 `json:"changePercent"`
}

// MarketBreadth represents market breadth indicators
type MarketBreadth struct {
	Advancing           int     `json:"advancing"`
	Declining           int     `json:"declining"`
	AdvanceDeclineRatio float64 `json:"advanceDeclineRatio"`
	NewHighs            int     `json:"newHighs"`
	NewLows             int     `json:"newLows"`
}

// EconomicEvent represents an economic calendar event
type EconomicEvent struct {
	Time        string `json:"time"`
	Event       string `json:"event"`
	Impact      string `json:"impact"`
	TimeUntil   string `json:"timeUntil"`
}

// CriticalAlert represents a market alert
type CriticalAlert struct {
	Symbol   string `json:"symbol"`
	Message  string `json:"message"`
	Severity string `json:"severity"`
	Time     string `json:"time"`
}

// ================================================
// CACHE KEYS
// ================================================

const (
	cacheKeyPortfolio = "api:portfolio:summary"
	cacheKeyWatchlist = "api:watchlist"
	cacheKeyMarketPulse = "api:market:pulse"
	cacheKeySnapshot  = "api:portfolio:snapshot"
	cacheKeyOpportunities = "api:opportunities:today"
	
	cacheTTL = 60 // seconds
)

// ================================================
// API ENDPOINTS
// ================================================

// handlePortfolioSummary returns portfolio summary
func handlePortfolioSummary(w http.ResponseWriter, r *http.Request) {
	enableCORS(w, r)
	if r.Method == "OPTIONS" {
		return
	}
	
	// Try cache first
	if cached, err := getFromCache(cacheKeyPortfolio); err == nil && cached != nil {
		respondJSON(w, cached)
		return
	}
	
	// Generate portfolio summary (would fetch from DB in production)
	summary := generatePortfolioSummary()
	
	// Cache result
	setCache(cacheKeyPortfolio, summary, cacheTTL)
	
	respondJSON(w, summary)
}

// handleWatchlist returns watchlist stocks
func handleWatchlist(w http.ResponseWriter, r *http.Request) {
	enableCORS(w, r)
	if r.Method == "OPTIONS" {
		return
	}
	
	// Try cache first
	if cached, err := getFromCache(cacheKeyWatchlist); err == nil && cached != nil {
		respondJSON(w, cached)
		return
	}
	
	// Generate watchlist (would fetch from DB + Polygon in production)
	watchlist := generateWatchlist()
	
	// Cache result
	setCache(cacheKeyWatchlist, watchlist, cacheTTL)
	
	respondJSON(w, watchlist)
}

// handleMarketPulse returns complete market overview
func handleMarketPulse(w http.ResponseWriter, r *http.Request) {
	enableCORS(w, r)
	if r.Method == "OPTIONS" {
		return
	}
	
	// Try cache first
	if cached, err := getFromCache(cacheKeyMarketPulse); err == nil && cached != nil {
		respondJSON(w, cached)
		return
	}
	
	// Generate market pulse
	pulse := generateMarketPulse()
	
	// Cache result
	setCache(cacheKeyMarketPulse, pulse, cacheTTL)
	
	respondJSON(w, pulse)
}

// handlePortfolioSnapshot returns detailed portfolio snapshot
func handlePortfolioSnapshot(w http.ResponseWriter, r *http.Request) {
	enableCORS(w, r)
	if r.Method == "OPTIONS" {
		return
	}
	
	// Try cache first
	if cached, err := getFromCache(cacheKeySnapshot); err == nil && cached != nil {
		respondJSON(w, cached)
		return
	}
	
	// Generate snapshot
	snapshot := generatePortfolioSnapshot()
	
	// Cache result
	setCache(cacheKeySnapshot, snapshot, cacheTTL)
	
	respondJSON(w, snapshot)
}

// handleTodaysOpportunities returns today's trade opportunities
func handleTodaysOpportunities(w http.ResponseWriter, r *http.Request) {
	enableCORS(w, r)
	if r.Method == "OPTIONS" {
		return
	}
	
	// Try cache first
	if cached, err := getFromCache(cacheKeyOpportunities); err == nil && cached != nil {
		respondJSON(w, cached)
		return
	}
	
	// Generate opportunities
	opportunities := generateOpportunities()
	
	// Cache result
	setCache(cacheKeyOpportunities, opportunities, cacheTTL)
	
	respondJSON(w, opportunities)
}

// ================================================
// DATA GENERATORS (Replace with real data in production)
// ================================================

func generatePortfolioSummary() PortfolioSummary {
	return PortfolioSummary{
		Value:            125430.50,
		DayChange:        2340.25,
		DayChangePercent: 1.9,
		Positions:        12,
		Alerts:           3,
		BuyingPower:      50000.00,
	}
}

func generateWatchlist() []WatchlistStock {
	symbols := []string{"AAPL", "TSLA", "NVDA", "MSFT", "GOOGL", "AMZN", "META", "AMD"}
	stocks := make([]WatchlistStock, len(symbols))
	
	for i, symbol := range symbols {
		price := 100.0 + float64(i*50) + rand.Float64()*100
		change := (rand.Float64() - 0.5) * 10
		changePercent := (change / price) * 100
		
		stocks[i] = WatchlistStock{
			Symbol:        symbol,
			Name:          getStockName(symbol),
			Price:         price,
			Change:        change,
			ChangePercent: changePercent,
			IntradayData:  generateIntradayData(price, 80),
			Metrics: StockMetrics{
				RelativeVolume: 1.0 + rand.Float64()*1.5,
				RSI:            40 + rand.Intn(40),
				OptionsFlow:    getRandomFlow(),
			},
		}
	}
	
	return stocks
}

func generateMarketPulse() MarketPulse {
	return MarketPulse{
		MarketIndices: MarketIndices{
			SP500:  IndexData{Value: 5780.50, ChangePercent: 0.5},
			NASDAQ: IndexData{Value: 18234.30, ChangePercent: 0.8},
			DOW:    IndexData{Value: 42458.60, ChangePercent: 0.3},
			VIX:    IndexData{Value: 14.25, ChangePercent: -2.1},
		},
		TopGainers:  generateWatchlist()[:3],
		TopLosers:   generateWatchlist()[3:6],
		SectorETFs:  generateSectorETFs(),
		MarketBreadth: MarketBreadth{
			Advancing:           2340,
			Declining:           980,
			AdvanceDeclineRatio: 2.39,
			NewHighs:            156,
			NewLows:             23,
		},
		EconomicCalendar: generateEconomicCalendar(),
		CriticalAlerts:   generateCriticalAlerts(),
	}
}

func generateSectorETFs() []WatchlistStock {
	sectors := []struct{symbol, name string}{
		{"XLK", "Technology"},
		{"XLV", "Healthcare"},
		{"XLF", "Financials"},
		{"XLE", "Energy"},
		{"XLI", "Industrials"},
		{"XLY", "Consumer Disc"},
	}
	
	stocks := make([]WatchlistStock, len(sectors))
	for i, sector := range sectors {
		change := (rand.Float64() - 0.5) * 4
		price := 100.0 + rand.Float64()*50
		
		stocks[i] = WatchlistStock{
			Symbol:        sector.symbol,
			Name:          sector.name,
			Price:         price,
			Change:        change,
			ChangePercent: (change / price) * 100,
			IntradayData:  generateIntradayData(price, 40),
			Metrics: StockMetrics{
				RelativeVolume: 0.9 + rand.Float64()*0.4,
				RSI:            45 + rand.Intn(20),
				OptionsFlow:    getRandomFlow(),
			},
		}
	}
	
	return stocks
}

func generateEconomicCalendar() []EconomicEvent {
	events := []EconomicEvent{
		{
			Time:      "2:00 PM",
			Event:     "Fed Meeting Minutes",
			Impact:    "high",
			TimeUntil: "2 hours",
		},
		{
			Time:      "4:00 PM",
			Event:     "Market Close",
			Impact:    "low",
			TimeUntil: "4 hours",
		},
	}
	
	// Add more based on day of week
	return events
}

func generateCriticalAlerts() []CriticalAlert {
	return []CriticalAlert{
		{
			Symbol:   "TSLA",
			Message:  "Unusual options activity detected: 2.5x normal volume",
			Severity: "high",
			Time:     time.Now().Format("3:04 PM"),
		},
		{
			Symbol:   "NVDA",
			Message:  "Breaking above 52-week high",
			Severity: "medium",
			Time:     time.Now().Format("3:04 PM"),
		},
	}
}

func generatePortfolioSnapshot() interface{} {
	return map[string]interface{}{
		"totalValue":      125430.50,
		"dayChange":       2340.25,
		"weekChange":      5678.90,
		"monthChange":     12345.67,
		"positions":       12,
		"topHoldings":     generateWatchlist()[:3],
		"recentTrades":    []interface{}{},
		"performanceChart": generateIntradayData(125430.50, 100),
	}
}

func generateOpportunities() interface{} {
	return map[string]interface{}{
		"opportunities": generateWatchlist()[:5],
		"criteria": []string{
			"High relative volume",
			"Strong price momentum",
			"Bullish options flow",
		},
	}
}

// ================================================
// UTILITY FUNCTIONS
// ================================================

func generateIntradayData(price float64, points int) []float64 {
	data := make([]float64, points)
	volatility := price * 0.02
	
	for i := 0; i < points; i++ {
		noise := (rand.Float64() - 0.5) * volatility
		trend := math.Sin(float64(i)/10) * volatility * 0.5
		data[i] = price + noise + trend
	}
	
	return data
}

func getStockName(symbol string) string {
	names := map[string]string{
		"AAPL":  "Apple Inc.",
		"TSLA":  "Tesla Inc.",
		"NVDA":  "NVIDIA Corp.",
		"MSFT":  "Microsoft Corp.",
		"GOOGL": "Alphabet Inc.",
		"AMZN":  "Amazon.com Inc.",
		"META":  "Meta Platforms",
		"AMD":   "Advanced Micro Devices",
	}
	if name, ok := names[symbol]; ok {
		return name
	}
	return symbol + " Inc."
}

func getRandomFlow() string {
	flows := []string{"bullish", "bearish", "neutral"}
	return flows[rand.Intn(len(flows))]
}

// ================================================
// CACHE HELPERS
// ================================================

func getFromCache(key string) (interface{}, error) {
	if redisClient == nil {
		return nil, fmt.Errorf("redis not available")
	}
	
	val, err := redisClient.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	
	var result interface{}
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		return nil, err
	}
	
	return result, nil
}

func setCache(key string, value interface{}, ttl int) {
	if redisClient == nil {
		return
	}
	
	data, err := json.Marshal(value)
	if err != nil {
		log.Printf("❌ Failed to marshal cache value: %v", err)
		return
	}
	
	if err := redisClient.Set(ctx, key, data, time.Duration(ttl)*time.Second).Err(); err != nil {
		log.Printf("❌ Failed to set cache: %v", err)
	}
}

// ================================================
// HTTP HELPERS
// ================================================

func enableCORS(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin == "" {
		origin = "*"
	}
	
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
}

func respondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("❌ Failed to encode JSON: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
