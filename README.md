# kata

The shared foundation for [MizuchiLabs](https://github.com/mizuchilabs) tools. Standard library only, zero dependencies.

| Package     | Purpose                                                                                                            |
| ----------- | ------------------------------------------------------------------------------------------------------------------ |
| `buildinfo` | Version/commit/date via ldflags, with `debug.ReadBuildInfo()` fallback so `go install` builds report real versions |
| `logx`      | Standard slog setup: text on a terminal, JSON when piped, always stderr                                            |
| `sigx`      | `signal.NotifyContext` with force-quit on second signal                                                            |

## Install

```sh
go get github.com/mizuchilabs/kata@latest
```

## Usage

```go
import (
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
		// ...
	}

	if err := cmd.Run(sigx.NotifyContext(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "myapp: %v\n", err)
		os.Exit(1)
	}
}
```

`buildinfo.UserAgent(name)` returns `name/version (goos/arch)` for HTTP clients.

## goreleaser

```yaml
ldflags:
  - -s -w
  - -X github.com/mizuchilabs/kata/buildinfo.Version={{ .Version }}
  - -X github.com/mizuchilabs/kata/buildinfo.Commit={{ .Commit }}
  - -X github.com/mizuchilabs/kata/buildinfo.Date={{ .CommitDate }}
```

Without ldflags, values fall back to VCS info stamped by the Go toolchain, so binaries installed via `go install` still report their tag, commit, and date.
