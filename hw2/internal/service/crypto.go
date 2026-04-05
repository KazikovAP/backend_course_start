package service

import (
	"fmt"
	"math"
	"time"

	"github.com/KazikovAP/backend_course_start/hw2/internal/coingecko"
	"github.com/KazikovAP/backend_course_start/hw2/internal/domain"
)

type CryptoService struct {
	cryptos   domain.CryptoRepository
	coingecko *coingecko.Client
}

func NewCryptoService(cryptos domain.CryptoRepository, cg *coingecko.Client) *CryptoService {
	return &CryptoService{cryptos: cryptos, coingecko: cg}
}

func (s *CryptoService) Add(symbol string) (*domain.Crypto, error) {
	if symbol == "" {
		return nil, fmt.Errorf("symbol is required")
	}

	coin := s.coingecko.Resolve(symbol)
	price := s.coingecko.FetchPrice(coin.ID, symbol)

	crypto := &domain.Crypto{
		Symbol:       symbol,
		Name:         coin.Name,
		CurrentPrice: price,
		LastUpdated:  time.Now(),
		History:      []domain.PricePoint{{Price: price, Timestamp: time.Now()}},
	}

	if err := s.cryptos.Add(crypto); err != nil {
		return nil, fmt.Errorf("already exists")
	}

	return crypto, nil
}

func (s *CryptoService) GetAll() []*domain.Crypto {
	return s.cryptos.GetAll()
}

func (s *CryptoService) GetBySymbol(symbol string) (*domain.Crypto, error) {
	c, ok := s.cryptos.GetBySymbol(symbol)
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return c, nil
}

func (s *CryptoService) Delete(symbol string) error {
	if !s.cryptos.Delete(symbol) {
		return fmt.Errorf("not found")
	}
	return nil
}

func (s *CryptoService) Refresh(symbol string) (*domain.Crypto, error) {
	_, ok := s.cryptos.GetBySymbol(symbol)
	if !ok {
		return nil, fmt.Errorf("not found")
	}

	coin := s.coingecko.Resolve(symbol)
	price := s.coingecko.FetchPrice(coin.ID, symbol)

	s.cryptos.Update(symbol, price, time.Now())

	updated, _ := s.cryptos.GetBySymbol(symbol)
	return updated, nil
}

func (s *CryptoService) History(symbol string) ([]domain.PricePoint, error) {
	c, ok := s.cryptos.GetBySymbol(symbol)
	if !ok {
		return nil, fmt.Errorf("not found")
	}

	history := make([]domain.PricePoint, len(c.History))
	copy(history, c.History)
	return history, nil
}

type Stats struct {
	MinPrice           float64 `json:"min_price"`
	MaxPrice           float64 `json:"max_price"`
	AvgPrice           float64 `json:"avg_price"`
	PriceChange        float64 `json:"price_change"`
	PriceChangePercent float64 `json:"price_change_percent"`
	RecordsCount       int     `json:"records_count"`
}

func (s *CryptoService) Stats(symbol string) (*domain.Crypto, *Stats, error) {
	c, ok := s.cryptos.GetBySymbol(symbol)
	if !ok {
		return nil, nil, fmt.Errorf("not found")
	}

	history := make([]domain.PricePoint, len(c.History))
	copy(history, c.History)

	if len(history) == 0 {
		return c, nil, nil
	}

	min := math.MaxFloat64
	max := 0.0
	sum := 0.0

	for _, p := range history {
		if p.Price < min {
			min = p.Price
		}
		if p.Price > max {
			max = p.Price
		}
		sum += p.Price
	}

	first := history[0].Price
	last := history[len(history)-1].Price
	change := last - first

	stats := &Stats{
		MinPrice:           min,
		MaxPrice:           max,
		AvgPrice:           sum / float64(len(history)),
		PriceChange:        change,
		PriceChangePercent: (change / first) * 100,
		RecordsCount:       len(history),
	}

	return c, stats, nil
}
