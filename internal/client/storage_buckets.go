package client

import (
	"context"
	"fmt"
)

type StorageBucket struct {
	ID                        int64          `json:"id"`
	Name                      string         `json:"name"`
	BucketName                string         `json:"bucketName"`
	ProviderType              string         `json:"providerType"`
	Config                    map[string]any `json:"config"`
	DefaultBackupTarget       *bool          `json:"defaultBackupTarget"`
	CopyToStore               *bool          `json:"copyToStore"`
	DefaultDeploymentTarget   *bool          `json:"defaultDeploymentTarget"`
	DefaultVirtualImageTarget *bool          `json:"defaultVirtualImageTarget"`
	RetentionPolicyType       string         `json:"retentionPolicyType"`
	RetentionPolicyDays       any            `json:"retentionPolicyDays"`
	RetentionProvider         string         `json:"retentionProvider"`
}

type StorageBucketInput struct {
	Name                      string
	BucketName                string
	AccessKey                 string
	SecretKey                 string
	Endpoint                  string
	StorageServer             *int64
	CreateBucket              *bool
	DefaultBackupTarget       *bool
	CopyToStore               *bool
	DefaultDeploymentTarget   *bool
	DefaultVirtualImageTarget *bool
	RetentionPolicyType       string
	RetentionPolicyDays       *int64
	RetentionProvider         string
}

func storageBucketBody(input StorageBucketInput, includeProviderType bool) map[string]any {
	bucket := map[string]any{
		"name":       input.Name,
		"bucketName": input.BucketName,
		"config": map[string]any{
			"accessKey": input.AccessKey,
			"secretKey": input.SecretKey,
			"endpoint":  input.Endpoint,
		},
	}
	if includeProviderType {
		bucket["providerType"] = "s3"
	}
	if input.StorageServer != nil {
		bucket["storageServer"] = *input.StorageServer
	}
	if input.CreateBucket != nil {
		bucket["createBucket"] = *input.CreateBucket
	}
	if input.DefaultBackupTarget != nil {
		bucket["defaultBackupTarget"] = *input.DefaultBackupTarget
	}
	if input.CopyToStore != nil {
		bucket["copyToStore"] = *input.CopyToStore
	}
	if input.DefaultDeploymentTarget != nil {
		bucket["defaultDeploymentTarget"] = *input.DefaultDeploymentTarget
	}
	if input.DefaultVirtualImageTarget != nil {
		bucket["defaultVirtualImageTarget"] = *input.DefaultVirtualImageTarget
	}
	if input.RetentionPolicyType != "" {
		bucket["retentionPolicyType"] = input.RetentionPolicyType
	}
	if input.RetentionPolicyDays != nil {
		bucket["retentionPolicyDays"] = *input.RetentionPolicyDays
	}
	if input.RetentionProvider != "" {
		bucket["retentionProvider"] = input.RetentionProvider
	}
	return map[string]any{"storageBucket": bucket}
}

func (c *Client) CreateStorageBucket(ctx context.Context, input StorageBucketInput) (*StorageBucket, error) {
	var response struct {
		StorageBucket StorageBucket `json:"storageBucket"`
	}
	if err := c.post(ctx, "/storage-buckets", storageBucketBody(input, true), &response); err != nil {
		return nil, err
	}
	return &response.StorageBucket, nil
}

func (c *Client) GetStorageBucket(ctx context.Context, id int64) (*StorageBucket, error) {
	var response struct {
		StorageBucket StorageBucket `json:"storageBucket"`
	}
	if err := c.get(ctx, fmt.Sprintf("/storage-buckets/%d", id), nil, &response); err != nil {
		return nil, err
	}
	return &response.StorageBucket, nil
}

func (c *Client) GetStorageBucketByName(ctx context.Context, name string) (*StorageBucket, error) {
	var response struct {
		StorageBuckets []StorageBucket `json:"storageBuckets"`
	}
	if err := c.get(ctx, "/storage-buckets", map[string]string{"name": name, "max": "1"}, &response); err != nil {
		return nil, err
	}
	if len(response.StorageBuckets) == 0 {
		return nil, &APIError{StatusCode: 404, Message: fmt.Sprintf("storage bucket %q not found", name)}
	}
	return &response.StorageBuckets[0], nil
}

func (c *Client) UpdateStorageBucket(ctx context.Context, id int64, input StorageBucketInput) (*StorageBucket, error) {
	var response struct {
		StorageBucket StorageBucket `json:"storageBucket"`
	}
	if err := c.put(ctx, fmt.Sprintf("/storage-buckets/%d", id), storageBucketBody(input, false), &response); err != nil {
		return nil, err
	}
	return &response.StorageBucket, nil
}

func (c *Client) DeleteStorageBucket(ctx context.Context, id int64, removeResources bool) error {
	var query map[string]string
	if removeResources {
		query = map[string]string{"removeResources": "true"}
	}
	return c.delete(ctx, fmt.Sprintf("/storage-buckets/%d", id), query)
}
