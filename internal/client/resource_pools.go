package client

import (
	"context"
	"fmt"
	"strconv"
)

type ResourcePool struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

func (c *Client) ListResourcePools(ctx context.Context, cloudID, groupID int64) ([]ResourcePool, error) {
	var response struct {
		Data []struct {
			ID   *int64 `json:"id"`
			Name string `json:"name"`
			// The option-source returns the pool code under "value"
			// (e.g. "pool-214"); "code" is accepted as a fallback.
			Value   string `json:"value"`
			Code    string `json:"code"`
			IsGroup bool   `json:"isGroup"`
		} `json:"data"`
	}
	query := map[string]string{
		"cloudId":           strconv.FormatInt(cloudID, 10),
		"groupId":           strconv.FormatInt(groupID, 10),
		"provisionTypeCode": "openstack",
	}
	if err := c.get(ctx, "/options/zonePools", query, &response); err != nil {
		return nil, err
	}
	pools := make([]ResourcePool, 0, len(response.Data))
	for _, item := range response.Data {
		if item.IsGroup || item.ID == nil {
			continue
		}
		code := item.Code
		if code == "" {
			code = item.Value
		}
		pools = append(pools, ResourcePool{ID: *item.ID, Name: item.Name, Code: code})
	}
	return pools, nil
}

func (c *Client) GetResourcePool(ctx context.Context, name string, group *Group) (*ResourcePool, error) {
	if len(group.CloudIDs) == 0 {
		return nil, &APIError{StatusCode: 400, Message: fmt.Sprintf("group %q has no associated cloud", group.Name)}
	}
	pools, err := c.ListResourcePools(ctx, group.CloudIDs[0], group.ID)
	if err != nil {
		return nil, err
	}
	for _, pool := range pools {
		if pool.Name == name || pool.Code == name {
			return &pool, nil
		}
	}
	return nil, &APIError{StatusCode: 404, Message: fmt.Sprintf("resource pool %q not found", name)}
}
