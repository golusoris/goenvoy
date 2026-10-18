package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"
)

const (
	// committedManifest is the manifest as the Renovate configuration names it, as a pattern.
	committedManifest     = `/^upstream/specs\.yaml$/`
	committedManifestPath = "../../../upstream/specs.yaml"
	manifestFile          = "upstream/specs.yaml"
	// tagSource lists the releases of a repository with the date and pre-release flag of each;
	// a tag list would also offer tags another project pushed into the repository (Sonarr/Sonarr
	// carries Radarr's). branchSource resolves a
	// branch to its newest commit and dates it; git-refs would do the first but gives no date,
	// and Renovate holds an update without one for as long as minimumReleaseAge is set.
	tagSource    = "github-releases"
	branchSource = "github-digest"
	// readers is the number of custom managers that read the manifest: one per datasource.
	readers = 2
)

// renovateConfig is the part of renovate.json that decides whether Renovate keeps the pins of the
// manifest current.
//
//nolint:tagliatelle // renovate.json is camelCase
type renovateConfig struct {
	CustomManagers []customManager `json:"customManagers"`
	PackageRules   []packageRule   `json:"packageRules"`
}

//nolint:tagliatelle // renovate.json is camelCase
type customManager struct {
	CustomType          string   `json:"customType"`
	ManagerFilePatterns []string `json:"managerFilePatterns"`
	MatchStrings        []string `json:"matchStrings"`
	DatasourceTemplate  string   `json:"datasourceTemplate"`
	VersioningTemplate  string   `json:"versioningTemplate"`
}

//nolint:tagliatelle // renovate.json is camelCase
type packageRule struct {
	MatchFileNames   []string `json:"matchFileNames"`
	MatchManagers    []string `json:"matchManagers"`
	MatchUpdateTypes []string `json:"matchUpdateTypes"`
	Automerge        *bool    `json:"automerge"`
}

// reading is what Renovate takes from one manifest entry, and the datasource it asks about it.
type reading struct{ datasource, repository, ref, commit string }

// wanted returns the reading that keeps p current: a tag pin is asked of the releases of its
// repository, a branch pin of the newest commit of its branch.
func wanted(p pin) reading {
	if p.Branch != "" {
		return reading{branchSource, p.Repository, p.Branch, p.Commit}
	}
	return reading{tagSource, p.Repository, p.Tag, p.Commit}
}

// versioningFor returns the versioning a manager of the datasource must name, and whether the
// datasource is one the pins are asked of. Tags have four parts, which only loose compares; a
// branch is compared with nothing, which is the default of its datasource.
func versioningFor(datasource string) (versioning string, known bool) {
	switch datasource {
	case tagSource:
		return "loose", true
	case branchSource:
		return "", true
	default:
		return "", false
	}
}

// renovateProblems returns why Renovate would not keep the pins current, one line per reason:
// the manifest is not read by one manager for tag pins and one for branch pins, a manager asks
// another datasource or compares with another versioning, the patterns do not capture a pin as
// the manifest has it or capture it for the wrong datasource, or a pin update would be merged
// without a person.
func renovateProblems(config renovateConfig, manifest string, pins []pin) []string {
	var problems []string
	if !config.mergedByHand() {
		problems = append(problems, "the last package rule that sets automerge for upstream/specs.yaml does not set it to false")
	}
	var reading []customManager
	for _, manager := range config.CustomManagers {
		if slices.Contains(manager.ManagerFilePatterns, committedManifest) {
			reading = append(reading, manager)
		}
	}
	if len(reading) != readers {
		return append(problems, fmt.Sprintf("%d custom manager(s) read %s; want %d, one for tag pins and one for branch pins",
			len(reading), manifestFile, readers))
	}
	var settings []string
	for _, manager := range reading {
		settings = append(settings, manager.settingProblems()...)
	}
	if len(settings) > 0 {
		return append(problems, settings...)
	}
	return append(problems, pinProblems(reading, manifest, pins)...)
}

