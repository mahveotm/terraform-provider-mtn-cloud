package provider

import (
	"reflect"
	"testing"
)

func TestMergeLabelsUnionDedupesPreservingOrder(t *testing.T) {
	t.Parallel()

	got := mergeLabels([]string{"terraform", "team"}, []string{"team", "web", ""})
	want := []string{"terraform", "team", "web"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("mergeLabels = %v, want %v", got, want)
	}
}

func TestMergeLabelsEmpty(t *testing.T) {
	t.Parallel()

	if got := mergeLabels(nil, nil); len(got) != 0 {
		t.Fatalf("expected empty, got %v", got)
	}
}

func TestMergeTagsResourceOverridesDefault(t *testing.T) {
	t.Parallel()

	got := mergeTags(
		map[string]string{"env": "dev", "owner": "platform"},
		map[string]string{"env": "prod"},
	)
	want := map[string]string{"env": "prod", "owner": "platform"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("mergeTags = %v, want %v", got, want)
	}
}
