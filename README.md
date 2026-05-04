# azuredevops-go

A Go SDK for Azure DevOps webhooks and API integrations.

## Packages

| Package | Description |
|---------|-------------|
| [`webhooks`](./webhooks/) | Typed event handling for Azure DevOps Service Hook events |

## Installation

```bash
go get github.com/jensholdgaard/azuredevops-go
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "net/http"

    "github.com/jensholdgaard/azuredevops-go/webhooks"
)

func main() {
    handler := webhooks.New(webhooks.WithBasicAuth("webhook", "secret"))

    handler.OnGitPush(func(ctx context.Context, deliveryID string, event *webhooks.GitPushEvent) error {
        fmt.Printf("Push to %s by %s\n",
            event.Resource.Repository.Name,
            event.Resource.PushedBy.DisplayName,
        )
        return nil
    })

    http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
        if err := handler.HandleEventRequest(r); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }
        w.WriteHeader(http.StatusOK)
    })

    http.ListenAndServe(":8080", nil)
}
```

## Features

- **Typed event structs** — Full ADO webhook payload types with JSON tags
- **Two-pass JSON decode** — Peek at `eventType`, then strict decode with `DisallowUnknownFields`
- **Constant-time auth** — Basic auth validation using `crypto/subtle`
- **Parallel dispatch** — Multiple callbacks run concurrently via `errgroup`
- **Panic recovery** — Individual callback panics don't crash the handler
- **Lifecycle hooks** — `OnBeforeAny`, `OnAfterAny`, `OnError` for cross-cutting concerns
- **Delivery correlation** — Extracts `X-Vss-Activityid` header as delivery ID

## Supported Events

| Event | Handler |
|-------|---------|
| `git.push` | `OnGitPush` |

More event types coming in future releases (`git.pullrequest.*`, `build.complete`, `workitem.*`).

## Design

Azure DevOps puts the event type in the JSON body (not an HTTP header), so parsing requires a two-pass approach:

1. Lenient decode to read `eventType`
2. Strict decode into the typed struct with `DisallowUnknownFields`

This catches API drift early — if Azure adds fields we don't model, tests fail.

## Contributing

Contributions welcome. Please open an issue first for new event types to coordinate work.

## License

MIT — see [LICENSE](./LICENSE).
