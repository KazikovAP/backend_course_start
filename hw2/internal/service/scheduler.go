package service

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/KazikovAP/backend_course_start/hw2/internal/coingecko"
	"github.com/KazikovAP/backend_course_start/hw2/internal/domain"
)

type SchedulerConfig struct {
	Enabled         bool
	IntervalSeconds int
}

type SchedulerStatus struct {
	Enabled         bool      `json:"enabled"`
	IntervalSeconds int       `json:"interval_seconds"`
	LastUpdate      time.Time `json:"last_update"`
	NextUpdate      time.Time `json:"next_update"`
}

type SchedulerService struct {
	mu              sync.RWMutex
	enabled         bool
	intervalSeconds int
	lastUpdate      time.Time

	triggerCh chan struct{}
	resetCh   chan time.Duration

	cryptos   domain.CryptoRepository
	coingecko *coingecko.Client
}

func NewSchedulerService(
	cryptos domain.CryptoRepository,
	cg *coingecko.Client,
) *SchedulerService {
	return &SchedulerService{
		enabled:         true,
		intervalSeconds: 30,
		triggerCh:       make(chan struct{}, 1),
		resetCh:         make(chan time.Duration, 1),
		cryptos:         cryptos,
		coingecko:       cg,
	}
}

func (s *SchedulerService) Run(ctx context.Context) {
	s.mu.RLock()
	interval := time.Duration(s.intervalSeconds) * time.Second
	s.mu.RUnlock()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.mu.RLock()
			enabled := s.enabled
			s.mu.RUnlock()

			if enabled {
				s.updateAllPrices()
			}

		case newInterval := <-s.resetCh:
			ticker.Stop()
			ticker = time.NewTicker(newInterval)

		case <-s.triggerCh:
			s.updateAllPrices()

		case <-ctx.Done():
			return
		}
	}
}

func (s *SchedulerService) Configure(cfg SchedulerConfig) error {
	if cfg.IntervalSeconds < 10 || cfg.IntervalSeconds > 3600 {
		return fmt.Errorf("interval_seconds must be between 10 and 3600")
	}

	s.mu.Lock()
	prevInterval := s.intervalSeconds
	s.enabled = cfg.Enabled
	s.intervalSeconds = cfg.IntervalSeconds
	s.mu.Unlock()

	if prevInterval != cfg.IntervalSeconds {
		select {
		case s.resetCh <- time.Duration(cfg.IntervalSeconds) * time.Second:
		default:
		}
	}

	return nil
}

func (s *SchedulerService) Trigger() int {
	select {
	case s.triggerCh <- struct{}{}:
	default:
	}
	return s.updateAllPrices()
}

func (s *SchedulerService) Status() SchedulerStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var nextUpdate time.Time
	if !s.lastUpdate.IsZero() {
		nextUpdate = s.lastUpdate.Add(time.Duration(s.intervalSeconds) * time.Second)
	}

	return SchedulerStatus{
		Enabled:         s.enabled,
		IntervalSeconds: s.intervalSeconds,
		LastUpdate:      s.lastUpdate,
		NextUpdate:      nextUpdate,
	}
}

func (s *SchedulerService) updateAllPrices() int {
	symbols := s.cryptos.Symbols()
	if len(symbols) == 0 {
		return 0
	}

	idToSymbol := make(map[string]string, len(symbols))
	for _, sym := range symbols {
		coin := s.coingecko.Resolve(sym)
		idToSymbol[coin.ID] = sym
	}

	prices := s.coingecko.FetchPrices(idToSymbol)

	now := time.Now()
	updated := 0

	for coinID, price := range prices {
		sym, ok := idToSymbol[coinID]
		if !ok {
			continue
		}
		if s.cryptos.Update(sym, price, now) {
			updated++
		}
	}

	s.mu.Lock()
	s.lastUpdate = now
	s.mu.Unlock()

	log.Printf("scheduler: updated %d coins", updated)
	return updated
}
