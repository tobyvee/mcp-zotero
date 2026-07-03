# mcp-zotero

A [Model Context Protocol](https://modelcontextprotocol.io) server that gives an
MCP client (Claude Desktop, Claude Code, etc.) read-only access to your **local**
Zotero library. It talks to the Zotero desktop app's [local API][local-api] on
`localhost:23119` — so it works offline, needs no API key, and never sends your
library to the cloud.

Built in Go with the [official MCP Go SDK][go-sdk].

## Prerequisites

- **Zotero 7 or newer**, running.
- The local API enabled: **Settings → Advanced → "Allow other applications on
  this computer to communicate with Zotero"**. Without this, requests return
  `403 Forbidden`.
- **Go 1.25+** to build (or use the included devcontainer).

## Build

```sh
make build      # -> ./bin/mcp-zotero (host platform)
make cross      # -> ./bin/mcp-zotero-<os>-<arch> for all platforms
make test       # run tests
make help       # list all targets
```

## Tools

All tools are read-only and return Zotero's JSON verbatim:

| Tool                      | Description                                        |
| ------------------------- | -------------------------------------------------- |
| `zotero_search_items`     | Search the library (title/creator/year or full text) |
| `zotero_list_items`       | List top-level items                               |
| `zotero_get_item`         | Fetch one item (or its child notes/attachments)    |
| `zotero_list_collections` | List all collections                               |
| `zotero_collection_items` | List items in a collection                         |
| `zotero_list_tags`        | List all tags                                      |

## Configuring an MCP client

Point your client at the built binary. For Claude Desktop
(`claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "zotero": {
      "command": "/absolute/path/to/bin/mcp-zotero"
    }
  }
}
```

### Configuration

| Env var          | Flag         | Default                        |
| ---------------- | ------------ | ------------------------------ |
| `ZOTERO_API_URL` | `--base-url` | `http://localhost:23119/api`   |
| `ZOTERO_USER_ID` | `--user-id`  | `0` (the logged-in user)       |

[local-api]: https://www.zotero.org/support/dev/web_api/v3/local_api
[go-sdk]: https://github.com/modelcontextprotocol/go-sdk
