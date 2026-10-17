package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// maxSpecBytes bounds one download; the pinned specs are about 300 kB each.
const maxSpecBytes = 4 << 20

// checker downloads each pinned file and compares its sha256 with the manifest.
type checker struct {
	client *http.Client
	base   string
	out    io.Writer
}

// all checks every pin and reports each that fails, so one run names every stale pin.
func (c checker) all(ctx context.Context, pins []pin) error {
	var failed []error
	for _, p := range pins {
		if err := c.one(ctx, p); err != nil {
			failed = append(failed, fmt.Errorf("%s: %w", p.Name, err))
			continue
		}
		if _, err := fmt.Fprintf(c.out, "upstreamspec: %s %s (%s): sha256 matches\n",
			p.Name, p.ref(), p.Commit[:12]); err != nil {
			return fmt.Errorf("write summary: %w", err)
		}
	}
	return errors.Join(failed...)
}

// one reads the pinned file and fails unless it has the pinned checksum.
func (c checker) one(ctx context.Context, p pin) error {
	source := p.source(c.base)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, source, http.NoBody)
	if err != nil {
		return fmt.Errorf("request %s: %w", source, err)
	}
	// #nosec G107 G704 -- the address is the fixed base plus manifest fields matched against patterns
	reply, err := c.client.Do(request)
	if err != nil {
		return fmt.Errorf("get %s: %w", source, err)
	}
	data, err := body(reply)
	if err != nil {
		return fmt.Errorf("get %s: %w", source, err)
	}
	if len(data) > maxSpecBytes {
		return fmt.Errorf("%s is larger than %d bytes", source, maxSpecBytes)
	}
	if got := digest(data); got != p.SHA256 {
		return fmt.Errorf("%s at commit %s has sha256 %s, the manifest pins %s", p.Path, p.Commit, got, p.SHA256)
	}
	return nil
}

// body closes the reply and returns what a 200 reply carried, at most one byte above the size bound.
func body(reply *http.Response) ([]byte, error) {
	if reply.StatusCode != http.StatusOK {
		return nil, errors.Join(fmt.Errorf("HTTP %d", reply.StatusCode), reply.Body.Close())
	}
	data, err := io.ReadAll(io.LimitReader(reply.Body, maxSpecBytes+1))
	if err = errors.Join(err, reply.Body.Close()); err != nil {
		return nil, fmt.Errorf("read the reply: %w", err)
	}
	return data, nil
}

func digest(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
