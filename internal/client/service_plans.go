package client

import (
	"context"
	"fmt"
	"strconv"
)

type ServicePlan struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Code       string `json:"code"`
	MaxCPU     *int64 `json:"maxCpu"`
	MaxMemory  *int64 `json:"maxMemory"`
	MaxStorage *int64 `json:"maxStorage"`
}

func (c *Client) ListServicePlans(ctx context.Context, zoneID, layoutID, groupID int64) ([]ServicePlan, error) {
	// Use the provisioning option-source (/options/servicePlans). The admin
	// /instances/service-plans endpoint is not available on tenant accounts.
	var response struct {
		Data []struct {
			ID         *int64 `json:"id"`
			Name       string `json:"name"`
			Code       string `json:"code"`
			MaxCPU     *int64 `json:"maxCpu"`
			MaxMemory  *int64 `json:"maxMemory"`
			MaxStorage *int64 `json:"maxStorage"`
		} `json:"data"`
	}
	query := map[string]string{
		"zoneId":   strconv.FormatInt(zoneID, 10),
		"layoutId": strconv.FormatInt(layoutID, 10),
		"siteId":   strconv.FormatInt(groupID, 10),
	}
	if err := c.get(ctx, "/options/servicePlans", query, &response); err != nil {
		return nil, err
	}
	plans := make([]ServicePlan, 0, len(response.Data))
	for _, item := range response.Data {
		if item.ID == nil {
			continue
		}
		plans = append(plans, ServicePlan{
			ID:         *item.ID,
			Name:       item.Name,
			Code:       item.Code,
			MaxCPU:     item.MaxCPU,
			MaxMemory:  item.MaxMemory,
			MaxStorage: item.MaxStorage,
		})
	}
	return plans, nil
}

func (c *Client) GetServicePlan(ctx context.Context, name string, zoneID, layoutID, groupID int64) (*ServicePlan, error) {
	plans, err := c.ListServicePlans(ctx, zoneID, layoutID, groupID)
	if err != nil {
		return nil, err
	}
	for _, plan := range plans {
		if plan.Name == name || plan.Code == name || strconv.FormatInt(plan.ID, 10) == name {
			return &plan, nil
		}
	}
	return nil, &APIError{StatusCode: 404, Message: fmt.Sprintf("service plan %q not found", name)}
}
