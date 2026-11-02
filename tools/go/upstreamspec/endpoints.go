package main

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"time"
)

type operation struct {
	verb string
	path string
}

type codeOp struct {
	funcName   string
	file       string
	line       int
	verb       string
	path       string
	unresolved bool
}

type endpointsChecker struct {
	checker checker
	root    string
}

func (c endpointsChecker) all(ctx context.Context, pins []pin, allowList []allowEntry) error {
	var failed []error
	matched := make(map[int]bool)
	for _, p := range pins {
		errs, err := c.processPin(ctx, p, allowList, matched)
		if err != nil {
			return err
		}
		failed = append(failed, errs...)
	}
	for i, entry := range allowList {
		if !matched[i] {
			failed = append(failed, fmt.Errorf("allow entry %s %s matches no off-spec call; remove it", entry.Module, entry.Func))
		}
	}
	return errors.Join(failed...)
}

func (c endpointsChecker) processPin(ctx context.Context, p pin, allowList []allowEntry, matched map[int]bool) ([]error, error) {
	if p.Module == "" {
		return nil, nil
	}
	findings, err := c.checkPin(ctx, p, allowList, matched)
	if err != nil {
		return nil, err
	}
	errs := make([]error, len(findings))
	for i, f := range findings {
		errs[i] = errors.New(f)
	}
	return errs, nil
}

func (c endpointsChecker) checkPin(ctx context.Context, p pin, allowList []allowEntry, matched map[int]bool) ([]string, error) {
	data, err := c.checker.fetch(ctx, p)
	if err != nil {
		return nil, err
	}
	specOps, err := parseSpec(data, p.Prefix, isYAML(p.Path))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", p.Name, err)
	}

	moduleDir := filepath.Join(c.root, p.Module)
	codeOps, err := parseModule(moduleDir)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", p.Name, err)
	}

	findings, inSpecCount, calledCount := evalOps(p, codeOps, specOps, allowList, matched)

	uncalled := len(specOps) - calledCount
	if _, err := fmt.Fprintf(c.checker.out, "upstreamspec: %s: code ops %d, in spec %d, spec ops not called %d\n",
		p.Name, len(codeOps), inSpecCount, uncalled); err != nil {
		return nil, fmt.Errorf("write summary: %w", err)
	}
	return findings, nil
}

func evalOps(p pin, ops []codeOp, specOps map[operation]bool, allowList []allowEntry, matched map[int]bool) ([]string, int, int) {
	inSpecCount := 0
	calledSpecOps := make(map[operation]bool)
	var findings []string
	now := time.Now()
	expiredReported := make(map[int]bool)

	tag := p.Tag
	if tag == "" {
		tag = p.Branch
	}

	for _, op := range ops {
		if op.unresolved {
			findings = append(findings, fmt.Sprintf("%s/%s:%d %s: path unresolved", p.Module, op.file, op.line, op.funcName))
			continue
		}
		if op.verb == "" {
			findings = append(findings, fmt.Sprintf("%s/%s:%d %s: verb unknown", p.Module, op.file, op.line, op.funcName))
			continue
		}
		key := operation{verb: op.verb, path: op.path}
		if specOps[key] {
			inSpecCount++
			calledSpecOps[key] = true
			continue
		}
		findings = append(findings, evalOffSpecOp(p.Module, tag, p.Name, op, allowList, matched, expiredReported, now)...)
	}
	return findings, inSpecCount, len(calledSpecOps)
}

func evalOffSpecOp(module, tag, specName string, op codeOp, allowList []allowEntry, matched, expiredReported map[int]bool, now time.Time) []string {
	matchIdx := findAllowMatch(module, op, allowList)
	if matchIdx != -1 {
		matched[matchIdx] = true
		entry := allowList[matchIdx]
		if expired, err := isExpired(entry.Expires, now); err == nil && !expired {
			return nil
		}
		if !expiredReported[matchIdx] {
			expiredReported[matchIdx] = true
			return []string{
				fmt.Sprintf("allow entry %s %s expired %s", entry.Module, entry.Func, entry.Expires),
				fmt.Sprintf("%s/%s:%d %s: %s %s not in %s %s spec", module, op.file, op.line, op.funcName, op.verb, op.path, specName, tag),
			}
		}
	}
	return []string{
		fmt.Sprintf("%s/%s:%d %s: %s %s not in %s %s spec", module, op.file, op.line, op.funcName, op.verb, op.path, specName, tag),
	}
}
