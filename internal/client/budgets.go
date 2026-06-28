package client

import (
	"context"
	"fmt"
)

// Budget is a cost budget over an interval (month/quarter/year).
type Budget struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Enabled     *bool     `json:"enabled"`
	Scope       string    `json:"refScope"`
	Interval    string    `json:"interval"`
	Year        string    `json:"year"`
	Timezone    string    `json:"timezone"`
	Currency    string    `json:"currency"`
	Rollover    *bool     `json:"rollover"`
	Costs       []float64 `json:"costs"`
}

// BudgetInput is the create/update payload. Costs length should match Interval
// (12 for month, 4 for quarter, 1 for year).
type BudgetInput struct {
	Name        string
	Description string
	Enabled     *bool
	Scope       string
	Interval    string
	Year        string
	Timezone    string
	Currency    string
	Rollover    *bool
	Costs       []float64
}

func budgetBody(input BudgetInput) map[string]any {
	budget := map[string]any{
		"name":     input.Name,
		"period":   "year",
		"interval": input.Interval,
		"costs":    input.Costs,
	}
	if input.Description != "" {
		budget["description"] = input.Description
	}
	if input.Enabled != nil {
		budget["enabled"] = *input.Enabled
	}
	if input.Scope != "" {
		budget["scope"] = input.Scope
	}
	if input.Year != "" {
		budget["year"] = input.Year
	}
	if input.Timezone != "" {
		budget["timezone"] = input.Timezone
	}
	if input.Currency != "" {
		budget["currency"] = input.Currency
	}
	if input.Rollover != nil {
		budget["rollover"] = *input.Rollover
	}
	return map[string]any{"budget": budget}
}

func (c *Client) CreateBudget(ctx context.Context, input BudgetInput) (*Budget, error) {
	return createObj[Budget](c, ctx, "/budgets", "budget", budgetBody(input))
}

func (c *Client) GetBudget(ctx context.Context, id int64) (*Budget, error) {
	return getByID[Budget](c, ctx, fmt.Sprintf("/budgets/%d", id), "budget")
}

func (c *Client) GetBudgetByName(ctx context.Context, name string) (*Budget, error) {
	return firstByName[Budget](c, ctx, "/budgets", "budgets", name)
}

func (c *Client) UpdateBudget(ctx context.Context, id int64, input BudgetInput) (*Budget, error) {
	return updateObj[Budget](c, ctx, fmt.Sprintf("/budgets/%d", id), "budget", budgetBody(input))
}

func (c *Client) DeleteBudget(ctx context.Context, id int64) error {
	return c.delete(ctx, fmt.Sprintf("/budgets/%d", id), nil)
}

func (c *Client) ListBudgets(ctx context.Context) ([]Budget, error) {
	return listObjects[Budget](c, ctx, "/budgets", "budgets")
}
