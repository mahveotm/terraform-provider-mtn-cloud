package client

import (
	"context"
	"fmt"
)

// ScaleThreshold is an autoscale rule (CPU/memory/disk based).
type ScaleThreshold struct {
	ID             int64    `json:"id"`
	Name           string   `json:"name"`
	AutoUp         *bool    `json:"autoUp"`
	AutoDown       *bool    `json:"autoDown"`
	MinCount       *int64   `json:"minCount"`
	MaxCount       *int64   `json:"maxCount"`
	ScaleIncrement *int64   `json:"scaleIncrement"`
	CPUEnabled     *bool    `json:"cpuEnabled"`
	MinCPU         *float64 `json:"minCpu"`
	MaxCPU         *float64 `json:"maxCpu"`
	MemoryEnabled  *bool    `json:"memoryEnabled"`
	MinMemory      *float64 `json:"minMemory"`
	MaxMemory      *float64 `json:"maxMemory"`
	DiskEnabled    *bool    `json:"diskEnabled"`
	MinDisk        *float64 `json:"minDisk"`
	MaxDisk        *float64 `json:"maxDisk"`
}

// ScaleThresholdInput is the create/update payload.
type ScaleThresholdInput struct {
	Name           string
	AutoUp         *bool
	AutoDown       *bool
	MinCount       *int64
	MaxCount       *int64
	ScaleIncrement *int64
	CPUEnabled     *bool
	MinCPU         *float64
	MaxCPU         *float64
	MemoryEnabled  *bool
	MinMemory      *float64
	MaxMemory      *float64
	DiskEnabled    *bool
	MinDisk        *float64
	MaxDisk        *float64
}

func scaleThresholdBody(input ScaleThresholdInput) map[string]any {
	st := map[string]any{"name": input.Name}
	putBool := func(k string, v *bool) {
		if v != nil {
			st[k] = *v
		}
	}
	putInt := func(k string, v *int64) {
		if v != nil {
			st[k] = *v
		}
	}
	putFloat := func(k string, v *float64) {
		if v != nil {
			st[k] = *v
		}
	}
	putBool("autoUp", input.AutoUp)
	putBool("autoDown", input.AutoDown)
	putInt("minCount", input.MinCount)
	putInt("maxCount", input.MaxCount)
	putInt("scaleIncrement", input.ScaleIncrement)
	putBool("cpuEnabled", input.CPUEnabled)
	putFloat("minCpu", input.MinCPU)
	putFloat("maxCpu", input.MaxCPU)
	putBool("memoryEnabled", input.MemoryEnabled)
	putFloat("minMemory", input.MinMemory)
	putFloat("maxMemory", input.MaxMemory)
	putBool("diskEnabled", input.DiskEnabled)
	putFloat("minDisk", input.MinDisk)
	putFloat("maxDisk", input.MaxDisk)
	return map[string]any{"scaleThreshold": st}
}

func (c *Client) CreateScaleThreshold(ctx context.Context, input ScaleThresholdInput) (*ScaleThreshold, error) {
	return createObj[ScaleThreshold](c, ctx, "/scale-thresholds", "scaleThreshold", scaleThresholdBody(input))
}

func (c *Client) GetScaleThreshold(ctx context.Context, id int64) (*ScaleThreshold, error) {
	return getByID[ScaleThreshold](c, ctx, fmt.Sprintf("/scale-thresholds/%d", id), "scaleThreshold")
}

func (c *Client) GetScaleThresholdByName(ctx context.Context, name string) (*ScaleThreshold, error) {
	return firstByName[ScaleThreshold](c, ctx, "/scale-thresholds", "scaleThresholds", name)
}

func (c *Client) UpdateScaleThreshold(ctx context.Context, id int64, input ScaleThresholdInput) (*ScaleThreshold, error) {
	return updateObj[ScaleThreshold](c, ctx, fmt.Sprintf("/scale-thresholds/%d", id), "scaleThreshold", scaleThresholdBody(input))
}

func (c *Client) DeleteScaleThreshold(ctx context.Context, id int64) error {
	return c.delete(ctx, fmt.Sprintf("/scale-thresholds/%d", id), nil)
}

func (c *Client) ListScaleThresholds(ctx context.Context) ([]ScaleThreshold, error) {
	return listObjects[ScaleThreshold](c, ctx, "/scale-thresholds", "scaleThresholds")
}
