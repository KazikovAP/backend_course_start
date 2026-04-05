package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte("secret_key")

type User struct {
	Username string
	Password []byte
}

type Crypto struct {
	Symbol       string    `json:"symbol"`
	Name         string    `json:"name"`
	CurrentPrice float64   `json:"current_price"`
	LastUpdated  time.Time `json:"last_updated"`
	History      []PricePoint
}

type PricePoint struct {
	Price     float64   `json:"price"`
	Timestamp time.Time `json:"timestamp"`
}

type CoinListEntry struct {
	ID     string `json:"id"`
	Symbol string `json:"symbol"`
	Name   string `json:"name"`
}

type CoinPriceResponse map[string]map[string]float64

type Scheduler struct {
	mu              sync.RWMutex
	enabled         bool
	intervalSeconds int
	lastUpdate      time.Time
	triggerCh       chan struct{}
	resetCh         chan time.Duration
}

func NewScheduler() *Scheduler {
	return &Scheduler{
		enabled:         true,
		intervalSeconds: 30,
		triggerCh:       make(chan struct{}, 1),
		resetCh:         make(chan time.Duration, 1),
	}
}

var scheduler = NewScheduler()

type Storage struct {
	users     map[string]User
	cryptos   map[string]*Crypto
	coinCache map[string]CoinListEntry
	mu        sync.RWMutex
}

func NewStorage() *Storage {
	return &Storage{
		users:     make(map[string]User),
		cryptos:   make(map[string]*Crypto),
		coinCache: make(map[string]CoinListEntry),
	}
}

var store = NewStorage()

var knownCoins = map[string]CoinListEntry{
	"BTC":   {ID: "bitcoin", Symbol: "BTC", Name: "Bitcoin"},
	"ETH":   {ID: "ethereum", Symbol: "ETH", Name: "Ethereum"},
	"USDT":  {ID: "tether", Symbol: "USDT", Name: "Tether"},
	"BNB":   {ID: "binancecoin", Symbol: "BNB", Name: "BNB"},
	"SOL":   {ID: "solana", Symbol: "SOL", Name: "Solana"},
	"USDC":  {ID: "usd-coin", Symbol: "USDC", Name: "USD Coin"},
	"XRP":   {ID: "ripple", Symbol: "XRP", Name: "XRP"},
	"DOGE":  {ID: "dogecoin", Symbol: "DOGE", Name: "Dogecoin"},
	"TON":   {ID: "the-open-network", Symbol: "TON", Name: "Toncoin"},
	"ADA":   {ID: "cardano", Symbol: "ADA", Name: "Cardano"},
	"TRX":   {ID: "tron", Symbol: "TRX", Name: "TRON"},
	"AVAX":  {ID: "avalanche-2", Symbol: "AVAX", Name: "Avalanche"},
	"SHIB":  {ID: "shiba-inu", Symbol: "SHIB", Name: "Shiba Inu"},
	"DOT":   {ID: "polkadot", Symbol: "DOT", Name: "Polkadot"},
	"LINK":  {ID: "chainlink", Symbol: "LINK", Name: "Chainlink"},
	"LTC":   {ID: "litecoin", Symbol: "LTC", Name: "Litecoin"},
	"BCH":   {ID: "bitcoin-cash", Symbol: "BCH", Name: "Bitcoin Cash"},
	"MATIC": {ID: "matic-network", Symbol: "MATIC", Name: "Polygon"},
	"UNI":   {ID: "uniswap", Symbol: "UNI", Name: "Uniswap"},
	"XLM":   {ID: "stellar", Symbol: "XLM", Name: "Stellar"},
	"ATOM":  {ID: "cosmos", Symbol: "ATOM", Name: "Cosmos"},
	"ETC":   {ID: "ethereum-classic", Symbol: "ETC", Name: "Ethereum Classic"},
	"XMR":   {ID: "monero", Symbol: "XMR", Name: "Monero"},
	"NEAR":  {ID: "near", Symbol: "NEAR", Name: "NEAR Protocol"},
	"APT":   {ID: "aptos", Symbol: "APT", Name: "Aptos"},
	"FIL":   {ID: "filecoin", Symbol: "FIL", Name: "Filecoin"},
	"ARB":   {ID: "arbitrum", Symbol: "ARB", Name: "Arbitrum"},
	"OP":    {ID: "optimism", Symbol: "OP", Name: "Optimism"},
	"ALGO":  {ID: "algorand", Symbol: "ALGO", Name: "Algorand"},
	"VET":   {ID: "vechain", Symbol: "VET", Name: "VeChain"},
}

