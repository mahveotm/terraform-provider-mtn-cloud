package client

import (
	"context"
	"fmt"
	"strings"
)

// WikiPage is a documentation/wiki page in MTN Cloud.
type WikiPage struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
	Content  string `json:"content"`
}

// WikiPageInput is the create/update payload.
type WikiPageInput struct {
	Name     string
	Category string
	Content  string
}

func wikiPageBody(input WikiPageInput) map[string]any {
	page := map[string]any{
		"name": input.Name,
		// Trim a single trailing newline to match the value the API stores,
		// avoiding a perpetual diff on round-trip.
		"content": strings.TrimSuffix(input.Content, "\n"),
	}
	if input.Category != "" {
		page["category"] = input.Category
	}
	return map[string]any{"page": page}
}

func (c *Client) CreateWikiPage(ctx context.Context, input WikiPageInput) (*WikiPage, error) {
	return createObj[WikiPage](c, ctx, "/wiki/pages", "page", wikiPageBody(input))
}

func (c *Client) GetWikiPage(ctx context.Context, id int64) (*WikiPage, error) {
	return getByID[WikiPage](c, ctx, fmt.Sprintf("/wiki/pages/%d", id), "page")
}

func (c *Client) GetWikiPageByName(ctx context.Context, name string) (*WikiPage, error) {
	return firstByName[WikiPage](c, ctx, "/wiki/pages", "pages", name)
}

func (c *Client) UpdateWikiPage(ctx context.Context, id int64, input WikiPageInput) (*WikiPage, error) {
	return updateObj[WikiPage](c, ctx, fmt.Sprintf("/wiki/pages/%d", id), "page", wikiPageBody(input))
}

func (c *Client) DeleteWikiPage(ctx context.Context, id int64) error {
	return c.delete(ctx, fmt.Sprintf("/wiki/pages/%d", id), nil)
}

func (c *Client) ListWikiPages(ctx context.Context) ([]WikiPage, error) {
	return listObjects[WikiPage](c, ctx, "/wiki/pages", "pages")
}
