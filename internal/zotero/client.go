// Package zotero is a thin read-only client for Zotero's *local* HTTP API,
// which the Zotero 7+ desktop app exposes on the loopback interface. It mirrors
// the Zotero Web API v3 but serves the user's local database, requires no
// authentication, and accepts only GET requests.
//
// See https://www.zotero.org/support/dev/web_api/v3/local_api
package zotero

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	// DefaultBaseURL is the local API endpoint exposed by the Zotero desktop
	// app. It must be enabled via Settings → Advanced → "Allow other
	// applications on this computer to communicate with Zotero".
	DefaultBaseURL = "http://localhost:23119/api"

	// DefaultUserID selects the locally logged-in user's library. The local
	// API accepts 0 as an alias for the current user.
	DefaultUserID = "0"

	// apiVersion pins the Zotero API version. The local API serves one version
	// at a time and currently supports v3.
	apiVersion = "3"
)

// Client talks to a single Zotero library over the local API.
type Client struct {
	baseURL string
	userID  string
	http    *http.Client
}

// Option configures a Client.
type Option func(*Client)

// WithBaseURL overrides the local API base URL (default DefaultBaseURL). Useful
// when Zotero runs on a different host — e.g. from inside a devcontainer, where
// the host is reachable as http://host.docker.internal:23119/api.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		if baseURL != "" {
			c.baseURL = strings.TrimRight(baseURL, "/")
		}
	}
}

// WithUserID overrides the library owner id (default DefaultUserID).
func WithUserID(userID string) Option {
	return func(c *Client) {
		if userID != "" {
			c.userID = userID
		}
	}
}

// WithHTTPClient supplies a custom *http.Client (e.g. for tests).
func WithHTTPClient(h *http.Client) Option {
	return func(c *Client) {
		if h != nil {
			c.http = h
		}
	}
}

// New builds a Client with sensible defaults for a local Zotero instance.
func New(opts ...Option) *Client {
	c := &Client{
		baseURL: DefaultBaseURL,
		userID:  DefaultUserID,
		http:    &http.Client{Timeout: 30 * time.Second},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// get issues a GET against <baseURL>/users/<userID><path> and returns the raw
// response body. The body is passed through to callers verbatim so tools can
// hand Zotero's JSON straight to the model without a lossy round-trip.
func (c *Client) get(ctx context.Context, path string, query url.Values) ([]byte, error) {
	u := fmt.Sprintf("%s/users/%s%s", c.baseURL, c.userID, path)
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Zotero-API-Version", apiVersion)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("zotero local API unreachable at %s: %w "+
			"(is the Zotero desktop app running?)", c.baseURL, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return body, nil
	case http.StatusForbidden:
		return nil, fmt.Errorf("zotero local API returned 403 Forbidden: enable "+
			"Settings → Advanced → \"Allow other applications on this computer to "+
			"communicate with Zotero\" in the desktop app (%s)", u)
	default:
		return nil, fmt.Errorf("zotero local API GET %s: %s: %s", u, resp.Status, strings.TrimSpace(string(body)))
	}
}

// SearchItems runs a quick search across the library. mode is the Zotero
// "qmode" ("titleCreatorYear" or "everything"); limit caps results (<=0 leaves
// it unset, and the local API imposes no default limit).
func (c *Client) SearchItems(ctx context.Context, q, mode string, limit int) ([]byte, error) {
	query := url.Values{}
	query.Set("q", q)
	if mode == "" {
		mode = "titleCreatorYear"
	}
	query.Set("qmode", mode)
	setLimit(query, limit)
	return c.get(ctx, "/items", query)
}

// TopItems lists top-level items (excluding child notes/attachments).
func (c *Client) TopItems(ctx context.Context, limit, start int) ([]byte, error) {
	query := url.Values{}
	setLimit(query, limit)
	if start > 0 {
		query.Set("start", strconv.Itoa(start))
	}
	return c.get(ctx, "/items/top", query)
}

// Item fetches a single item by its 8-character key.
func (c *Client) Item(ctx context.Context, key string) ([]byte, error) {
	return c.get(ctx, "/items/"+url.PathEscape(key), nil)
}

// ItemChildren fetches the child notes and attachments of an item.
func (c *Client) ItemChildren(ctx context.Context, key string) ([]byte, error) {
	return c.get(ctx, "/items/"+url.PathEscape(key)+"/children", nil)
}

// Collections lists all collections in the library.
func (c *Client) Collections(ctx context.Context) ([]byte, error) {
	return c.get(ctx, "/collections", nil)
}

// CollectionItems lists the items in a collection.
func (c *Client) CollectionItems(ctx context.Context, key string, limit, start int) ([]byte, error) {
	query := url.Values{}
	setLimit(query, limit)
	if start > 0 {
		query.Set("start", strconv.Itoa(start))
	}
	return c.get(ctx, "/collections/"+url.PathEscape(key)+"/items", query)
}

// Tags lists tags in the library.
func (c *Client) Tags(ctx context.Context) ([]byte, error) {
	return c.get(ctx, "/tags", nil)
}

func setLimit(query url.Values, limit int) {
	if limit > 0 {
		query.Set("limit", strconv.Itoa(limit))
	}
}
