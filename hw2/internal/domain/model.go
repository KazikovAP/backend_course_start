package domain

import "time"

type User struct {
	Username string
	Password []byte
}

type Crypto struct {
	Symbol       string       `json:"symbol"`
	Name         string       `json:"name"`
	CurrentPrice float64      `json:"current_price"`
	LastUpdated  time.Time    `json:"last_updated"`
	History      []PricePoint `json:"history,omitempty"`
}

type PricePoint struct {
	Price     float64   `json:"price"`
	Timestamp time.Time `json:"timestamp"`
}

type UserRepository interface {
	Create(user User) error
	FindByUsername(username string) (User, bool)
}

type CryptoRepository interface {
	Add(crypto *Crypto) error
	GetAll() []*Crypto
	GetBySymbol(symbol string) (*Crypto, bool)
	Delete(symbol string) bool
	Update(symbol string, price float64, at time.Time) bool
	Symbols() []string
}