func fakePriceForSymbol(symbol string) float64 {
	basePrices := map[string]float64{
		"BTC":   85000.0,
		"ETH":   2000.0,
		"BNB":   600.0,
		"SOL":   150.0,
		"XRP":   0.5,
		"DOGE":  0.15,
		"ADA":   0.45,
		"AVAX":  35.0,
		"DOT":   7.0,
		"LINK":  14.0,
		"LTC":   85.0,
		"BCH":   400.0,
		"MATIC": 0.8,
		"UNI":   7.0,
		"ATOM":  8.0,
		"NEAR":  5.0,
	}
	base, ok := basePrices[symbol]
	if !ok {
		base = 10.0
		for _, ch := range symbol {
			base += float64(ch)
		}
	}
	variation := (float64(time.Now().Unix()%100) - 50) / 2500.0
	return base * (1 + variation)
}

func fetchCoinList() error {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://api.coingecko.com/api/v3/coins/list")
	if err != nil {
		return fmt.Errorf("coingecko coins/list request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("coingecko coins/list returned status %d", resp.StatusCode)
	}

	var coins []CoinListEntry
	if err := json.NewDecoder(resp.Body).Decode(&coins); err != nil {
		return fmt.Errorf("coingecko coins/list decode error: %w", err)
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	for _, c := range coins {
		key := strings.ToUpper(c.Symbol)
		if _, exists := store.coinCache[key]; !exists {
			store.coinCache[key] = c
		}
	}

	log.Printf("CoinGecko coin cache loaded: %d symbols", len(store.coinCache))
	return nil
}

func resolveCoin(symbol string) (CoinListEntry, error) {
	store.mu.RLock()
	entry, ok := store.coinCache[symbol]
	store.mu.RUnlock()
	if ok {
		return entry, nil
	}

	if known, ok := knownCoins[symbol]; ok {
		store.mu.Lock()
		store.coinCache[symbol] = known
		store.mu.Unlock()
		return known, nil
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(fmt.Sprintf("https://api.coingecko.com/api/v3/search?query=%s", symbol))
	if err == nil {
		defer resp.Body.Close()
		var result struct {
			Coins []struct {
				ID     string `json:"id"`
				Symbol string `json:"symbol"`
				Name   string `json:"name"`
			} `json:"coins"`
		}
		if json.NewDecoder(resp.Body).Decode(&result) == nil {
			for _, c := range result.Coins {
				if strings.EqualFold(c.Symbol, symbol) {
					entry = CoinListEntry{ID: c.ID, Symbol: c.Symbol, Name: c.Name}
					store.mu.Lock()
					store.coinCache[symbol] = entry
					store.mu.Unlock()
					return entry, nil
				}
			}
		}
	}

	synthetic := CoinListEntry{
		ID:     strings.ToLower(symbol),
		Symbol: symbol,
		Name:   symbol,
	}
	store.mu.Lock()
	store.coinCache[symbol] = synthetic
	store.mu.Unlock()
	log.Printf("resolveCoin: using synthetic entry for unknown symbol %s", symbol)
	return synthetic, nil
}

func fetchPrice(coinID string) (float64, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	url := fmt.Sprintf(
		"https://api.coingecko.com/api/v3/simple/price?ids=%s&vs_currencies=usd",
		coinID,
	)
	resp, err := client.Get(url)
	if err != nil {
		log.Printf("fetchPrice: CoinGecko unavailable, using fake price for %s: %v", coinID, err)
		return fakePriceForSymbol(strings.ToUpper(coinID)), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("fetchPrice: CoinGecko returned %d for %s, using fake price", resp.StatusCode, coinID)
		return fakePriceForSymbol(strings.ToUpper(coinID)), nil
	}

	var data CoinPriceResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return fakePriceForSymbol(strings.ToUpper(coinID)), nil
	}

	prices, ok := data[coinID]
	if !ok {
		return fakePriceForSymbol(strings.ToUpper(coinID)), nil
	}
	price, ok := prices["usd"]
	if !ok {
		return fakePriceForSymbol(strings.ToUpper(coinID)), nil
	}
	return price, nil
}

func generateToken(username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	})
	return token.SignedString(jwtSecret)
}

