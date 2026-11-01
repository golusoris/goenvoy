// Package lintexceptions checks the generated complexity exceptions in .golangci.yml.
//
// Every rule between the praetor:generated-exceptions markers under linters.exclusions.rules
// names one test file and one of funlen, gocyclo, gocognit or cyclop, carries the comment
// "category <A-D>, entered <date>, expires <date>, #<issue>" on the line above it, and is not
// past its date. A "baseline #<issue>" rule instead pins one function of a test file under
// funlen, gocyclo or gocognit through a "text:" line; the engine ratchet only lets it shrink.
// A complexity linter outside that block fails, so no file tier comes back.
//
// This check is itself an exception to one mechanism per behavior (HISS-19; category C,
// entered 2026-11-01, expires 2026-11-11, #141): the engine audit judges the same block in the
// local hooks, and this check keeps a CI barrier until the CI runs that audit.
// Adapted from tekmora/epitrop tools/lintexceptions (same author).
package lintexceptions

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// selfExpiry is the date this package's own exception runs out.
const selfExpiry = "2026-11-11"

// maxLines bounds the configuration lines one check reads.
const maxLines = 100000

const (
	beginMarker = "      # praetor:generated-exceptions begin (praetorctl linters exceptions; never edit by hand)"
	endMarker   = "      # praetor:generated-exceptions end"
	lintersKey  = "        linters:"
)

var (
	baselineLine   = regexp.MustCompile(`^      # baseline #\d+$`)
	baselineLinter = regexp.MustCompile(`^          - (?:funlen|gocyclo|gocognit)$`)
	textLine       = regexp.MustCompile(`^        text: \S.*$`)
	commentLine    = regexp.MustCompile(`^      # category [A-D], (?:entered \d{4}-\d{2}-\d{2}, )?expires (\d{4}-\d{2}-\d{2}), #\d+$`)
	pathLine       = regexp.MustCompile(`^      - path: \^\S+_test\\\.go\$$`)
	linterLine     = regexp.MustCompile(`^          - (?:funlen|gocyclo|gocognit|cyclop)$`)
	complexityName = regexp.MustCompile(`\b(?:funlen|gocyclo|gocognit|cyclop)\b`)
)

// errNoRules is returned for a configuration without a linters.exclusions.rules block.
var errNoRules = errors.New("no linters.exclusions.rules block")

// scan holds the reader state across the lines of the rules block.
type scan struct {
	today     time.Time
	generated bool
	seenBlock bool
	expiry    string
	baseline  bool
	rules     int
}

// Check reads the configuration text and returns the number of generated rules it found, or
// an error naming the first line that breaks the contract in the package comment.
func Check(config string, today time.Time) (int, error) {
	// A Windows checkout ends lines with CRLF; the rules read LF lines.
	lines := strings.Split(strings.ReplaceAll(config, "\r\n", "\n"), "\n")
	start, err := rulesStart(lines)
	if err != nil {
		return 0, err
	}
	s := &scan{today: today}
	for i := start; i < len(lines) && i < maxLines; i++ {
		line := lines[i]
		if line != "" && !strings.HasPrefix(line, "      ") {
			break
		}
		if err := s.line(line, i+1); err != nil {
			return s.rules, err
		}
	}
	if s.generated {
		return s.rules, errors.New("generated exceptions block has no end marker")
	}
	return s.rules, nil
}

// line judges one line of the rules block.
func (s *scan) line(line string, lineNumber int) error {
	switch {
	case line == beginMarker && !s.seenBlock:
		s.generated, s.seenBlock = true, true
		return nil
	case line == endMarker && s.generated:
		s.generated = false
		return nil
	case !s.generated:
		if complexityName.MatchString(line) {
			return fmt.Errorf("line %d: complexity linter excluded outside the generated block: %q", lineNumber, line)
		}
		return nil
	}
	return s.generatedLine(line, lineNumber)
}

// generatedLine judges one line inside the generated block against the allowed shapes.
func (s *scan) generatedLine(line string, lineNumber int) error {
	if match := commentLine.FindStringSubmatch(line); match != nil {
		s.expiry, s.baseline = match[1], false
		return nil
	}
	if baselineLine.MatchString(line) {
		s.expiry, s.baseline = "", true
		return nil
	}
	if pathLine.MatchString(line) {
		return s.pathLine(lineNumber)
	}
	if line == "" || line == lintersKey || s.ruleLine(line) {
		return nil
	}
	return fmt.Errorf("line %d: %q is not an exception comment, test-file path or complexity linter line", lineNumber, line)
}

// pathLine opens one rule: a dated rule must be current, a baseline rule needs its comment.
func (s *scan) pathLine(lineNumber int) error {
	if !s.baseline {
		if err := checkRule(s.expiry, s.today, lineNumber); err != nil {
			return err
		}
	}
	s.rules++
	s.expiry = ""
	return nil
}

// ruleLine reports whether the line is a linter or text line the current rule kind allows.
func (s *scan) ruleLine(line string) bool {
	if s.baseline {
		return baselineLinter.MatchString(line) || textLine.MatchString(line)
	}
	return linterLine.MatchString(line)
}

// rulesStart returns the index of the first line after "    rules:" inside "  exclusions:".
func rulesStart(lines []string) (int, error) {
	inExclusions := false
	for i := 0; i < len(lines) && i < maxLines; i++ {
		switch {
		case lines[i] == "  exclusions:":
			inExclusions = true
		case inExclusions && lines[i] == "    rules:":
			return i + 1, nil
		}
	}
	return 0, errNoRules
}

// checkRule refuses a rule whose comment is missing or whose expiry is before today.
func checkRule(expiry string, today time.Time, lineNumber int) error {
	if expiry == "" {
		return fmt.Errorf("line %d: rule has no \"category <A-D>, entered <date>, expires <date>, #<issue>\" comment above it", lineNumber)
	}
	expired, err := Expired(expiry, today)
	if err != nil {
		return fmt.Errorf("line %d: %w", lineNumber, err)
	}
	if expired {
		return fmt.Errorf("line %d: exception expired on %s", lineNumber, expiry)
	}
	return nil
}

// Expired reports whether the date has passed: an exception holds through its expiry day.
func Expired(expiry string, today time.Time) (bool, error) {
	date, err := time.Parse(time.DateOnly, expiry)
	if err != nil {
		return false, fmt.Errorf("expiry %q: %w", expiry, err)
	}
	day := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	return day.After(date), nil
}
