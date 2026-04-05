package handler

import (
	"encoding/json"
	"net/http"

	"github.com/KazikovAP/backend_course_start/hw2/internal/service"
)

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid request body"))
		return
	}

	token, err := h.auth.Register(req.Username, req.Password)
	if err != nil {
		switch err.Error() {
		case "invalid input":
			writeJSON(w, http.StatusBadRequest, errorResponse(err.Error()))
		case "user exists":
			writeJSON(w, http.StatusConflict, errorResponse(err.Error()))
		default:
			writeJSON(w, http.StatusInternalServerError, errorResponse(err.Error()))
		}
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"token": token})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid request body"))
		return
	}

	token, err := h.auth.Login(req.Username, req.Password)
	if err != nil {
		switch err.Error() {
		case "invalid input":
			writeJSON(w, http.StatusBadRequest, errorResponse(err.Error()))
		default:
			writeJSON(w, http.StatusUnauthorized, errorResponse(err.Error()))
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}