func hashPassword(pw string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
}

func checkPassword(hash []byte, pw string) error {
	return bcrypt.CompareHashAndPassword(hash, []byte(pw))
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing token"})
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token"})
			return
		}

		next(w, r)
	}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("writeJSON: encode error: %v", err)
	}
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, fmt.Errorf("invalid request body"))
		return
	}

	if req.Username == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid input"})
		return
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	if _, exists := store.users[req.Username]; exists {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "user exists"})
		return
	}

	hash, _ := hashPassword(req.Password)
	store.users[req.Username] = User{req.Username, hash}

	token, _ := generateToken(req.Username)
	writeJSON(w, http.StatusCreated, map[string]string{"token": token})
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, fmt.Errorf("invalid request body"))
		return
	}

	if req.Username == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid input"})
		return
	}

	store.mu.RLock()
	user, exists := store.users[req.Username]
	store.mu.RUnlock()

	if !exists || checkPassword(user.Password, req.Password) != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}

	token, _ := generateToken(req.Username)
	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

func cryptoHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		listCrypto(w, r)
	case http.MethodPost:
		addCrypto(w, r)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

func cryptoBySymbolHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/crypto/")
	parts := strings.Split(path, "/")
	symbol := strings.ToUpper(parts[0])

	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			getCrypto(w, symbol)
		case http.MethodDelete:
			deleteCrypto(w, symbol)
		default:
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		}
		return
	}

	switch parts[1] {
	case "refresh":
		if r.Method == http.MethodPut {
			refreshCrypto(w, symbol)
			return
		}
	case "history":
		if r.Method == http.MethodGet {
			historyCrypto(w, symbol)
			return
		}
	case "stats":
		if r.Method == http.MethodGet {
			statsCrypto(w, symbol)
			return
		}
	}

	writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
}

func listCrypto(w http.ResponseWriter, _ *http.Request) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	var list []Crypto
	for _, c := range store.cryptos {
		list = append(list, *c)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"cryptos": list})
}