// mergedByHand reports whether the last package rule that sets automerge and can apply to an
// update of the manifest sets it to false. Later rules win in Renovate. A rule counts as
// applicable unless it names other files or other managers, which is stricter than Renovate: its
// other matchers are not read here. A rule limited to some update types counts only when it
// turns automerge on: turning it off for majors leaves a digest update merged by the bot.
func (c renovateConfig) mergedByHand() bool {
	byHand := false
	for _, rule := range c.PackageRules {
		if rule.Automerge == nil {
			continue
		}
		if len(rule.MatchFileNames) > 0 && !slices.Contains(rule.MatchFileNames, manifestFile) {
			continue
		}
		if len(rule.MatchManagers) > 0 && !slices.Contains(rule.MatchManagers, "custom.regex") {
			continue
		}
		if len(rule.MatchUpdateTypes) > 0 && !*rule.Automerge {
			continue
		}
		byHand = !*rule.Automerge
	}
	return byHand
}

func (m customManager) settingProblems() []string {
	versioning, known := versioningFor(m.DatasourceTemplate)
	if !known {
		return []string{fmt.Sprintf("datasourceTemplate is %q; want %q for tag pins or %q for branch pins",
			m.DatasourceTemplate, tagSource, branchSource)}
	}
	settings := []struct{ name, got, want string }{
		{"customType", m.CustomType, "regex"},
		{"versioningTemplate", m.VersioningTemplate, versioning},
	}
	var problems []string
	for _, setting := range settings {
		if setting.got != setting.want {
			problems = append(problems, fmt.Sprintf("the %s manager: %s is %q; want %q", m.DatasourceTemplate, setting.name, setting.got, setting.want))
		}
	}
	return problems
}

// readInto applies matchStrings to the manifest text the way the regex manager does (every match
// of every pattern is one dependency), stores what it captures by dependency name and returns the
// number of matches.
func (m customManager) readInto(read map[string]reading, manifest string) (int, error) {
	matches := 0
	for _, pattern := range m.MatchStrings {
		expression, err := regexp.Compile(pattern)
		if err != nil {
			return 0, fmt.Errorf("matchStrings: %w", err)
		}
		for _, match := range expression.FindAllStringSubmatch(manifest, -1) {
			group := func(name string) string {
				if index := expression.SubexpIndex(name); index >= 0 {
					return match[index]
				}
				return ""
			}
			read[group("depName")] = reading{m.DatasourceTemplate, group("packageName"), group("currentValue"), group("currentDigest")}
			matches++
		}
	}
	return matches, nil
}

// pinProblems compares what the managers capture from the manifest text with the pins.
func pinProblems(managers []customManager, manifest string, pins []pin) []string {
	read := map[string]reading{}
	matches := 0
	for _, manager := range managers {
		count, err := manager.readInto(read, manifest)
		if err != nil {
			return []string{err.Error()}
		}
		matches += count
	}
	var problems []string
	for _, p := range pins {
		if got, want := read[p.Name], wanted(p); got != want {
			problems = append(problems, fmt.Sprintf("%s: Renovate reads %+v; the manifest pins %+v", p.Name, got, want))
		}
	}
	if matches != len(pins) {
		problems = append(problems, fmt.Sprintf("matchStrings match %d time(s); the manifest has %d pin(s)", matches, len(pins)))
	}
	return problems
}

// committed returns the Renovate configuration, the manifest text and its pins as committed.
func committed(t *testing.T) (config renovateConfig, manifest string, pins []pin) {
	t.Helper()
	data, err := os.ReadFile("../../../renovate.json")
	if err != nil {
		t.Fatalf("read the Renovate configuration: %v", err)
	}
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("decode the Renovate configuration: %v", err)
	}
	text, err := os.ReadFile(committedManifestPath)
	if err != nil {
		t.Fatalf("read the manifest: %v", err)
	}
	pins, err = readManifest(committedManifestPath)
	if err != nil {
		t.Fatalf("readManifest: %v", err)
	}
	return config, string(text), pins
}

