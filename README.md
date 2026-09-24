# sitegpt-go

Official Go client for the [SiteGPT](https://sitegpt.ai) API v2 —
zero dependencies, mirroring the
[Python SDK](https://pypi.org/project/sitegpt/) and
[TypeScript SDK](https://www.npmjs.com/package/@sitegpt/sdk).

```bash
go get github.com/sitegpt/sitegpt-go
```

## Quickstart

```go
package main

import (
	"context"
	"fmt"
	"os"

	sitegpt "github.com/sitegpt/sitegpt-go"
)

func main() {
	client := sitegpt.NewClient(os.Getenv("SITEGPT_API_TOKEN"))
	ctx := context.Background()

	chatbots, err := client.Chatbots.List(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Println(chatbots)

	// Destructive calls demand explicit confirmation:
	// client.Chatbots.Delete(ctx, chatbotID, true)
}
```

Create an API token on the dashboard's Agents page — or, for agents
without an account, use the anonymous onboarding flow documented at
[sitegpt.ai/auth.md](https://sitegpt.ai/auth.md)
(`client.Onboarding.Start`).

## Surface

Namespaces have full helper parity with the Python and TypeScript
SDKs: `Chatbots` (including `Analytics`), `Knowledge` (documents,
sources, ingest jobs), `Conversations` (including `TakeOver` and
`SwitchToAI` for the human-handover lifecycle), `Leads`, `Messages`,
`Onboarding`, plus `Me`, `Health`, and a generic `Request` escape
hatch for anything the API adds next — the contract is the OpenAPI
document at
[sitegpt.ai/api/v2/openapi.json](https://sitegpt.ai/api/v2/openapi.json)
(markdown description: [openapi.json.md](https://sitegpt.ai/api/v2/openapi.json.md)).

- Errors are structured: `*sitegpt.Error` with `Status`, `Code`,
  `Message`, `Hint`.
- Destructive helpers refuse to run without `confirm=true` — locally,
  before any network call.
- Cross-origin redirects never leak the bearer token (Go's http.Client
  strips Authorization on host changes).

## More SiteGPT for agents

- MCP server (17 tools, inline app views): `https://sitegpt.ai/mcp`
- CLI: `npm install -g @sitegpt/cli`
- All agent surfaces: [sitegpt.ai/agents](https://sitegpt.ai/agents)
