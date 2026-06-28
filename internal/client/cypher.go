package client

import (
	"context"
	"net/http"
	"strconv"
)

// Cypher is an entry in the MTN Cloud (Morpheus) Cypher secret store. The
// plaintext value is never returned on writes; reads decrypt it under `data`.
type Cypher struct {
	ID      int64  `json:"id"`
	ItemKey string `json:"itemKey"`
}

// CypherResult bundles the cypher record with the lease metadata that the API
// returns alongside it.
type CypherResult struct {
	Cypher        Cypher
	Value         string
	LeaseDuration *int64
}

// cypherSecretPath builds the path for a secret-mount key. The caller supplies
// the key without the `secret/` mount prefix.
func cypherSecretPath(key string) string {
	return "/cypher/secret/" + key
}

// CreateCypher writes a string secret to secret/<key>. ttl is optional; when nil
// the backend applies its default lease.
func (c *Client) CreateCypher(ctx context.Context, key, value string, ttl *int64) (*CypherResult, error) {
	query := map[string]string{"type": "string"}
	if ttl != nil {
		query["ttl"] = strconv.FormatInt(*ttl, 10)
	}
	var response struct {
		Cypher        Cypher `json:"cypher"`
		Data          string `json:"data"`
		LeaseDuration *int64 `json:"lease_duration"`
	}
	// post() has no query support, so call do() directly to pass ttl/type.
	if err := c.do(ctx, http.MethodPost, cypherSecretPath(key), query, map[string]any{"value": value}, &response); err != nil {
		return nil, err
	}
	return &CypherResult{Cypher: response.Cypher, Value: response.Data, LeaseDuration: response.LeaseDuration}, nil
}

// GetCypher reads secret/<key>, returning the decrypted value under Value.
func (c *Client) GetCypher(ctx context.Context, key string) (*CypherResult, error) {
	var response struct {
		Cypher        Cypher `json:"cypher"`
		Data          string `json:"data"`
		LeaseDuration *int64 `json:"lease_duration"`
	}
	if err := c.get(ctx, cypherSecretPath(key), nil, &response); err != nil {
		return nil, err
	}
	return &CypherResult{Cypher: response.Cypher, Value: response.Data, LeaseDuration: response.LeaseDuration}, nil
}

func (c *Client) DeleteCypher(ctx context.Context, key string) error {
	return c.delete(ctx, cypherSecretPath(key), nil)
}