// TestRenovateKeepsThePinsCurrent binds renovate.json to the manifest: every pin is read, by the
// manager of its kind, as the manifest has it, and is merged by a person.
func TestRenovateKeepsThePinsCurrent(t *testing.T) {
	t.Parallel()
	config, manifest, pins := committed(t)
	if problems := renovateProblems(config, manifest, pins); len(problems) > 0 {
		t.Fatalf("renovate.json does not keep the pins current:\n%s", strings.Join(problems, "\n"))
	}
}

// managerOf returns the custom manager of the datasource in config.
func managerOf(t *testing.T, config *renovateConfig, datasource string) *customManager {
	t.Helper()
	for index := range config.CustomManagers {
		manager := &config.CustomManagers[index]
		if manager.DatasourceTemplate == datasource && slices.Contains(manager.ManagerFilePatterns, committedManifest) {
			return manager
		}
	}
	t.Fatalf("no custom manager of %s reads the manifest", datasource)
	return nil
}

// TestRenovateRefuses plants each defect in a copy of the committed configuration or manifest
// and expects renovateProblems to name it.
func TestRenovateRefuses(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		change func(t *testing.T, config *renovateConfig, manifest *string)
		want   string
	}{
		{"the tag manager reads another file", func(t *testing.T, c *renovateConfig, _ *string) {
			t.Helper()
			managerOf(t, c, tagSource).ManagerFilePatterns = []string{"specs.yaml"}
		}, "1 custom manager(s) read"},
		{"a datasource that lists every tag", func(t *testing.T, c *renovateConfig, _ *string) {
			t.Helper()
			managerOf(t, c, tagSource).DatasourceTemplate = "github-tags"
		}, `datasourceTemplate is "github-tags"`},
		{"the default versioning for tags", func(t *testing.T, c *renovateConfig, _ *string) {
			t.Helper()
			managerOf(t, c, tagSource).VersioningTemplate = ""
		}, `the github-releases manager: versioningTemplate is ""`},
		{"a pattern that does not compile", func(t *testing.T, c *renovateConfig, _ *string) {
			t.Helper()
			managerOf(t, c, tagSource).MatchStrings = []string{"("}
		}, "matchStrings: "},
		{"branch pins asked of the tags", func(t *testing.T, c *renovateConfig, _ *string) {
			t.Helper()
			branches := managerOf(t, c, branchSource)
			branches.DatasourceTemplate, branches.VersioningTemplate = tagSource, "loose"
		}, ": Renovate reads {datasource:github-releases"},
		{"no rule keeps the pins from automerge", func(_ *testing.T, c *renovateConfig, _ *string) {
			c.PackageRules = slices.DeleteFunc(c.PackageRules, func(r packageRule) bool {
				return slices.Contains(r.MatchFileNames, manifestFile)
			})
		}, "automerge"},
		{"the rule for the pins comes before the general automerge rule", func(_ *testing.T, c *renovateConfig, _ *string) {
			index := slices.IndexFunc(c.PackageRules, func(r packageRule) bool { return slices.Contains(r.MatchFileNames, manifestFile) })
			rule := c.PackageRules[index]
			c.PackageRules = slices.Insert(slices.Delete(c.PackageRules, index, index+1), 0, rule)
		}, "automerge"},
		{"fields in another order", func(_ *testing.T, _ *renovateConfig, m *string) {
			*m = strings.Replace(*m, "    repository: Sonarr/Sonarr\n    tag: v4.0.20.3014\n",
				"    tag: v4.0.20.3014\n    repository: Sonarr/Sonarr\n", 1)
		}, "sonarr: Renovate reads"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			config, manifest, pins := committed(t)
			test.change(t, &config, &manifest)
			problems := strings.Join(renovateProblems(config, manifest, pins), "\n")
			if !strings.Contains(problems, test.want) {
				t.Fatalf("problems = %q, want one with %q", problems, test.want)
			}
		})
	}
}
