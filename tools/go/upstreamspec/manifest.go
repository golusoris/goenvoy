package main

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"

	"go.yaml.in/yaml/v3"
)

// pin is one upstream OpenAPI spec, fixed to the bytes of one file at one commit.
//
// Tag is the upstream release the commit belongs to. Renovate reads Repository, Tag and Commit
// (renovate.json) and proposes the next stable release; SHA256 then decides whether the spec file
// changed with it. An upstream project that tags no release is pinned with Branch in the place of
// Tag: the commit is then one of that branch, and Renovate proposes the newest one.
type pin struct {
	Name       string `yaml:"name"`
	Repository string `yaml:"repository"`
	Tag        string `yaml:"tag"`
	Branch     string `yaml:"branch"`
	Commit     string `yaml:"commit"`
	Path       string `yaml:"path"`
	SHA256     string `yaml:"sha256"`
	License    string `yaml:"license"`
	Prefix     string `yaml:"prefix"`
}

var (
	pinName    = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)
	repository = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9-]*/[A-Za-z0-9._-]+$`)
	tagName    = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)
	branchName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*(/[A-Za-z0-9][A-Za-z0-9._-]*)*$`)
	commitID   = regexp.MustCompile(`^[0-9a-f]{40}$`)
	sha256Hex  = regexp.MustCompile(`^[0-9a-f]{64}$`)
	licenseID  = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9.+-]*$`)
	prefixPath = regexp.MustCompile(`^(/[A-Za-z0-9._-]+)*$`)
)

// readManifest decodes the manifest and refuses one that is not well formed. A key the tool does
// not read is refused, so a misspelled field cannot pass as an absent one.
func readManifest(file string) ([]pin, error) {
	// #nosec G304 -- the manifest path is the caller's argument; the file is only decoded
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}
	var manifest struct {
		Specs []pin `yaml:"specs"`
	}
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&manifest); err != nil {
		return nil, fmt.Errorf("%s: %w", file, err)
	}
	if len(manifest.Specs) == 0 {
		return nil, fmt.Errorf("%s: no \"specs\" entry", file)
	}
	names, commits := map[string]bool{}, map[string]string{}
	for index, p := range manifest.Specs {
		problem := p.problem()
		if problem == "" {
			problem = p.repeatProblem(names, commits)
		}
		if problem != "" {
			return nil, fmt.Errorf("%s: specs[%d] (%s): %s", file, index, p.Name, problem)
		}
	}
	return manifest.Specs, nil
}

// repeatProblem refuses an entry that repeats the name of an earlier one, or its commit under
// another repository. A repository can carry tags of another project: Sonarr/Sonarr holds the
// v6 release tags of Radarr, so a tag lookup proposes a commit of the other project. Two
// projects never share a commit, so the second entry names the wrong one.
func (p pin) repeatProblem(names map[string]bool, commits map[string]string) string {
	owner, shared := commits[p.Commit]
	switch {
	case names[p.Name]:
		return "repeats the name of an earlier entry"
	case shared && owner != p.Repository:
		return fmt.Sprintf("repeats the commit of %s; a commit belongs to one repository, so this tag or branch names another project", owner)
	}
	names[p.Name], commits[p.Commit] = true, p.Repository
	return ""
}

func (p pin) problem() string {
	checks := []struct {
		ok   bool
		text string
	}{
		{pinName.MatchString(p.Name), "name must be lower-case letters, digits and hyphens"},
		{repository.MatchString(p.Repository), "repository must be <owner>/<name>"},
		{(p.Tag == "") != (p.Branch == ""), "exactly one of tag and branch must be set"},
		{p.Tag == "" || tagName.MatchString(p.Tag), "tag must name one release tag"},
		{p.Branch == "" || branchName.MatchString(p.Branch), "branch must name one branch"},
		{commitID.MatchString(p.Commit), "commit must be a full 40-digit commit id"},
		{cleanRelative(p.Path), "path must be one clean file path inside the upstream repository"},
		{sha256Hex.MatchString(p.SHA256), "sha256 must be 64 lower-case hex digits"},
		{licenseID.MatchString(p.License), "license must be the SPDX id of the upstream repository"},
		{prefixPath.MatchString(p.Prefix), "prefix must be empty or a path like /api/v1"},
	}
	for _, check := range checks {
		if !check.ok {
			return check.text
		}
	}
	return ""
}

// cleanRelative reports whether file names one path below a root: not absolute, no "..", no
// trailing slash, no "." segment.
func cleanRelative(file string) bool {
	if file == "" || path.IsAbs(file) || strings.ContainsAny(file, "\\ \t\r\n") {
		return false
	}
	return path.Clean(file) == file && file != "." && file != ".." && !strings.HasPrefix(file, "../")
}

// source is where the pinned file is read from below base.
func (p pin) source(base string) string {
	return strings.TrimSuffix(base, "/") + "/" + p.Repository + "/" + p.Commit + "/" + p.Path
}

// ref is the tag or the branch the commit belongs to, as the summary of a check names it.
func (p pin) ref() string {
	if p.Branch != "" {
		return "branch " + p.Branch
	}
	return p.Tag
}
