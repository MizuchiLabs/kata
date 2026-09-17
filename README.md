# kata

The shared foundation for [MizuchiLabs](https://github.com/mizuchilabs) tools. Standard library only, zero dependencies.

| Package     | Purpose                                                                                                            |
| ----------- | ------------------------------------------------------------------------------------------------------------------ |
| `buildinfo` | Version/commit/date via ldflags, with `debug.ReadBuildInfo()` fallback so `go install` builds report real versions |
| `httpx`     | `StatusWriter` (status/size capture, flush and hijack passthrough) and slog access-log middleware                 |
| `logx`      | Standard slog setup: text on a terminal, JSON when piped, always stderr                                            |
| `sigx`      | `signal.NotifyContext` with force-quit on second signal                                                            |

## Install

```sh
go get github.com/mizuchilabs/kata@latest
```

## Usage

```go
import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/mizuchilabs/kata/buildinfo"
	"github.com/mizuchilabs/kata/logx"
	"github.com/mizuchilabs/kata/sigx"
	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:    "myapp",
		Version: buildinfo.String(),
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			logx.Init(cmd.Bool("debug"))
			return ctx, nil
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			slog.Info("running")
			return nil
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
                Name: "debug",
                Usage: "enable debug logging",
            },
		},
	}

	if err := cmd.Run(sigx.NotifyContext(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "myapp: %v\n", err)
		os.Exit(1)
	}
}
```

`buildinfo.UserAgent(name)` returns `name/version (goos/arch)` for HTTP clients.

`logx` redacts credentials (`password`, `token`, `authorization`, and so on) from every attribute, including nested groups. Register project-specific keys before calling `logx.Init`:

```go
logx.AddSensitiveKeys("dsn", "client_secret")
logx.Init(cmd.Bool("debug"))
```

`httpx.Logger` logs one line per request through the default slog logger: 4xx at Warn, 5xx at Error, the rest at the level you pick. The optional skip callback suppresses noisy routes. The query string and headers are never logged.

```go
mux.Handle("GET /{$}", httpx.Logger(slog.LevelInfo, func(r *http.Request, sw *httpx.StatusWriter) bool {
	return r.URL.Path == "/healthz" && sw.Status() == http.StatusOK
})(handler))
```

## goreleaser

```yaml
ldflags:
  - -s -w
  - -X github.com/mizuchilabs/kata/buildinfo.Version={{ .Version }}
  - -X github.com/mizuchilabs/kata/buildinfo.Commit={{ .Commit }}
  - -X github.com/mizuchilabs/kata/buildinfo.Date={{ .CommitDate }}
```

Without ldflags, values fall back to VCS info stamped by the Go toolchain, so binaries installed via `go install` still report their tag, commit, and date.
