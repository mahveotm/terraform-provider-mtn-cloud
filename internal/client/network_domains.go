package client

import (
	"context"
	"fmt"
)

// NetworkDomain is a DNS / Active Directory domain in MTN Cloud.
type NetworkDomain struct {
	ID               int64  `json:"id"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	FQDN             string `json:"fqdn"`
	Visibility       string `json:"visibility"`
	Active           *bool  `json:"active"`
	PublicZone       *bool  `json:"publicZone"`
	DomainController *bool  `json:"domainController"`
	DomainUsername   string `json:"domainUsername"`
}

// NetworkDomainInput is the create/update payload.
type NetworkDomainInput struct {
	Name             string
	Description      string
	FQDN             string
	Visibility       string
	Active           *bool
	PublicZone       *bool
	DomainController *bool
	DomainUsername   string
	DomainPassword   string
}

func networkDomainBody(input NetworkDomainInput) map[string]any {
	domain := map[string]any{"name": input.Name}
	if input.Description != "" {
		domain["description"] = input.Description
	}
	if input.FQDN != "" {
		domain["fqdn"] = input.FQDN
	}
	if input.Visibility != "" {
		domain["visibility"] = input.Visibility
	}
	if input.Active != nil {
		domain["active"] = *input.Active
	}
	if input.PublicZone != nil {
		domain["publicZone"] = *input.PublicZone
	}
	if input.DomainController != nil {
		domain["domainController"] = *input.DomainController
	}
	if input.DomainUsername != "" {
		domain["domainUsername"] = input.DomainUsername
	}
	if input.DomainPassword != "" {
		domain["domainPassword"] = input.DomainPassword
	}
	return map[string]any{"networkDomain": domain}
}

func (c *Client) CreateNetworkDomain(ctx context.Context, input NetworkDomainInput) (*NetworkDomain, error) {
	return createObj[NetworkDomain](c, ctx, "/networks/domains", "networkDomain", networkDomainBody(input))
}

func (c *Client) GetNetworkDomain(ctx context.Context, id int64) (*NetworkDomain, error) {
	return getByID[NetworkDomain](c, ctx, fmt.Sprintf("/networks/domains/%d", id), "networkDomain")
}

func (c *Client) GetNetworkDomainByName(ctx context.Context, name string) (*NetworkDomain, error) {
	return firstByName[NetworkDomain](c, ctx, "/networks/domains", "networkDomains", name)
}

func (c *Client) UpdateNetworkDomain(ctx context.Context, id int64, input NetworkDomainInput) (*NetworkDomain, error) {
	return updateObj[NetworkDomain](c, ctx, fmt.Sprintf("/networks/domains/%d", id), "networkDomain", networkDomainBody(input))
}

func (c *Client) DeleteNetworkDomain(ctx context.Context, id int64) error {
	return c.delete(ctx, fmt.Sprintf("/networks/domains/%d", id), nil)
}

func (c *Client) ListNetworkDomains(ctx context.Context) ([]NetworkDomain, error) {
	return listObjects[NetworkDomain](c, ctx, "/networks/domains", "networkDomains")
}