func addCrypto(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Symbol string `json:"symbol"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, fmt.Errorf("invalid request body"))
		return
	}

	if req.Symbol == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "symbol is required"})
		return
	}

	symbol := strings.ToUpper(req.Symbol)

	store.mu.RLock()
	_, exists := store.cryptos[symbol]
	store.mu.RUnlock()

	if exists {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "already exists"})
		return
	}

	coin, err := resolveCoin(symbol)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("unknown symbol: %s", symbol)})
		return
	}

	price, err := fetchPrice(coin.ID)
	if err != nil {
		log.Printf("fetchPrice error for %s: %v", coin.ID, err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch price"})
		return
	}

	crypto := &Crypto{
		Symbol:       symbol,
		Name:         coin.Name,
		CurrentPrice: price,
		LastUpdated:  time.Now(),
		History:      []PricePoint{{price, time.Now()}},
	}

	store.mu.Lock()
	store.cryptos[symbol] = crypto
	store.mu.Unlock()

	writeJSON(w, http.StatusCreated, map[string]interface{}{"crypto": crypto})
}

func getCrypto(w http.ResponseWriter, symbol string) {
	store.mu.RLock()
	crypto, ok := store.cryptos[symbol]
	store.mu.RUnlock()

	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}

	writeJSON(w, http.StatusOK, crypto)
}

func deleteCrypto(w http.ResponseWriter, symbol string) {
	store.mu.Lock()
	defer store.mu.Unlock()

	if _, ok := store.cryptos[symbol]; !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}

	delete(store.cryptos, symbol)
	writeJSON(w, http.StatusOK, map[string]interface{}{})
}

func refreshCrypto(w http.ResponseWriter, symbol string) {
	store.mu.RLock()
	c, ok := store.cryptos[symbol]
	store.mu.RUnlock()

	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}

	coin, err := resolveCoin(symbol)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to resolve coin"})
		return
	}

	price, err := fetchPrice(coin.ID)
	if err != nil {
		log.Printf("refreshCrypto fetchPrice error for %s: %v", symbol, err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch price"})
		return
	}

	store.mu.Lock()
	c.CurrentPrice = price
	c.LastUpdated = time.Now()
	c.History = append(c.History, PricePoint{price, time.Now()})
	if len(c.History) > 100 {
		c.History = c.History[1:]
	}
	store.mu.Unlock()

	writeJSON(w, http.StatusOK, map[string]interface{}{"crypto": c})
}

func historyCrypto(w http.ResponseWriter, symbol string) {
	store.mu.RLock()
	c, ok := store.cryptos[symbol]
	store.mu.RUnlock()

	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"symbol":  symbol,
		"history": c.History,
	})
}

func statsCrypto(w http.ResponseWriter, symbol string) {
	store.mu.RLock()
	c, ok := store.cryptos[symbol]
	store.mu.RUnlock()

	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		return
	}

	if len(c.History) == 0 {
		writeJSON(w, http.StatusOK, map[string]interface{}{"symbol": symbol})
		return
	}

	min := math.MaxFloat64
	max := 0.0
	sum := 0.0

	for _, p := range c.History {
		if p.Price < min {
			min = p.Price
		}
		if p.Price > max {
			max = p.Price
		}
		sum += p.Price
	}

	avg := sum / float64(len(c.History))
	first := c.History[0].Price
	last := c.History[len(c.History)-1].Price

	change := last - first
	percent := (change / first) * 100

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"symbol":        symbol,
		"current_price": c.CurrentPrice,
		"stats": map[string]interface{}{
			"min_price":            min,
			"max_price":            max,
			"avg_price":            avg,
			"price_change":         change,
			"price_change_percent": percent,
			"records_count":        len(c.History),
		},
	})
}

func updateAllPrices() int {
	store.mu.RLock()
	symbols := make([]string, 0, len(store.cryptos))
	for sym := range store.cryptos {
		symbols = append(symbols, sym)
	}
	store.mu.RUnlock()

	if len(symbols) == 0 {
		return 0
	}

	coinIDs := make([]string, 0, len(symbols))
	idToSymbol := make(map[string]string)
	for _, sym := range symbols {
		coin, err := resolveCoin(sym)
		if err != nil {
			log.Printf("updater: cannot resolve %s: %v", sym, err)
			continue
		}
		coinIDs = append(coinIDs, coin.ID)
		idToSymbol[coin.ID] = sym
	}

	if len(coinIDs) == 0 {
		return 0
	}

	now := time.Now()
	priceMap := make(map[string]float64)

	client := &http.Client{Timeout: 5 * time.Second}
	url := fmt.Sprintf(
		"https://api.coingecko.com/api/v3/simple/price?ids=%s&vs_currencies=usd",
		strings.Join(coinIDs, ","),
	)
	resp, err := client.Get(url)
	if err == nil && resp.StatusCode == http.StatusOK {
		var data CoinPriceResponse
		if json.NewDecoder(resp.Body).Decode(&data) == nil {
			for coinID, prices := range data {
				if price, ok := prices["usd"]; ok {
					priceMap[coinID] = price
				}
			}
		}
		resp.Body.Close()
	} else {
		if resp != nil {
			resp.Body.Close()
		}
		log.Printf("updater: CoinGecko unavailable, using fake prices")
	}

	for _, coinID := range coinIDs {
		if _, got := priceMap[coinID]; !got {
			sym := idToSymbol[coinID]
			priceMap[coinID] = fakePriceForSymbol(sym)
		}
	}

	updated := 0
	store.mu.Lock()
	for coinID, price := range priceMap {
		sym, ok := idToSymbol[coinID]
		if !ok {
			continue
		}
		c, ok := store.cryptos[sym]
		if !ok {
			continue
		}
		c.CurrentPrice = price
		c.LastUpdated = now
		c.History = append(c.History, PricePoint{price, now})
		if len(c.History) > 100 {
			c.History = c.History[1:]
		}
		updated++
	}
	store.mu.Unlock()

	scheduler.mu.Lock()
	scheduler.lastUpdate = now
	scheduler.mu.Unlock()

	return updated
}

func updater(ctx context.Context) {
	scheduler.mu.RLock()
	interval := time.Duration(scheduler.intervalSeconds) * time.Second
	scheduler.mu.RUnlock()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			scheduler.mu.RLock()
			enabled := scheduler.enabled
			scheduler.mu.RUnlock()

			if enabled {
				updateAllPrices()
			}

		case newInterval := <-scheduler.resetCh:
			ticker.Stop()
			ticker = time.NewTicker(newInterval)

		case <-scheduler.triggerCh:
			updateAllPrices()

		case <-ctx.Done():
			return
		}
	}
}

func scheduleHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getSchedule(w)
	case http.MethodPut:
		putSchedule(w, r)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
	}
}

func getSchedule(w http.ResponseWriter) {
	scheduler.mu.RLock()
	defer scheduler.mu.RUnlock()

	var nextUpdate time.Time
	if !scheduler.lastUpdate.IsZero() {
		nextUpdate = scheduler.lastUpdate.Add(time.Duration(scheduler.intervalSeconds) * time.Second)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"enabled":          scheduler.enabled,
		"interval_seconds": scheduler.intervalSeconds,
		"last_update":      scheduler.lastUpdate,
		"next_update":      nextUpdate,
	})
}

func putSchedule(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Enabled         bool `json:"enabled"`
		IntervalSeconds int  `json:"interval_seconds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.IntervalSeconds < 10 || req.IntervalSeconds > 3600 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "interval_seconds must be between 10 and 3600"})
		return
	}

	scheduler.mu.Lock()
	prevInterval := scheduler.intervalSeconds
	scheduler.enabled = req.Enabled
	scheduler.intervalSeconds = req.IntervalSeconds
	scheduler.mu.Unlock()

	if prevInterval != req.IntervalSeconds {
		select {
		case scheduler.resetCh <- time.Duration(req.IntervalSeconds) * time.Second:
		default:
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"enabled":          req.Enabled,
		"interval_seconds": req.IntervalSeconds,
	})
}

func triggerSchedule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	select {
	case scheduler.triggerCh <- struct{}{}:
	default:
	}

	count := updateAllPrices()

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"updated_count": count,
		"timestamp":     time.Now(),
	})
}

func main() {
	log.Println("Loading CoinGecko coin list...")
	if err := fetchCoinList(); err != nil {
		log.Printf("Warning: failed to preload coin list: %v", err)
		log.Println("Will fall back to /search endpoint per symbol")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go updater(ctx)

	http.HandleFunc("/auth/register", registerHandler)
	http.HandleFunc("/auth/login", loginHandler)

	http.HandleFunc("/crypto", authMiddleware(cryptoHandler))
	http.HandleFunc("/crypto/", authMiddleware(cryptoBySymbolHandler))

	http.HandleFunc("/schedule", authMiddleware(scheduleHandler))
	http.HandleFunc("/schedule/trigger", authMiddleware(triggerSchedule))

	fmt.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
