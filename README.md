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

## Install

### Homebrew (macOS / Linux)

```sh
brew install tobyvee/tap/mcp-zotero
```

### Prebuilt binary

Download the binary for your platform from the [latest release][releases], make
it executable, and put it on your `PATH`:

```sh
curl -fL -o mcp-zotero \
  https://github.com/tobyvee/mcp-zotero/releases/latest/download/mcp-zotero-darwin-arm64
chmod +x mcp-zotero && mv mcp-zotero ~/.local/bin/
```

Swap `darwin-arm64` for `darwin-amd64`, `linux-arm64`, `linux-amd64`, or
`windows-amd64.exe` as needed.

### From source

Requires **Go 1.25+** (or the included devcontainer):

```sh
make build      # -> ./bin/mcp-zotero
```

## Development

```sh
make build      # host binary  -> ./bin/mcp-zotero
make cross      # all platforms -> ./bin/mcp-zotero-<os>-<arch>
make test       # go test -race ./...
make lint       # vet + gofmt + golangci-lint
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

## License

[MIT](LICENSE)

[local-api]: https://www.zotero.org/support/dev/web_api/v3/local_api
[go-sdk]: https://github.com/modelcontextprotocol/go-sdk
[releases]: https://github.com/tobyvee/mcp-zotero/releases/latest
