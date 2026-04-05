package coingecko

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

type CoinEntry struct {
	ID     string `json:"id"`
	Symbol string `json:"symbol"`
	Name   string `json:"name"`
}

type priceResponse map[string]map[string]float64

type Client struct {
	http      *http.Client
	baseURL   string
	coinCache map[string]CoinEntry
}

func NewClient() *Client {
	return &Client{
		http:      &http.Client{Timeout: 5 * time.Second},
		baseURL:   "https://api.coingecko.com/api/v3",
		coinCache: make(map[string]CoinEntry),
	}
}

func (c *Client) WarmCache() error {
	for sym, entry := range knownCoins {
		c.coinCache[sym] = entry
	}

	resp, err := c.http.Get(c.baseURL + "/coins/list")
	if err != nil {
		return fmt.Errorf("coins/list request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("coins/list returned status %d", resp.StatusCode)
	}

	var coins []CoinEntry
	if err := json.NewDecoder(resp.Body).Decode(&coins); err != nil {
		return fmt.Errorf("coins/list decode error: %w", err)
	}

	for _, coin := range coins {
		key := strings.ToUpper(coin.Symbol)
		if _, exists := c.coinCache[key]; !exists {
			c.coinCache[key] = coin
		}
	}

	log.Printf("coingecko: cache warmed with %d symbols", len(c.coinCache))
	return nil
}

func (c *Client) Resolve(symbol string) CoinEntry {
	if entry, ok := c.coinCache[symbol]; ok {
		return entry
	}

	resp, err := c.http.Get(fmt.Sprintf("%s/search?query=%s", c.baseURL, symbol))
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
			for _, coin := range result.Coins {
				if strings.EqualFold(coin.Symbol, symbol) {
					entry := CoinEntry{ID: coin.ID, Symbol: coin.Symbol, Name: coin.Name}
					c.coinCache[symbol] = entry
					return entry
				}
			}
		}
	}

	log.Printf("coingecko: unknown symbol %s, using synthetic entry", symbol)
	synthetic := CoinEntry{ID: strings.ToLower(symbol), Symbol: symbol, Name: symbol}
	c.coinCache[symbol] = synthetic
	return synthetic
}

func (c *Client) FetchPrices(idToSymbol map[string]string) map[string]float64 {
	result := make(map[string]float64, len(idToSymbol))

	if len(idToSymbol) == 0 {
		return result
	}

	ids := make([]string, 0, len(idToSymbol))
	for id := range idToSymbol {
		ids = append(ids, id)
	}

	url := fmt.Sprintf("%s/simple/price?ids=%s&vs_currencies=usd", c.baseURL, strings.Join(ids, ","))
	resp, err := c.http.Get(url)
	if err == nil && resp.StatusCode == http.StatusOK {
		var data priceResponse
		if json.NewDecoder(resp.Body).Decode(&data) == nil {
			for coinID, prices := range data {
				if price, ok := prices["usd"]; ok {
					result[coinID] = price
				}
			}
		}
		resp.Body.Close()
	} else {
		if resp != nil {
			resp.Body.Close()
		}
		log.Printf("coingecko: price fetch unavailable, using fake prices")
	}

	for coinID, sym := range idToSymbol {
		if _, got := result[coinID]; !got {
			result[coinID] = FakePrice(sym)
		}
	}

	return result
}

func (c *Client) FetchPrice(coinID, symbol string) float64 {
	prices := c.FetchPrices(map[string]string{coinID: symbol})
	return prices[coinID]
}

var knownCoins = map[string]CoinEntry{
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

var basePrices = map[string]float64{
	"BTC": 85000.0, "ETH": 2000.0, "BNB": 600.0, "SOL": 150.0,
	"XRP": 0.5, "DOGE": 0.15, "ADA": 0.45, "AVAX": 35.0,
	"DOT": 7.0, "LINK": 14.0, "LTC": 85.0, "BCH": 400.0,
	"MATIC": 0.8, "UNI": 7.0, "ATOM": 8.0, "NEAR": 5.0,
}

func FakePrice(symbol string) float64 {
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
