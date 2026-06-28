package client

import (
	"context"
	"encoding/json"
	"fmt"
)

// Generic envelope helpers. The MTN Cloud (Morpheus) API wraps single objects in
// a named key (e.g. {"credential": {...}}) and collections in a plural key
// (e.g. {"credentials": [...]}). These helpers decode that envelope once so the
// per-area client files reduce to type definitions, payload builders, and thin
// method wrappers. They build on the low-level c.get/post/put in client.go.

// decodeEnvelope pulls a single named object out of a JSON envelope.
func decodeEnvelope[T any](env map[string]json.RawMessage, key string) (*T, error) {
	raw, ok := env[key]
	if !ok {
		return nil, &APIError{StatusCode: 500, Message: fmt.Sprintf("response missing %q key", key)}
	}
	var out T
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// getByID fetches GET {path} and returns the object under the single-object key.
func getByID[T any](c *Client, ctx context.Context, path, key string) (*T, error) {
	var env map[string]json.RawMessage
	if err := c.get(ctx, path, nil, &env); err != nil {
		return nil, err
	}
	return decodeEnvelope[T](env, key)
}

// firstByName fetches GET {path}?name=&max=1 and returns the first element of the
// collection under listKey, or a 404 APIError if none match.
func firstByName[T any](c *Client, ctx context.Context, path, listKey, name string) (*T, error) {
	var env map[string]json.RawMessage
	if err := c.get(ctx, path, map[string]string{"name": name, "max": "1"}, &env); err != nil {
		return nil, err
	}
	raw, ok := env[listKey]
	if !ok {
		return nil, &APIError{StatusCode: 404, Message: fmt.Sprintf("%s %q not found", listKey, name)}
	}
	var list []T
	if err := json.Unmarshal(raw, &list); err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, &APIError{StatusCode: 404, Message: fmt.Sprintf("%s %q not found", listKey, name)}
	}
	return &list[0], nil
}

// listObjects fetches a whole collection under listKey (used by test sweepers).
func listObjects[T any](c *Client, ctx context.Context, path, listKey string) ([]T, error) {
	var env map[string]json.RawMessage
	if err := c.get(ctx, path, map[string]string{"max": "500"}, &env); err != nil {
		return nil, err
	}
	raw, ok := env[listKey]
	if !ok {
		return nil, nil
	}
	var list []T
	if err := json.Unmarshal(raw, &list); err != nil {
		return nil, err
	}
	return list, nil
}

// createObj POSTs body to {path} and returns the created object under key.
func createObj[T any](c *Client, ctx context.Context, path, key string, body any) (*T, error) {
	var env map[string]json.RawMessage
	if err := c.post(ctx, path, body, &env); err != nil {
		return nil, err
	}
	return decodeEnvelope[T](env, key)
}

// updateObj PUTs body to {path} and returns the updated object under key.
func updateObj[T any](c *Client, ctx context.Context, path, key string, body any) (*T, error) {
	var env map[string]json.RawMessage
	if err := c.put(ctx, path, body, &env); err != nil {
		return nil, err
	}
	return decodeEnvelope[T](env, key)
}
