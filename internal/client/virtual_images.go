package client

import (
	"context"
	"fmt"
)

type VirtualImage struct {
	ID        int64          `json:"id"`
	Name      string         `json:"name"`
	Code      string         `json:"code"`
	ImageType string         `json:"imageType"`
	OsType    map[string]any `json:"osType"`
	IsPublic  *bool          `json:"isPublic"`
}

func (c *Client) GetVirtualImageByName(ctx context.Context, name string) (*VirtualImage, error) {
	var response struct {
		VirtualImages []VirtualImage `json:"virtualImages"`
	}
	if err := c.get(ctx, "/virtual-images", map[string]string{"name": name, "max": "1"}, &response); err != nil {
		return nil, err
	}
	if len(response.VirtualImages) == 0 {
		return nil, &APIError{StatusCode: 404, Message: fmt.Sprintf("virtual image %q not found", name)}
	}
	return &response.VirtualImages[0], nil
}

func (c *Client) GetVirtualImage(ctx context.Context, id int64) (*VirtualImage, error) {
	var response struct {
		VirtualImage VirtualImage `json:"virtualImage"`
	}
	if err := c.get(ctx, fmt.Sprintf("/virtual-images/%d", id), nil, &response); err != nil {
		return nil, err
	}
	return &response.VirtualImage, nil
}
