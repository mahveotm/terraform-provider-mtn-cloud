package client

import (
	"context"
	"fmt"
)

type InstanceType struct {
	ID              int64       `json:"id"`
	Name            string      `json:"name"`
	Code            string      `json:"code"`
	DefaultLayoutID *int64      `json:"defaultLayoutId"`
	Layouts         []LayoutRef `json:"instanceTypeLayouts"`
}

type LayoutRef struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

// normalize derives DefaultLayoutID from the embedded instanceTypeLayouts when
// the API does not provide a top-level defaultLayoutId.
func (t *InstanceType) normalize() {
	if t.DefaultLayoutID == nil && len(t.Layouts) > 0 {
		id := t.Layouts[0].ID
		t.DefaultLayoutID = &id
	}
}

func (c *Client) GetInstanceTypeByCode(ctx context.Context, code string) (*InstanceType, error) {
	var response struct {
		InstanceTypes []InstanceType `json:"instanceTypes"`
	}
	query := map[string]string{"code": code, "max": "1", "sort": "name", "direction": "asc"}
	if err := c.get(ctx, "/instance-types", query, &response); err != nil {
		return nil, err
	}
	if len(response.InstanceTypes) == 0 {
		return nil, &APIError{StatusCode: 404, Message: fmt.Sprintf("instance type %q not found", code)}
	}
	instanceType := response.InstanceTypes[0]
	instanceType.normalize()
	return &instanceType, nil
}
