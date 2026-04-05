package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/KazikovAP/backend_course_start/hw2/internal/service"
)

type ScheduleHandler struct {
	scheduler *service.SchedulerService
}

func NewScheduleHandler(scheduler *service.SchedulerService) *ScheduleHandler {
	return &ScheduleHandler{scheduler: scheduler}
}

func (h *ScheduleHandler) Schedule(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getStatus(w)
	case http.MethodPut:
		h.configure(w, r)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse("method not allowed"))
	}
}

func (h *ScheduleHandler) Trigger(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse("method not allowed"))
		return
	}

	count := h.scheduler.Trigger()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"updated_count": count,
		"timestamp":     time.Now(),
	})
}

func (h *ScheduleHandler) getStatus(w http.ResponseWriter) {
	status := h.scheduler.Status()
	writeJSON(w, http.StatusOK, status)
}

func (h *ScheduleHandler) configure(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Enabled         bool `json:"enabled"`
		IntervalSeconds int  `json:"interval_seconds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid request body"))
		return
	}

	cfg := service.SchedulerConfig{
		Enabled:         req.Enabled,
		IntervalSeconds: req.IntervalSeconds,
	}

	if err := h.scheduler.Configure(cfg); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"enabled":          req.Enabled,
		"interval_seconds": req.IntervalSeconds,
	})
}
