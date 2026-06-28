package provider

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mahveotm/terraform-provider-mtncloud/internal/client"
)

// testAccProviderConfig is the provider block for acceptance tests. The provider
// reads MTN_CLOUD_URL/MTN_CLOUD_TOKEN (etc.) from the environment, so the block
// itself is empty.
const testAccProviderConfig = `provider "mtncloud" {}` + "\n"

// accNamePrefix is applied to every object an acceptance test creates so the
// sweeper can find and delete leftovers.
const accNamePrefix = "tf-acc-"

// accName returns a unique, sweepable name for an acceptance-test object.
func accName(kind string) string {
	return fmt.Sprintf("%s%s-%d", accNamePrefix, kind, time.Now().UnixNano())
}

// isSweepable reports whether a name was created by an acceptance test or the
// live smoke harness, and is therefore safe to delete during a sweep.
func isSweepable(name string) bool {
	return strings.HasPrefix(name, accNamePrefix) || strings.HasPrefix(name, "tf-smoke-")
}

// sweepClient builds an API client from the same MTN_CLOUD_* environment the
// provider uses, for sweepers (which run outside the provider lifecycle).
func sweepClient() (*client.Client, error) {
	url := os.Getenv("MTN_CLOUD_URL")
	if url == "" {
		url = client.DefaultURL
	}
	return client.New(client.Config{
		URL:      url,
		Token:    os.Getenv("MTN_CLOUD_TOKEN"),
		Username: os.Getenv("MTN_CLOUD_USERNAME"),
		Password: os.Getenv("MTN_CLOUD_PASSWORD"),
		Timeout:  30 * time.Second,
	})
}
