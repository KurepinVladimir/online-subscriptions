package model

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const yearMonthLayout = "01-2006" // "MM-YYYY"

// in JSON  "MM-YYYY".
type YearMonth struct {
	time.Time
}

func ParseYearMonth(s string) (YearMonth, error) {
	t, err := time.Parse(yearMonthLayout, s)
	if err != nil {
		return YearMonth{}, fmt.Errorf("invalid year-month %q: %w", s, err)
	}
	// to 1 data of month
	return YearMonth{Time: time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)}, nil
}

func (ym YearMonth) String() string {
	return ym.Time.Format(yearMonthLayout)
}

func (ym *YearMonth) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	v, err := ParseYearMonth(s)
	if err != nil {
		return err
	}
	*ym = v
	return nil
}

func (ym YearMonth) MarshalJSON() ([]byte, error) {
	return json.Marshal(ym.String())
}

// In base model
type Subscription struct {
	ID          uuid.UUID  `db:"id" json:"id"`
	ServiceName string     `db:"service_name" json:"service_name"`
	Price       int        `db:"price" json:"price"`
	UserID      uuid.UUID  `db:"user_id" json:"user_id"`
	StartMonth  time.Time  `db:"start_month" json:"-"`
	EndMonth    *time.Time `db:"end_month" json:"-"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
}

// DTO for in/out JSON.

type SubscriptionCreateRequest struct {
	ServiceName string     `json:"service_name"`
	Price       int        `json:"price"`
	UserID      uuid.UUID  `json:"user_id"`
	StartDate   YearMonth  `json:"start_date"`
	EndDate     *YearMonth `json:"end_date,omitempty"`
}

type SubscriptionUpdateRequest struct {
	ServiceName *string    `json:"service_name,omitempty"`
	Price       *int       `json:"price,omitempty"`
	StartDate   *YearMonth `json:"start_date,omitempty"`
	EndDate     *YearMonth `json:"end_date,omitempty"`
}

type SubscriptionResponse struct {
	ID          uuid.UUID  `json:"id"`
	ServiceName string     `json:"service_name"`
	Price       int        `json:"price"`
	UserID      uuid.UUID  `json:"user_id"`
	StartDate   YearMonth  `json:"start_date"`
	EndDate     *YearMonth `json:"end_date,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func NewSubscriptionResponse(s Subscription) SubscriptionResponse {
	res := SubscriptionResponse{
		ID:          s.ID,
		ServiceName: s.ServiceName,
		Price:       s.Price,
		UserID:      s.UserID,
		StartDate:   YearMonth{Time: s.StartMonth},
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
	}
	if s.EndMonth != nil {
		res.EndDate = &YearMonth{Time: *s.EndMonth}
	}
	return res
}

type ListFilter struct {
	Limit       int
	Offset      int
	UserID      *uuid.UUID
	ServiceName *string
}

type PeriodFilter struct {
	From        YearMonth
	To          YearMonth
	UserID      *uuid.UUID
	ServiceName *string
}
