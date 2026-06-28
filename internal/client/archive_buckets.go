package client

import (
	"context"
	"fmt"
)

type ArchiveBucket struct {
	ID              int64          `json:"id"`
	Name            string         `json:"name"`
	Description     string         `json:"description"`
	Code            string         `json:"code"`
	Visibility      string         `json:"visibility"`
	IsPublic        *bool          `json:"isPublic"`
	StorageProvider map[string]any `json:"storageProvider"`
	FileCount       *int64         `json:"fileCount"`
	RawSize         *int64         `json:"rawSize"`
}

type ArchiveBucketInput struct {
	Name              string
	StorageProviderID int64
	Description       string
	Visibility        string
	IsPublic          *bool
	AccountID         *int64
}

func archiveBucketBody(input ArchiveBucketInput, includeProvider bool) map[string]any {
	bucket := map[string]any{"name": input.Name}
	if includeProvider {
		bucket["storageProvider"] = map[string]any{"id": input.StorageProviderID}
	}
	if input.Description != "" {
		bucket["description"] = input.Description
	}
	if input.Visibility != "" {
		bucket["visibility"] = input.Visibility
	}
	if input.IsPublic != nil {
		bucket["isPublic"] = *input.IsPublic
	}
	if input.AccountID != nil {
		bucket["accounts"] = map[string]any{"id": *input.AccountID}
	}
	return map[string]any{"archiveBucket": bucket}
}

func (c *Client) CreateArchiveBucket(ctx context.Context, input ArchiveBucketInput) (*ArchiveBucket, error) {
	var response struct {
		ArchiveBucket ArchiveBucket `json:"archiveBucket"`
	}
	if err := c.post(ctx, "/archives/buckets", archiveBucketBody(input, true), &response); err != nil {
		return nil, err
	}
	return &response.ArchiveBucket, nil
}

func (c *Client) GetArchiveBucket(ctx context.Context, id int64) (*ArchiveBucket, error) {
	var response struct {
		ArchiveBucket ArchiveBucket `json:"archiveBucket"`
	}
	if err := c.get(ctx, fmt.Sprintf("/archives/buckets/%d", id), nil, &response); err != nil {
		return nil, err
	}
	return &response.ArchiveBucket, nil
}

func (c *Client) GetArchiveBucketByName(ctx context.Context, name string) (*ArchiveBucket, error) {
	var response struct {
		ArchiveBuckets []ArchiveBucket `json:"archiveBuckets"`
	}
	if err := c.get(ctx, "/archives/buckets", map[string]string{"name": name, "max": "1"}, &response); err != nil {
		return nil, err
	}
	if len(response.ArchiveBuckets) == 0 {
		return nil, &APIError{StatusCode: 404, Message: fmt.Sprintf("archive bucket %q not found", name)}
	}
	return &response.ArchiveBuckets[0], nil
}

func (c *Client) UpdateArchiveBucket(ctx context.Context, id int64, input ArchiveBucketInput) (*ArchiveBucket, error) {
	var response struct {
		ArchiveBucket ArchiveBucket `json:"archiveBucket"`
	}
	if err := c.put(ctx, fmt.Sprintf("/archives/buckets/%d", id), archiveBucketBody(input, false), &response); err != nil {
		return nil, err
	}
	return &response.ArchiveBucket, nil
}

func (c *Client) DeleteArchiveBucket(ctx context.Context, id int64) error {
	return c.delete(ctx, fmt.Sprintf("/archives/buckets/%d", id), nil)
}
