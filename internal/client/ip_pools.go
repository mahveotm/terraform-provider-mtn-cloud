package client

import (
	"context"
	"fmt"
)

// IPRange is one contiguous range within an IP pool.
type IPRange struct {
	ID           int64  `json:"id"`
	StartAddress string `json:"startAddress"`
	EndAddress   string `json:"endAddress"`
}

// IPPool is a Morpheus-managed IPv4 address pool.
type IPPool struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Gateway   string    `json:"gateway"`
	Netmask   string    `json:"netmask"`
	DNSDomain string    `json:"dnsDomain"`
	IPRanges  []IPRange `json:"ipRanges"`
}

// IPPoolInput is the create/update payload.
type IPPoolInput struct {
	Name       string
	Gateway    string
	Netmask    string
	DNSDomain  string
	DNSServers []string
	IPRanges   []IPRange
}

func ipPoolBody(input IPPoolInput) map[string]any {
	ranges := make([]map[string]any, 0, len(input.IPRanges))
	for _, r := range input.IPRanges {
		ranges = append(ranges, map[string]any{"startAddress": r.StartAddress, "endAddress": r.EndAddress})
	}
	pool := map[string]any{
		"name":     input.Name,
		"type":     "morpheus",
		"ipRanges": ranges,
	}
	if input.Gateway != "" {
		pool["gateway"] = input.Gateway
	}
	if input.Netmask != "" {
		pool["netmask"] = input.Netmask
	}
	if input.DNSDomain != "" {
		pool["dnsDomain"] = input.DNSDomain
	}
	if len(input.DNSServers) > 0 {
		pool["dnsServers"] = input.DNSServers
	}
	return map[string]any{"networkPool": pool}
}

func (c *Client) CreateIPPool(ctx context.Context, input IPPoolInput) (*IPPool, error) {
	return createObj[IPPool](c, ctx, "/networks/pools", "networkPool", ipPoolBody(input))
}

func (c *Client) GetIPPool(ctx context.Context, id int64) (*IPPool, error) {
	return getByID[IPPool](c, ctx, fmt.Sprintf("/networks/pools/%d", id), "networkPool")
}

func (c *Client) GetIPPoolByName(ctx context.Context, name string) (*IPPool, error) {
	return firstByName[IPPool](c, ctx, "/networks/pools", "networkPools", name)
}

func (c *Client) UpdateIPPool(ctx context.Context, id int64, input IPPoolInput) (*IPPool, error) {
	return updateObj[IPPool](c, ctx, fmt.Sprintf("/networks/pools/%d", id), "networkPool", ipPoolBody(input))
}

func (c *Client) DeleteIPPool(ctx context.Context, id int64) error {
	return c.delete(ctx, fmt.Sprintf("/networks/pools/%d", id), nil)
}

func (c *Client) ListIPPools(ctx context.Context) ([]IPPool, error) {
	return listObjects[IPPool](c, ctx, "/networks/pools", "networkPools")
}
