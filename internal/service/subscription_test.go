package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/KurepinVladimir/online-subscriptions/internal/model"
)

// in-memory реализация репозитория для тестов.
type fakeRepo struct {
	subs              map[uuid.UUID]*model.Subscription
	listForPeriodSubs []model.Subscription
	listForPeriodErr  error
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{subs: make(map[uuid.UUID]*model.Subscription)}
}

func (f *fakeRepo) Create(_ context.Context, s *model.Subscription) error {
	if f.subs == nil {
		f.subs = make(map[uuid.UUID]*model.Subscription)
	}
	now := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	s.CreatedAt = now
	s.UpdatedAt = now
	f.subs[s.ID] = s
	return nil
}

func (f *fakeRepo) GetByID(_ context.Context, id uuid.UUID) (*model.Subscription, error) {
	if s, ok := f.subs[id]; ok {
		return s, nil
	}
	return nil, ErrNotFound
}

func (f *fakeRepo) Update(_ context.Context, s *model.Subscription) error {
	if _, ok := f.subs[s.ID]; !ok {
		return ErrNotFound
	}
	s.UpdatedAt = time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)
	f.subs[s.ID] = s
	return nil
}

func (f *fakeRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(f.subs, id)
	return nil
}

func (f *fakeRepo) List(_ context.Context, _ model.ListFilter) ([]model.Subscription, error) {
	res := make([]model.Subscription, 0, len(f.subs))
	for _, s := range f.subs {
		res = append(res, *s)
	}
	return res, nil
}

func (f *fakeRepo) ListForPeriod(_ context.Context, _ model.PeriodFilter) ([]model.Subscription, error) {
	if f.listForPeriodErr != nil {
		return nil, f.listForPeriodErr
	}
	return f.listForPeriodSubs, nil
}

func (f *fakeRepo) SumForPeriod(_ context.Context, filter model.PeriodFilter) (int, error) {
	// Будем считать сумму по тем данным, которые тест уже подготовил в listForPeriodSubs.
	// Это оставляет тест unit-тестом (без БД), но проверяет бизнес-логику.
	total := 0

	for _, sub := range f.listForPeriodSubs {
		// подписка может быть "бесконечной", но в тестах EndMonth всегда задан.
		// если EndMonth nil — считаем, что она действует до конца запрошенного периода.
		subTo := filter.To
		if sub.EndMonth != nil {
			subTo = model.YearMonth{Time: *sub.EndMonth}
		}

		m := monthsOverlap(
			model.YearMonth{Time: sub.StartMonth},
			subTo,
			filter.From,
			filter.To,
		)
		if m > 0 {
			total += sub.Price * m
		}
	}

	return total, nil
}

// ErrNotFound локальная ошибка, чтобы не поднимать database/sql в тестах.
type notFoundError struct{}

func (notFoundError) Error() string { return "not found" }

var ErrNotFound = notFoundError{}

func TestMonthsOverlap(t *testing.T) {
	mustYM := func(s string) model.YearMonth {
		ym, err := model.ParseYearMonth(s)
		if err != nil {
			t.Fatalf("parse %s: %v", s, err)
		}
		return ym
	}

	tests := []struct {
		name     string
		aFrom    string
		aTo      string
		bFrom    string
		bTo      string
		expected int
	}{
		{
			name:     "no overlap",
			aFrom:    "01-2025",
			aTo:      "03-2025",
			bFrom:    "04-2025",
			bTo:      "06-2025",
			expected: 0,
		},
		{
			name:     "full overlap",
			aFrom:    "01-2025",
			aTo:      "03-2025",
			bFrom:    "01-2025",
			bTo:      "03-2025",
			expected: 3,
		},
		{
			name:     "partial overlap left",
			aFrom:    "01-2025",
			aTo:      "05-2025",
			bFrom:    "03-2025",
			bTo:      "07-2025",
			expected: 3, // 03,04,05
		},
		{
			name:     "partial overlap right",
			aFrom:    "03-2025",
			aTo:      "07-2025",
			bFrom:    "01-2025",
			bTo:      "05-2025",
			expected: 3, // 03,04,05
		},
		{
			name:     "single month",
			aFrom:    "03-2025",
			aTo:      "03-2025",
			bFrom:    "03-2025",
			bTo:      "03-2025",
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aFrom := mustYM(tt.aFrom)
			aTo := mustYM(tt.aTo)
			bFrom := mustYM(tt.bFrom)
			bTo := mustYM(tt.bTo)

			got := monthsOverlap(aFrom, aTo, bFrom, bTo)
			if got != tt.expected {
				t.Fatalf("monthsOverlap = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestCalculateTotalForPeriod(t *testing.T) {
	repo := newFakeRepo()

	mustYM := func(s string) model.YearMonth {
		ym, err := model.ParseYearMonth(s)
		if err != nil {
			t.Fatalf("parse %s: %v", s, err)
		}
		return ym
	}

	// Одна подписка 400 руб, с 01-2025 по 03-2025 (3 месяца)
	start1 := mustYM("01-2025").Time
	end1 := mustYM("03-2025").Time
	sub1 := model.Subscription{
		ID:          uuid.New(),
		ServiceName: "Yandex Plus",
		Price:       400,
		UserID:      uuid.New(),
		StartMonth:  start1,
		EndMonth:    &end1,
	}

	// Вторая 500 руб, с 02-2025 по 04-2025 (3 месяца)
	start2 := mustYM("02-2025").Time
	end2 := mustYM("04-2025").Time
	sub2 := model.Subscription{
		ID:          uuid.New(),
		ServiceName: "IVI",
		Price:       500,
		UserID:      uuid.New(),
		StartMonth:  start2,
		EndMonth:    &end2,
	}

	repo.listForPeriodSubs = []model.Subscription{sub1, sub2}

	svc := NewSubscriptionService(repo)

	period := model.PeriodFilter{
		From: mustYM("02-2025"),
		To:   mustYM("03-2025"),
	}

	total, err := svc.CalculateTotalForPeriod(context.Background(), period)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// sub1 400 * 2 (02,03) = 800
	// sub2 500 * 2 (02,03) = 1000
	// total = 1800
	if total != 1800 {
		t.Fatalf("unexpected total: got %d, want %d", total, 1800)
	}
}

func TestCreateValidation(t *testing.T) {
	repo := newFakeRepo()
	svc := NewSubscriptionService(repo)

	ym, err := model.ParseYearMonth("07-2025")
	if err != nil {
		t.Fatalf("parse year month: %v", err)
	}

	req := model.SubscriptionCreateRequest{
		ServiceName: "Test",
		Price:       0, // невалидно
		UserID:      uuid.New(),
		StartDate:   ym,
	}

	_, err = svc.Create(context.Background(), req)
	if err == nil {
		t.Fatalf("expected error for zero price, got nil")
	}
}
