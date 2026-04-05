package repository

import (
	"fmt"
	"sync"
	"time"

	"github.com/KazikovAP/backend_course_start/hw2/internal/domain"
)

type UserMemoryRepository struct {
	mu    sync.RWMutex
	users map[string]domain.User
}

func NewUserMemoryRepository() *UserMemoryRepository {
	return &UserMemoryRepository{
		users: make(map[string]domain.User),
	}
}

func (r *UserMemoryRepository) Create(user domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.Username]; exists {
		return fmt.Errorf("user already exists")
	}
	r.users[user.Username] = user
	return nil
}

func (r *UserMemoryRepository) FindByUsername(username string) (domain.User, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.users[username]
	return user, ok
}

const maxHistorySize = 100

type CryptoMemoryRepository struct {
	mu      sync.RWMutex
	cryptos map[string]*domain.Crypto
}

func NewCryptoMemoryRepository() *CryptoMemoryRepository {
	return &CryptoMemoryRepository{
		cryptos: make(map[string]*domain.Crypto),
	}
}

func (r *CryptoMemoryRepository) Add(crypto *domain.Crypto) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.cryptos[crypto.Symbol]; exists {
		return fmt.Errorf("crypto already exists")
	}
	r.cryptos[crypto.Symbol] = crypto
	return nil
}

func (r *CryptoMemoryRepository) GetAll() []*domain.Crypto {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]*domain.Crypto, 0, len(r.cryptos))
	for _, c := range r.cryptos {
		list = append(list, c)
	}
	return list
}

func (r *CryptoMemoryRepository) GetBySymbol(symbol string) (*domain.Crypto, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	c, ok := r.cryptos[symbol]
	if !ok {
		return nil, false
	}

	snapshot := *c
	snapshot.History = make([]domain.PricePoint, len(c.History))
	copy(snapshot.History, c.History)
	return &snapshot, true
}

func (r *CryptoMemoryRepository) Delete(symbol string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.cryptos[symbol]; !ok {
		return false
	}
	delete(r.cryptos, symbol)
	return true
}

func (r *CryptoMemoryRepository) Update(symbol string, price float64, at time.Time) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	c, ok := r.cryptos[symbol]
	if !ok {
		return false
	}

	c.CurrentPrice = price
	c.LastUpdated = at
	c.History = append(c.History, domain.PricePoint{Price: price, Timestamp: at})
	if len(c.History) > maxHistorySize {
		c.History = c.History[1:]
	}
	return true
}

func (r *CryptoMemoryRepository) Symbols() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	symbols := make([]string, 0, len(r.cryptos))
	for sym := range r.cryptos {
		symbols = append(symbols, sym)
	}
	return symbols
}
