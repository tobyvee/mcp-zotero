package zotero

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSearchItemsBuildsRequest(t *testing.T) {
	var gotPath, gotQuery, gotVersion string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		gotVersion = r.Header.Get("Zotero-API-Version")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"key":"ABCD1234"}]`))
	}))
	defer srv.Close()

	c := New(WithBaseURL(srv.URL+"/api"), WithUserID("0"))
	body, err := c.SearchItems(context.Background(), "deep learning", "everything", 5)
	if err != nil {
		t.Fatalf("SearchItems: %v", err)
	}

	if want := "/api/users/0/items"; gotPath != want {
		t.Errorf("path = %q, want %q", gotPath, want)
	}
	if gotVersion != "3" {
		t.Errorf("Zotero-API-Version = %q, want 3", gotVersion)
	}
	for _, want := range []string{"q=deep+learning", "qmode=everything", "limit=5"} {
		if !containsParam(gotQuery, want) {
			t.Errorf("query %q missing %q", gotQuery, want)
		}
	}
	if string(body) != `[{"key":"ABCD1234"}]` {
		t.Errorf("body passed through incorrectly: %s", body)
	}
}

func TestForbiddenGivesActionableError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	c := New(WithBaseURL(srv.URL + "/api"))
	_, err := c.Collections(context.Background())
	if err == nil {
		t.Fatal("expected error on 403, got nil")
	}
	if !containsParam(err.Error(), "Allow other applications") {
		t.Errorf("403 error should mention the Zotero preference, got: %v", err)
	}
}

func containsParam(haystack, needle string) bool {
	return len(needle) > 0 && stringContains(haystack, needle)
}

func stringContains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
