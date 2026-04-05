package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/KazikovAP/backend_course_start/hw2/internal/domain"
	"github.com/KazikovAP/backend_course_start/hw2/internal/service"
)

type CryptoHandler struct {
	crypto *service.CryptoService
}

func NewCryptoHandler(crypto *service.CryptoService) *CryptoHandler {
	return &CryptoHandler{crypto: crypto}
}

func (h *CryptoHandler) Collection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.list(w, r)
	case http.MethodPost:
		h.add(w, r)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse("method not allowed"))
	}
}

func (h *CryptoHandler) Item(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/crypto/")
	parts := strings.SplitN(path, "/", 2)
	symbol := strings.ToUpper(parts[0])

	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			h.get(w, symbol)
		case http.MethodDelete:
			h.delete(w, symbol)
		default:
			writeJSON(w, http.StatusMethodNotAllowed, errorResponse("method not allowed"))
		}
		return
	}

	switch parts[1] {
	case "refresh":
		if r.Method == http.MethodPut {
			h.refresh(w, symbol)
			return
		}
	case "history":
		if r.Method == http.MethodGet {
			h.history(w, symbol)
			return
		}
	case "stats":
		if r.Method == http.MethodGet {
			h.stats(w, symbol)
			return
		}
	}

	writeJSON(w, http.StatusNotFound, errorResponse("not found"))
}

func (h *CryptoHandler) list(w http.ResponseWriter, _ *http.Request) {
	cryptos := h.crypto.GetAll()
	if cryptos == nil {
		cryptos = []*domain.Crypto{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"cryptos": cryptos})
}

func (h *CryptoHandler) add(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Symbol string `json:"symbol"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid request body"))
		return
	}

	symbol := strings.ToUpper(req.Symbol)
	crypto, err := h.crypto.Add(symbol)
	if err != nil {
		switch err.Error() {
		case "symbol is required":
			writeJSON(w, http.StatusBadRequest, errorResponse(err.Error()))
		case "already exists":
			writeJSON(w, http.StatusConflict, errorResponse(err.Error()))
		default:
			writeJSON(w, http.StatusInternalServerError, errorResponse(err.Error()))
		}
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{"crypto": crypto})
}

func (h *CryptoHandler) get(w http.ResponseWriter, symbol string) {
	crypto, err := h.crypto.GetBySymbol(symbol)
	if err != nil {
		writeJSON(w, http.StatusNotFound, errorResponse("not found"))
		return
	}
	writeJSON(w, http.StatusOK, crypto)
}

func (h *CryptoHandler) delete(w http.ResponseWriter, symbol string) {
	if err := h.crypto.Delete(symbol); err != nil {
		writeJSON(w, http.StatusNotFound, errorResponse("not found"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{})
}

func (h *CryptoHandler) refresh(w http.ResponseWriter, symbol string) {
	crypto, err := h.crypto.Refresh(symbol)
	if err != nil {
		writeJSON(w, http.StatusNotFound, errorResponse("not found"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"crypto": crypto})
}

func (h *CryptoHandler) history(w http.ResponseWriter, symbol string) {
	history, err := h.crypto.History(symbol)
	if err != nil {
		writeJSON(w, http.StatusNotFound, errorResponse("not found"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"symbol": symbol, "history": history})
}

func (h *CryptoHandler) stats(w http.ResponseWriter, symbol string) {
	crypto, stats, err := h.crypto.Stats(symbol)
	if err != nil {
		writeJSON(w, http.StatusNotFound, errorResponse("not found"))
		return
	}

	if stats == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"symbol": symbol})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"symbol":        symbol,
		"current_price": crypto.CurrentPrice,
		"stats":         stats,
	})
}
