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
	"path/filepath"
	"time"
)

const (
	// rawBase serves a file of a GitHub repository at a commit: <base>/<owner>/<name>/<commit>/<path>.
	rawBase = "https://raw.githubusercontent.com"
	// deadline bounds one run: seven downloads of about 300 kB each.
	deadline = 2 * time.Minute
)

// errUsage is returned for a command line the tool does not understand.
var errUsage = errors.New("usage: upstreamspec check [-manifest <specs.yaml>] | endpoints [-manifest <specs.yaml>] [-root <dir>]")

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), deadline)
	err := run(ctx, os.Args[1:], os.Stdout, rawBase)
	cancel()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, out io.Writer, base string) error {
	if len(args) == 0 {
		return errUsage
	}
	switch args[0] {
	case "check":
		return runCheck(ctx, args[1:], out, base)
	case "endpoints":
		return runEndpoints(ctx, args[1:], out, base)
	default:
		return errUsage
	}
}

func runCheck(ctx context.Context, args []string, out io.Writer, base string) error {
	flags := flag.NewFlagSet("check", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	manifest := flags.String("manifest", "upstream/specs.yaml", "manifest of pinned specs")
	if err := flags.Parse(args); err != nil {
		return fmt.Errorf("%w; %w", err, errUsage)
	}
	if flags.NArg() != 0 {
		return errUsage
	}
	pins, err := readManifest(*manifest)
	if err != nil {
		return err
	}
	return checker{client: &http.Client{Timeout: deadline, Transport: sharedTransport}, base: base, out: out}.all(ctx, pins)
}

func runEndpoints(ctx context.Context, args []string, out io.Writer, base string) error {
	flags := flag.NewFlagSet("endpoints", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	manifest := flags.String("manifest", "upstream/specs.yaml", "manifest of pinned specs")
	root := flags.String("root", ".", "repository root directory")
	if err := flags.Parse(args); err != nil {
		return fmt.Errorf("%w; %w", err, errUsage)
	}
	if flags.NArg() != 0 {
		return errUsage
	}
	pins, err := readManifest(*manifest)
	if err != nil {
		return err
	}
	allowList, err := readAllowList(filepath.Join(*root, "upstream", "endpoints-allow.yaml"))
	if err != nil {
		return err
	}
	chk := endpointsChecker{
		checker: checker{client: &http.Client{Timeout: deadline, Transport: sharedTransport}, base: base, out: out},
		root:    *root,
	}
	return chk.all(ctx, pins, allowList)
}
