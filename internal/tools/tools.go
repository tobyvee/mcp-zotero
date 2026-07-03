// Package tools registers the Zotero MCP tools on an mcp.Server. Each tool is a
// thin wrapper over a zotero.Client method: it takes typed input, calls the
// local API, and returns the raw JSON body as text content so the model sees
// exactly what Zotero returned. Because the local API is read-only, every tool
// here is read-only too.
package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/tobyvee/mcp-zotero/internal/zotero"
)

// Register adds every Zotero tool to server, backed by client.
func Register(server *mcp.Server, client *zotero.Client) {
	mcp.AddTool(server, &mcp.Tool{
		Name: "zotero_search_items",
		Description: "Search the local Zotero library. Matches titles, creators, and " +
			"year by default; set mode to \"everything\" to also search full text and notes.",
	}, searchItems(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "zotero_list_items",
		Description: "List top-level items in the Zotero library, most recently modified first.",
	}, listItems(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "zotero_get_item",
		Description: "Fetch a single Zotero item (and optionally its child notes/attachments) by item key.",
	}, getItem(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "zotero_list_collections",
		Description: "List all collections (folders) in the Zotero library.",
	}, listCollections(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "zotero_collection_items",
		Description: "List the items contained in a Zotero collection, given its collection key.",
	}, collectionItems(client))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "zotero_list_tags",
		Description: "List all tags used in the Zotero library.",
	}, listTags(client))
}

// SearchInput is the input for zotero_search_items.
type SearchInput struct {
	Query string `json:"query" jsonschema:"the search text (e.g. an author, title, or keyword)"`
	Mode  string `json:"mode,omitempty" jsonschema:"search mode: 'titleCreatorYear' (default) or 'everything' to include full text and notes"`
	Limit int    `json:"limit,omitempty" jsonschema:"maximum number of items to return (default: no limit)"`
}

func searchItems(c *zotero.Client) mcp.ToolHandlerFor[SearchInput, any] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in SearchInput) (*mcp.CallToolResult, any, error) {
		return result(c.SearchItems(ctx, in.Query, in.Mode, in.Limit))
	}
}

// ListItemsInput is the input for zotero_list_items.
type ListItemsInput struct {
	Limit int `json:"limit,omitempty" jsonschema:"maximum number of items to return (default: no limit)"`
	Start int `json:"start,omitempty" jsonschema:"index of the first item to return, for pagination"`
}

func listItems(c *zotero.Client) mcp.ToolHandlerFor[ListItemsInput, any] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in ListItemsInput) (*mcp.CallToolResult, any, error) {
		return result(c.TopItems(ctx, in.Limit, in.Start))
	}
}

// GetItemInput is the input for zotero_get_item.
type GetItemInput struct {
	ItemKey  string `json:"itemKey" jsonschema:"the 8-character Zotero item key"`
	Children bool   `json:"children,omitempty" jsonschema:"if true, return the item's child notes and attachments instead of the item itself"`
}

func getItem(c *zotero.Client) mcp.ToolHandlerFor[GetItemInput, any] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in GetItemInput) (*mcp.CallToolResult, any, error) {
		if in.Children {
			return result(c.ItemChildren(ctx, in.ItemKey))
		}
		return result(c.Item(ctx, in.ItemKey))
	}
}

// CollectionItemsInput is the input for zotero_collection_items.
type CollectionItemsInput struct {
	CollectionKey string `json:"collectionKey" jsonschema:"the 8-character Zotero collection key"`
	Limit         int    `json:"limit,omitempty" jsonschema:"maximum number of items to return (default: no limit)"`
	Start         int    `json:"start,omitempty" jsonschema:"index of the first item to return, for pagination"`
}

func collectionItems(c *zotero.Client) mcp.ToolHandlerFor[CollectionItemsInput, any] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, in CollectionItemsInput) (*mcp.CallToolResult, any, error) {
		return result(c.CollectionItems(ctx, in.CollectionKey, in.Limit, in.Start))
	}
}

func listCollections(c *zotero.Client) mcp.ToolHandlerFor[struct{}, any] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		return result(c.Collections(ctx))
	}
}

func listTags(c *zotero.Client) mcp.ToolHandlerFor[struct{}, any] {
	return func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		return result(c.Tags(ctx))
	}
}

// result adapts a (body, err) client call into a tool result. A client error is
// surfaced to the model as an error result (IsError) rather than a protocol
// error, so the model can read the message and, e.g., prompt the user to start
// Zotero or enable the local API.
func result(body []byte, err error) (*mcp.CallToolResult, any, error) {
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
		}, nil, nil
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(body)}},
	}, nil, nil
}
