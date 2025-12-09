package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/KurepinVladimir/online-subscriptions/internal/model"
	"github.com/KurepinVladimir/online-subscriptions/internal/service"
)

// in-memory репозиторий для хендлер тестов.
type handlerFakeRepo struct {
	subs map[uuid.UUID]*model.Subscription
}

func newHandlerFakeRepo() *handlerFakeRepo {
	return &handlerFakeRepo{subs: make(map[uuid.UUID]*model.Subscription)}
}

func (f *handlerFakeRepo) Create(_ context.Context, s *model.Subscription) error {
	if f.subs == nil {
		f.subs = make(map[uuid.UUID]*model.Subscription)
	}
	f.subs[s.ID] = s
	return nil
}

func (f *handlerFakeRepo) GetByID(_ context.Context, id uuid.UUID) (*model.Subscription, error) {
	if s, ok := f.subs[id]; ok {
		return s, nil
	}
	// Любая ошибка подойдёт _ хендлер по ней отдаёт 404.
	return nil, errors.New("not found")
}

func (f *handlerFakeRepo) Update(_ context.Context, s *model.Subscription) error {
	if _, ok := f.subs[s.ID]; !ok {
		return errors.New("not found")
	}
	f.subs[s.ID] = s
	return nil
}

func (f *handlerFakeRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(f.subs, id)
	return nil
}

func (f *handlerFakeRepo) List(_ context.Context, _ model.ListFilter) ([]model.Subscription, error) {
	res := make([]model.Subscription, 0, len(f.subs))
	for _, s := range f.subs {
		res = append(res, *s)
	}
	return res, nil
}

func (f *handlerFakeRepo) ListForPeriod(_ context.Context, _ model.PeriodFilter) ([]model.Subscription, error) {
	// Для хендлер тестов реальное поведение не критично
	// SummaryHandler здесь тестируем только на валидацию.
	res := make([]model.Subscription, 0, len(f.subs))
	for _, s := range f.subs {
		res = append(res, *s)
	}
	return res, nil
}

func TestCreateSubscriptionHandler_OK(t *testing.T) {
	repo := newHandlerFakeRepo()
	svc := service.NewSubscriptionService(repo)
	logg := zap.NewNop()

	h := NewSubscriptionHandler(svc, logg)

	r := chi.NewRouter()
	h.RegisterRoutes(r)

	body := map[string]any{
		"service_name": "Yandex Plus",
		"price":        400,
		"user_id":      uuid.New().String(),
		"start_date":   "07-2025",
	}
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/subscriptions/", bytes.NewReader(b))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusCreated)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	var got model.SubscriptionResponse
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal resp: %v", err)
	}

	if got.ID == uuid.Nil {
		t.Fatalf("expected non-empty id")
	}
	if got.ServiceName != "Yandex Plus" {
		t.Fatalf("service_name = %q, want %q", got.ServiceName, "Yandex Plus")
	}
	if got.Price != 400 {
		t.Fatalf("price = %d, want %d", got.Price, 400)
	}
}

func TestSummaryHandler_Validation(t *testing.T) {
	repo := newHandlerFakeRepo()
	svc := service.NewSubscriptionService(repo)
	logg := zap.NewNop()

	h := NewSubscriptionHandler(svc, logg)

	r := chi.NewRouter()
	h.RegisterRoutes(r)

	// Нет параметров from/to
	req := httptest.NewRequest(http.MethodGet, "/api/v1/subscriptions/summary", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}
