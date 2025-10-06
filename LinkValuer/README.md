# LinkValuer client library

Lightweight Go client for Links Valuers Integration APIs:
- Login: POST /api/get-token
- Refresh: GET /api/refresh-token
- Create valuation: POST /api/create-api-request
- View assessments: GET /api/view-assessment
- Download report (PDF): GET /api/download-pdf/{booking_no}

## Features
- Token caching with auto-login and auto-refresh on 401
- Minimal configuration with sensible defaults
- Debug logging toggle
- Decoded responses: methods return typed structs (no json.RawMessage exposure)

## Install and import
Common ways to use this package:

1) As a standalone module (when published)
- Use your repository module path + `/LinkValuer` in imports, for example:
  
  ```go
  import linkvaluer "github.com/nana-tec/gopackages/LinkValuer"
  ```
- Then add the dependency:
  
  ```bash
  go get github.com/nana-tec/gopackages/LinkValuer
  ```

Tip: Check your root go.mod `module` line to know the correct import prefix for your setup.

## Quick start
- Set your credentials (email and password).
- Create a client with `linkvaluer.NewClient(cfg)`.
- Optionally call `Login()`; other methods will auto-login if needed.

Minimal example:

```go
cfg := &linkvaluer.Config{
    Credentials: linkvaluer.Credentials{Email: "you@example.com", Password: "secret"},
    Debug:       true, // optional
}
c, err := linkvaluer.NewClient(cfg)
if err != nil { panic(err) }

// Optional explicit login
if err := c.Login(); err != nil { panic(err) }

// Create valuation
res, err := c.CreateValuation(&linkvaluer.CreateRequest{
    CustomerName:       "Test User",
    CustomerPhone:      "0712345678",
    RegistrationNumber: "KAA000A",
    PolicyNumber:       "POL123",
    CustomerEmail:      "test@example.com",
})
if err != nil { panic(err) }
fmt.Println(res.Message, res.Data.BookingNo)

// View assessments (typed)
ass, err := c.ViewAssessments()
if err != nil { panic(err) }
fmt.Println(len(ass.Data), ass.Pagination.CurrentPage)
for _, it := range ass.Data {
    fmt.Println(it.BookingNo, it.RegNo, it.Status)
}
```

## Try the included example
Set env vars and run the example program.

- From this repository root:

```bash
export LINKVALUER_EMAIL=you@example.com
export LINKVALUER_PASSWORD=secret
# optional for PDF download in the example
export LINKVALUER_BOOKING_NO=booking_no

go run ./LinkValuer/examples
```

- If you keep a `Comprehensive/Libs` layout, run from the `Comprehensive` root:

```bash
export LINKVALUER_EMAIL=you@example.com
export LINKVALUER_PASSWORD=secret

go run ./Libs/LinkValuer/examples
```

## Configuration
- Credentials: email and password required for token generation.
- Environment/CustomEndpoint: defaults to production `https://portal.linksvaluers.com/api`; override with `Config.CustomEndpoint` if needed.
- Timeout: default 30s.
- TokenTTL: default 12h; used as a fallback cache TTL for access tokens.
- InsecureSkipVerify: false by default; set true only for testing self-signed TLS.
- Debug: logs request/response status and bodies (avoid in production).

## API notes
- Access token caching with automatic refresh on 401 if a refresh token is available.
- DownloadReport returns raw bytes and the content-type (e.g., `application/pdf`).

## Troubleshooting
- Set `Config.Debug = true` to inspect requests/responses during integration.
- If API response shapes change, update the typed models in `types.go` accordingly.
- Ensure your import path matches your module name. Use a `replace` directive when developing locally.
