package client

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Instance struct {
	ID         int64          `json:"id"`
	Name       string         `json:"name"`
	Status     string         `json:"status"`
	StatusMsg  string         `json:"statusMessage"`
	IPAddress  string         `json:"ipAddress"`
	ExternalIP string         `json:"externalIp"`
	Cloud      map[string]any `json:"cloud"`
	Group      map[string]any `json:"site"`
	Plan       map[string]any `json:"plan"`
	Layout     map[string]any `json:"layout"`
	Labels     []string       `json:"labels"`
	Config     map[string]any `json:"config"`
}

type CreateInstanceInput struct {
	Name                string
	Cloud               string
	Type                string
	GroupID             int64
	LayoutID            int64
	PlanID              int64
	ResourcePoolID      string
	Description         string
	Environment         string
	Labels              []string
	Tags                map[string]string
	AvailabilityZone    string
	SecurityGroup       string
	SecurityGroups      []string
	OSExternalNetworkID string
	CreateUser          *bool
	WorkflowID          *int64
	ShutdownDays        *int64
	ExpireDays          *int64
	CreateBackup        *bool
}

type UpdateInstanceInput struct {
	Description *string
	Labels      []string
}

func (c *Client) ProvisionInstance(ctx context.Context, input CreateInstanceInput) (*Instance, error) {
	payload := map[string]any{
		"instance": map[string]any{
			"name":         input.Name,
			"cloud":        input.Cloud,
			"type":         input.Type,
			"instanceType": map[string]any{"code": input.Type},
			"site":         map[string]any{"id": input.GroupID},
			"layout":       map[string]any{"id": input.LayoutID},
			"plan":         map[string]any{"id": input.PlanID},
			"labels":       input.Labels,
			"tags":         mapTags(input.Tags),
		},
		"config": map[string]any{
			"resourcePoolId": input.ResourcePoolID,
		},
	}
	instancePayload := payload["instance"].(map[string]any)
	configPayload := payload["config"].(map[string]any)

	if input.Description != "" {
		instancePayload["description"] = input.Description
	}
	if input.Environment != "" {
		instancePayload["environment"] = input.Environment
	}
	if input.AvailabilityZone != "" {
		configPayload["availabilityZone"] = input.AvailabilityZone
	}
	if input.SecurityGroup != "" {
		configPayload["securityGroup"] = input.SecurityGroup
	}
	if len(input.SecurityGroups) > 0 {
		instancePayload["securityGroups"] = input.SecurityGroups
	}
	if input.OSExternalNetworkID != "" {
		configPayload["osExternalNetworkId"] = input.OSExternalNetworkID
	}
	if input.CreateUser != nil {
		configPayload["createUser"] = *input.CreateUser
	}
	if input.WorkflowID != nil {
		instancePayload["workflow"] = map[string]any{"id": *input.WorkflowID}
	}
	if input.ShutdownDays != nil {
		instancePayload["shutdownDays"] = *input.ShutdownDays
	}
	if input.ExpireDays != nil {
		instancePayload["expireDays"] = *input.ExpireDays
	}
	if input.CreateBackup != nil {
		instancePayload["createBackup"] = *input.CreateBackup
	}

	var response struct {
		Instance Instance `json:"instance"`
	}
	if err := c.post(ctx, "/instances", payload, &response); err != nil {
		return nil, err
	}
	return &response.Instance, nil
}

func (c *Client) GetInstance(ctx context.Context, id int64) (*Instance, error) {
	var response struct {
		Instance Instance `json:"instance"`
	}
	if err := c.get(ctx, fmt.Sprintf("/instances/%d", id), nil, &response); err != nil {
		return nil, err
	}
	return &response.Instance, nil
}

func (c *Client) UpdateInstance(ctx context.Context, id int64, input UpdateInstanceInput) (*Instance, error) {
	instancePayload := map[string]any{}
	if input.Description != nil {
		instancePayload["description"] = *input.Description
	}
	if input.Labels != nil {
		instancePayload["labels"] = input.Labels
	}
	payload := map[string]any{"instance": instancePayload}

	var response struct {
		Instance Instance `json:"instance"`
	}
	if err := c.put(ctx, fmt.Sprintf("/instances/%d", id), payload, &response); err != nil {
		return nil, err
	}
	return &response.Instance, nil
}

func (c *Client) ResizeInstance(ctx context.Context, id, planID int64) error {
	payload := map[string]any{"instance": map[string]any{"plan": map[string]any{"id": planID}}}
	var response map[string]any
	return c.put(ctx, fmt.Sprintf("/instances/%d/resize", id), payload, &response)
}

func (c *Client) DeleteInstance(ctx context.Context, id int64) error {
	return c.delete(ctx, fmt.Sprintf("/instances/%d", id), map[string]string{"force": "on"})
}

func (c *Client) WaitForInstanceStatus(ctx context.Context, id int64, target string, interval time.Duration) (*Instance, error) {
	target = strings.ToLower(target)
	if interval == 0 {
		interval = 5 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		instance, err := c.GetInstance(ctx, id)
		if err != nil {
			return nil, err
		}
		status := strings.ToLower(instance.Status)
		if status == target {
			return instance, nil
		}
		if status == "failed" {
			return nil, &APIError{StatusCode: 500, Message: fmt.Sprintf("instance %d entered failed state: %s", id, c.failureDetail(ctx, id, instance.StatusMsg))}
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
		}
	}
}

// failureDetail enriches a failed-instance message with the most recent process
// history events, falling back to the status message alone on any error.
func (c *Client) failureDetail(ctx context.Context, id int64, statusMsg string) string {
	detail := strings.TrimSpace(statusMsg)
	events, err := c.GetInstanceHistory(ctx, id)
	if err == nil && len(events) > 0 {
		start := len(events) - 2
		if start < 0 {
			start = 0
		}
		recent := strings.Join(events[start:], "; ")
		if detail == "" {
			detail = recent
		} else {
			detail = detail + " (" + recent + ")"
		}
	}
	if detail == "" {
		return "no status message provided"
	}
	return detail
}

// GetInstanceHistory returns human-readable messages from the instance's process
// history, newest last. Only processes carrying an error/status message are kept.
func (c *Client) GetInstanceHistory(ctx context.Context, id int64) ([]string, error) {
	var response struct {
		Processes []struct {
			DisplayName string `json:"displayName"`
			ProcessType struct {
				Name string `json:"name"`
			} `json:"processType"`
			Status        string `json:"status"`
			StatusMessage string `json:"statusMessage"`
			Error         string `json:"error"`
		} `json:"processes"`
	}
	if err := c.get(ctx, fmt.Sprintf("/instances/%d/history", id), nil, &response); err != nil {
		return nil, err
	}
	messages := make([]string, 0, len(response.Processes))
	for _, process := range response.Processes {
		detail := process.Error
		if detail == "" {
			detail = process.StatusMessage
		}
		if strings.TrimSpace(detail) == "" {
			continue
		}
		name := process.DisplayName
		if name == "" {
			name = process.ProcessType.Name
		}
		if name == "" {
			name = process.Status
		}
		messages = append(messages, strings.TrimSpace(name+": "+detail))
	}
	return messages, nil
}

func NormalizeResourcePoolID(value string) string {
	text := strings.TrimSpace(value)
	if _, err := strconv.ParseInt(text, 10, 64); err == nil {
		return "pool-" + text
	}
	return text
}

func mapTags(tags map[string]string) []map[string]string {
	mapped := make([]map[string]string, 0, len(tags))
	for key, value := range tags {
		mapped = append(mapped, map[string]string{"name": key, "value": value})
	}
	return mapped
}
