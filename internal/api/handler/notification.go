package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/aylinkaplan/notification-system/internal/domain"
	"github.com/aylinkaplan/notification-system/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

const (
	MaxContentSMS   = 1600  // 10 segments
	MaxContentEmail = 10000
	MaxContentPush  = 256
)

type NotificationHandler struct {
	svc *service.NotificationService
}

func NewNotificationHandler(svc *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

func (h *NotificationHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateNotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Recipient == "" || req.Channel == "" || req.Content == "" {
		respondError(w, http.StatusBadRequest, "recipient, channel, content required")
		return
	}
	if req.Channel != domain.ChannelSMS && req.Channel != domain.ChannelEmail && req.Channel != domain.ChannelPush {
		respondError(w, http.StatusBadRequest, "channel must be sms, email, or push")
		return
	}
	if req.Priority == "" {
		req.Priority = domain.PriorityNormal
	}
	if req.Priority != domain.PriorityHigh && req.Priority != domain.PriorityNormal && req.Priority != domain.PriorityLow {
		respondError(w, http.StatusBadRequest, "priority must be high, normal, or low")
		return
	}
	if maxLen := maxContentLength(req.Channel); utf8.RuneCountInString(req.Content) > maxLen {
		respondError(w, http.StatusBadRequest, "content exceeds character limit for channel")
		return
	}

	n, err := h.svc.Create(r.Context(), req, nil)
	if err == service.ErrDuplicateSend {
		respondJSON(w, http.StatusOK, n)
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, n)
}

func (h *NotificationHandler) CreateBatch(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Notifications []domain.CreateNotificationRequest `json:"notifications"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if len(req.Notifications) == 0 || len(req.Notifications) > 1000 {
		respondError(w, http.StatusBadRequest, "notifications count must be 1-1000")
		return
	}
	for i, n := range req.Notifications {
		if n.Recipient == "" || n.Channel == "" || n.Content == "" {
			respondError(w, http.StatusBadRequest, "notification "+strconv.Itoa(i)+": recipient, channel, content required")
			return
		}
		if maxLen := maxContentLength(n.Channel); utf8.RuneCountInString(n.Content) > maxLen {
			respondError(w, http.StatusBadRequest, "notification "+strconv.Itoa(i)+": content exceeds character limit")
			return
		}
	}

	batchID := uuid.New()
	list, err := h.svc.CreateBatch(r.Context(), req.Notifications, &batchID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"batch_id":       batchID,
		"notifications":  list,
		"count":          len(list),
	})
}

func (h *NotificationHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}
	n, err := h.svc.GetByID(r.Context(), id)
	if err == service.ErrNotFound {
		respondError(w, http.StatusNotFound, "notification not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, n)
}

func (h *NotificationHandler) GetByBatchID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "batchId")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid batch id")
		return
	}
	list, err := h.svc.GetByBatchID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"batch_id":      id,
		"notifications": list,
	})
}

func (h *NotificationHandler) List(w http.ResponseWriter, r *http.Request) {
	f := domain.ListFilter{Limit: 50, Offset: 0}
	if v := r.URL.Query().Get("status"); v != "" {
		s := domain.Status(v)
		f.Status = &s
	}
	if v := r.URL.Query().Get("channel"); v != "" {
		c := domain.Channel(v)
		f.Channel = &c
	}
	if v := r.URL.Query().Get("from"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err == nil {
			f.From = &t
		}
	}
	if v := r.URL.Query().Get("to"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err == nil {
			f.To = &t
		}
	}
	if v := r.URL.Query().Get("limit"); v != "" {
		if i, err := strconv.Atoi(v); err == nil && i > 0 {
			f.Limit = i
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if i, err := strconv.Atoi(v); err == nil && i >= 0 {
			f.Offset = i
		}
	}

	list, err := h.svc.List(r.Context(), f)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"notifications": list,
		"count":         len(list),
	})
}

func (h *NotificationHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}
	cancelled, err := h.svc.Cancel(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !cancelled {
		respondError(w, http.StatusConflict, "notification cannot be cancelled (not pending)")
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

func maxContentLength(channel domain.Channel) int {
	switch channel {
	case domain.ChannelSMS:
		return MaxContentSMS
	case domain.ChannelEmail:
		return MaxContentEmail
	case domain.ChannelPush:
		return MaxContentPush
	default:
		return MaxContentEmail
	}
}
