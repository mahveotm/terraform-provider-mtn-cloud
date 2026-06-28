package client

import (
	"context"
	"fmt"
)

// KeyPair is an SSH key pair stored in MTN Cloud. The API only ever returns a
// hash of the private key (never the plaintext), so PrivateKeyHash is what Read
// can reconcile against.
type KeyPair struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	PublicKey      string `json:"publicKey"`
	PrivateKeyHash string `json:"privateKeyHash"`
}

// KeyPairInput is the create payload. PrivateKey/Passphrase are write-only.
type KeyPairInput struct {
	Name       string
	PublicKey  string
	PrivateKey string
	Passphrase string
}

func (c *Client) CreateKeyPair(ctx context.Context, input KeyPairInput) (*KeyPair, error) {
	keyPair := map[string]any{
		"name":      input.Name,
		"publicKey": input.PublicKey,
	}
	if input.PrivateKey != "" {
		keyPair["privateKey"] = input.PrivateKey
	}
	if input.Passphrase != "" {
		keyPair["passphrase"] = input.Passphrase
	}
	return createObj[KeyPair](c, ctx, "/key-pairs", "keyPair", map[string]any{"keyPair": keyPair})
}

func (c *Client) GetKeyPair(ctx context.Context, id int64) (*KeyPair, error) {
	return getByID[KeyPair](c, ctx, fmt.Sprintf("/key-pairs/%d", id), "keyPair")
}

// GetKeyPairByName looks a key pair up by name via the Morpheus collection list
// endpoint (GET /key-pairs?name=). The endpoint is part of the Morpheus API even
// though the curated openapi.yaml only documents create/get-by-id/delete.
func (c *Client) GetKeyPairByName(ctx context.Context, name string) (*KeyPair, error) {
	return firstByName[KeyPair](c, ctx, "/key-pairs", "keyPairs", name)
}

func (c *Client) DeleteKeyPair(ctx context.Context, id int64) error {
	return c.delete(ctx, fmt.Sprintf("/key-pairs/%d", id), nil)
}

func (c *Client) ListKeyPairs(ctx context.Context) ([]KeyPair, error) {
	return listObjects[KeyPair](c, ctx, "/key-pairs", "keyPairs")
}
