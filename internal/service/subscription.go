package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/KurepinVladimir/online-subscriptions/internal/model"
	"github.com/KurepinVladimir/online-subscriptions/internal/repository"
)

type SubscriptionService struct {
	repo repository.SubscriptionRepository
}

func NewSubscriptionService(repo repository.SubscriptionRepository) *SubscriptionService {
	return &SubscriptionService{repo: repo}
}

func (s *SubscriptionService) Create(ctx context.Context, req model.SubscriptionCreateRequest) (*model.Subscription, error) {
	if req.Price <= 0 {
		return nil, fmt.Errorf("price must be positive")
	}

	sub := &model.Subscription{
		ID:          uuid.New(),
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      req.UserID,
		StartMonth:  req.StartDate.Time,
	}
	if req.EndDate != nil {
		t := req.EndDate.Time
		sub.EndMonth = &t
	}

	if err := s.repo.Create(ctx, sub); err != nil {
		return nil, err
	}
	return sub, nil
}

func (s *SubscriptionService) Get(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *SubscriptionService) Update(ctx context.Context, id uuid.UUID, req model.SubscriptionUpdateRequest) (*model.Subscription, error) {
	sub, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.ServiceName != nil {
		sub.ServiceName = *req.ServiceName
	}
	if req.Price != nil {
		if *req.Price <= 0 {
			return nil, fmt.Errorf("price must be positive")
		}
		sub.Price = *req.Price
	}
	if req.StartDate != nil {
		sub.StartMonth = req.StartDate.Time
	}
	if req.EndDate != nil {
		t := req.EndDate.Time
		sub.EndMonth = &t
	}

	if err := s.repo.Update(ctx, sub); err != nil {
		return nil, err
	}
	return sub, nil
}

func (s *SubscriptionService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *SubscriptionService) List(ctx context.Context, filter model.ListFilter) ([]model.Subscription, error) {
	return s.repo.List(ctx, filter)
}

func (s *SubscriptionService) CalculateTotalForPeriod(ctx context.Context, filter model.PeriodFilter) (int, error) {
	if filter.To.Time.Before(filter.From.Time) {
		return 0, fmt.Errorf("invalid period: to < from")
	}

	subs, err := s.repo.ListForPeriod(ctx, filter)
	if err != nil {
		return 0, err
	}

	total := 0
	for _, sub := range subs {
		months := monthsOverlap(
			model.YearMonth{Time: sub.StartMonth},
			yearMonthOrInfinity(sub.EndMonth),
			filter.From,
			filter.To,
		)
		if months > 0 {
			total += sub.Price * months
		}
	}
	return total, nil
}

func yearMonthOrInfinity(t *time.Time) model.YearMonth {
	if t == nil {
		return model.YearMonth{Time: time.Date(9999, 12, 1, 0, 0, 0, 0, time.UTC)}
	}
	return model.YearMonth{Time: *t}
}

func ymToInt(ym model.YearMonth) int {
	return ym.Year()*12 + int(ym.Month())
}

func monthsOverlap(aFrom, aTo, bFrom, bTo model.YearMonth) int {
	from := maxYM(aFrom, bFrom)
	to := minYM(aTo, bTo)

	if to.Time.Before(from.Time) {
		return 0
	}

	start := ymToInt(from)
	end := ymToInt(to)

	return end - start + 1
}

func maxYM(a, b model.YearMonth) model.YearMonth {
	if a.Time.After(b.Time) {
		return a
	}
	return b
}

func minYM(a, b model.YearMonth) model.YearMonth {
	if a.Time.Before(b.Time) {
		return a
	}
	return b
}
