package client

import (
	"context"
	"fmt"
)

// Credential is an entry in the MTN Cloud credential store. Secret material is
// never returned by the API (masked), so only the non-secret metadata round-trips.
type Credential struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     *bool  `json:"enabled"`
	Username    string `json:"username"`
	Type        struct {
		Code string `json:"code"`
		Name string `json:"name"`
	} `json:"type"`
}

// CredentialInput carries the union of fields across credential types. Which ones
// are used depends on Type; mapCredentialPayload selects them.
type CredentialInput struct {
	Type         string
	Name         string
	Description  string
	Enabled      *bool
	AccessKey    string
	SecretKey    string
	Username     string
	Password     string
	ClientID     string
	ClientSecret string
	Email        string
	APIKey       string
	Tenant       string
	KeyPairID    *int64
}

// CredentialTypes accepted by the API (see GET /credential-types).
var CredentialTypes = []string{
	"access-key-secret", "api-key", "client-id-secret", "email-private-key",
	"tenant-username-keypair", "username-password", "username-api-key",
	"username-keypair", "username-password-keypair",
}

func mapCredentialPayload(input CredentialInput) map[string]any {
	cred := map[string]any{
		"type": input.Type,
		"name": input.Name,
	}
	if input.Description != "" {
		cred["description"] = input.Description
	}
	if input.Enabled != nil {
		cred["enabled"] = *input.Enabled
	}
	setKeyPair := func() {
		if input.KeyPairID != nil {
			cred["keyPair"] = map[string]any{"id": *input.KeyPairID}
		}
	}
	switch input.Type {
	case "access-key-secret":
		cred["username"] = input.AccessKey
		cred["password"] = input.SecretKey
	case "api-key":
		cred["password"] = input.APIKey
	case "client-id-secret":
		cred["username"] = input.ClientID
		cred["password"] = input.ClientSecret
	case "email-private-key":
		cred["username"] = input.Email
		setKeyPair()
	case "tenant-username-keypair":
		cred["authPath"] = input.Tenant
		cred["username"] = input.Username
		setKeyPair()
	case "username-api-key":
		cred["username"] = input.Username
		cred["password"] = input.APIKey
	case "username-keypair":
		cred["username"] = input.Username
		setKeyPair()
	case "username-password":
		cred["username"] = input.Username
		cred["password"] = input.Password
	case "username-password-keypair":
		cred["username"] = input.Username
		cred["password"] = input.Password
		setKeyPair()
	}
	return map[string]any{"credential": cred}
}

func (c *Client) CreateCredential(ctx context.Context, input CredentialInput) (*Credential, error) {
	return createObj[Credential](c, ctx, "/credentials", "credential", mapCredentialPayload(input))
}

func (c *Client) GetCredential(ctx context.Context, id int64) (*Credential, error) {
	return getByID[Credential](c, ctx, fmt.Sprintf("/credentials/%d", id), "credential")
}

func (c *Client) GetCredentialByName(ctx context.Context, name string) (*Credential, error) {
	return firstByName[Credential](c, ctx, "/credentials", "credentials", name)
}

func (c *Client) UpdateCredential(ctx context.Context, id int64, input CredentialInput) (*Credential, error) {
	return updateObj[Credential](c, ctx, fmt.Sprintf("/credentials/%d", id), "credential", mapCredentialPayload(input))
}

func (c *Client) DeleteCredential(ctx context.Context, id int64) error {
	return c.delete(ctx, fmt.Sprintf("/credentials/%d", id), nil)
}

func (c *Client) ListCredentials(ctx context.Context) ([]Credential, error) {
	return listObjects[Credential](c, ctx, "/credentials", "credentials")
}
