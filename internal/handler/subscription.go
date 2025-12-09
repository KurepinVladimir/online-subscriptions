package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/KurepinVladimir/online-subscriptions/internal/model"
	"github.com/KurepinVladimir/online-subscriptions/internal/service"
)

type SubscriptionHandler struct {
	svc    *service.SubscriptionService
	logger *zap.Logger
}

func NewSubscriptionHandler(svc *service.SubscriptionService, logger *zap.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{svc: svc, logger: logger}
}

func (h *SubscriptionHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/subscriptions", func(r chi.Router) {
		r.Post("/", h.Create)
		r.Get("/", h.List)
		r.Get("/summary", h.Summary)

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.Get)
			r.Put("/", h.Update)
			r.Delete("/", h.Delete)
		})
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func (h *SubscriptionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.SubscriptionCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("decode create request", zap.Error(err))
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	sub, err := h.svc.Create(r.Context(), req)
	if err != nil {
		h.logger.Error("create subscription", zap.Error(err))
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp := model.NewSubscriptionResponse(*sub)
	writeJSON(w, http.StatusCreated, resp)
}

func (h *SubscriptionHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	sub, err := h.svc.Get(r.Context(), id)
	if err != nil {
		h.logger.Error("get subscription", zap.Error(err))
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	resp := model.NewSubscriptionResponse(*sub)
	writeJSON(w, http.StatusOK, resp)
}

func (h *SubscriptionHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req model.SubscriptionUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("decode update request", zap.Error(err))
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	sub, err := h.svc.Update(r.Context(), id, req)
	if err != nil {
		h.logger.Error("update subscription", zap.Error(err))
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp := model.NewSubscriptionResponse(*sub)
	writeJSON(w, http.StatusOK, resp)
}

func (h *SubscriptionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.svc.Delete(r.Context(), id); err != nil {
		h.logger.Error("delete subscription", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *SubscriptionHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	var filter model.ListFilter

	if limitStr := q.Get("limit"); limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = v
		}
	}
	if offsetStr := q.Get("offset"); offsetStr != "" {
		if v, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = v
		}
	}

	if userIDStr := q.Get("user_id"); userIDStr != "" {
		if id, err := uuid.Parse(userIDStr); err == nil {
			filter.UserID = &id
		} else {
			writeError(w, http.StatusBadRequest, "invalid user_id")
			return
		}
	}

	if svcName := q.Get("service_name"); svcName != "" {
		filter.ServiceName = &svcName
	}

	subs, err := h.svc.List(r.Context(), filter)
	if err != nil {
		h.logger.Error("list subscriptions", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	resp := make([]model.SubscriptionResponse, 0, len(subs))
	for _, s := range subs {
		resp = append(resp, model.NewSubscriptionResponse(s))
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *SubscriptionHandler) Summary(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	fromStr := q.Get("from")
	toStr := q.Get("to")
	if fromStr == "" || toStr == "" {
		writeError(w, http.StatusBadRequest, "from and to are required (MM-YYYY)")
		return
	}

	from, err := model.ParseYearMonth(fromStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid from")
		return
	}
	to, err := model.ParseYearMonth(toStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid to")
		return
	}

	var filter model.PeriodFilter
	filter.From = from
	filter.To = to

	if userIDStr := q.Get("user_id"); userIDStr != "" {
		id, err := uuid.Parse(userIDStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid user_id")
			return
		}
		filter.UserID = &id
	}
	if svcName := q.Get("service_name"); svcName != "" {
		filter.ServiceName = &svcName
	}

	total, err := h.svc.CalculateTotalForPeriod(r.Context(), filter)
	if err != nil {
		h.logger.Error("summary", zap.Error(err))
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]int{"total": total})
}
