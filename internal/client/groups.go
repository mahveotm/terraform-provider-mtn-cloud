package client

import (
	"context"
	"fmt"
)

type Group struct {
	ID       int64     `json:"id"`
	Name     string    `json:"name"`
	Location string    `json:"location"`
	Active   *bool     `json:"active"`
	CloudIDs []int64   `json:"cloudIds"`
	Zones    []ZoneRef `json:"zones"`
}

type ZoneRef struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// normalize populates CloudIDs from the embedded zones array. The MTN Cloud API
// returns a group's clouds under "zones" (e.g. [{id:4}]) rather than a flat
// "cloudIds" list, so this keeps callers using group.CloudIDs working.
func (g *Group) normalize() {
	if len(g.CloudIDs) == 0 {
		for _, zone := range g.Zones {
			g.CloudIDs = append(g.CloudIDs, zone.ID)
		}
	}
}

func (c *Client) GetGroupByName(ctx context.Context, name string) (*Group, error) {
	var response struct {
		Groups []Group `json:"groups"`
	}
	if err := c.get(ctx, "/groups", map[string]string{"name": name, "max": "1"}, &response); err != nil {
		return nil, err
	}
	if len(response.Groups) == 0 {
		return nil, &APIError{StatusCode: 404, Message: fmt.Sprintf("group %q not found", name)}
	}
	group := response.Groups[0]
	group.normalize()
	return &group, nil
}

func (c *Client) GetGroup(ctx context.Context, id int64) (*Group, error) {
	var response struct {
		Group Group `json:"group"`
	}
	if err := c.get(ctx, fmt.Sprintf("/groups/%d", id), nil, &response); err != nil {
		return nil, err
	}
	response.Group.normalize()
	return &response.Group, nil
}
