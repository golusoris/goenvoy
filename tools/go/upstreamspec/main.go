// Command upstreamspec checks the upstream OpenAPI spec pins in upstream/specs.yaml.
//
// Each entry pins one spec file at one commit of an upstream release, with the sha256 of that
// file. `upstreamspec check` refuses a manifest that is not well formed, downloads every pinned
// file and refuses one whose sha256 differs, naming the sha256 it found. Nothing is written to
// disk: the Sonarr, Radarr, Lidarr and Readarr-fork specs are GPL-3.0 and this repository is MIT
// (upstream/README.md).
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	// rawBase serves a file of a GitHub repository at a commit: <base>/<owner>/<name>/<commit>/<path>.
	rawBase = "https://raw.githubusercontent.com"
	// deadline bounds one run: seven downloads of about 300 kB each.
	deadline = 2 * time.Minute
)

// errUsage is returned for a command line the tool does not understand.
var errUsage = errors.New("usage: upstreamspec check [-manifest <specs.yaml>]")

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), deadline)
	err := run(ctx, os.Args[1:], os.Stdout, rawBase)
	cancel()
	if err != nil {
		fmt.Fprintln(os.Stderr, "upstreamspec:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, out io.Writer, base string) error {
	if len(args) == 0 || args[0] != "check" {
		return errUsage
	}
	flags := flag.NewFlagSet("upstreamspec", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	manifest := flags.String("manifest", "upstream/specs.yaml", "manifest of pinned specs")
	if err := flags.Parse(args[1:]); err != nil {
		return fmt.Errorf("%w; %w", err, errUsage)
	}
	if flags.NArg() != 0 {
		return errUsage
	}
	pins, err := readManifest(*manifest)
	if err != nil {
		return err
	}
	return checker{client: &http.Client{Timeout: deadline}, base: base, out: out}.all(ctx, pins)
}
