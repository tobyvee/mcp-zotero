// Command mcp-zotero is a Model Context Protocol server that exposes a read-only
// view of the user's local Zotero library. It speaks MCP over stdio and reads
// data from the Zotero desktop app's local HTTP API (localhost:23119), so it
// works offline and needs no API key.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/tobyvee/mcp-zotero/internal/tools"
	"github.com/tobyvee/mcp-zotero/internal/zotero"
)

// version is overridden at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	baseURL := flag.String("base-url", envOr("ZOTERO_API_URL", zotero.DefaultBaseURL),
		"Zotero local API base URL (env ZOTERO_API_URL)")
	userID := flag.String("user-id", envOr("ZOTERO_USER_ID", zotero.DefaultUserID),
		"Zotero library user id; 0 means the locally logged-in user (env ZOTERO_USER_ID)")
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		return
	}

	client := zotero.New(
		zotero.WithBaseURL(*baseURL),
		zotero.WithUserID(*userID),
	)

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "mcp-zotero",
		Version: version,
	}, nil)
	tools.Register(server, client)

	// Log to stderr; stdout is reserved for the MCP protocol.
	log.SetOutput(os.Stderr)
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("mcp-zotero: %v", err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
