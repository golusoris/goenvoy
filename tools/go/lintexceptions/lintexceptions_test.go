package lintexceptions

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const (
	header = "linters:\n  exclusions:\n    rules:\n" +
		"      - path: doc\\.go\n        linters:\n          - godoclint\n"
	begin = beginMarker + "\n"
	end   = endMarker + "\n"
	rule  = "      # category C, entered 2026-10-19, expires 2026-10-29, #141\n" +
		"      - path: ^arr/a_test\\.go$\n        linters:\n          - gocognit\n"
)

func day(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.DateOnly, value)
	if err != nil {
		t.Fatalf("parse %s: %v", value, err)
	}
	return parsed
}

func TestRepositoryExceptionsAreDatedAndCurrent(t *testing.T) {
	t.Parallel()
	body, err := os.ReadFile(filepath.Join("..", "..", "..", ".golangci.yml"))
	if err != nil {
		t.Fatalf("read .golangci.yml: %v", err)
	}
	rules, err := Check(string(body), time.Now().UTC())
	if err != nil {
		t.Fatalf(".golangci.yml: %v", err)
	}
	t.Logf("%d dated exception rules", rules)
}

func TestThisCheckIsItselfCurrent(t *testing.T) {
	t.Parallel()
	expired, err := Expired(selfExpiry, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	if expired {
		t.Fatalf("tools/go/lintexceptions expired on %s: adopt the engine audit (cordanaLLM/praetor#1388) or renew with a lower count in #141", selfExpiry)
	}
}

func TestCheckAcceptsDatedRules(t *testing.T) {
	t.Parallel()
	config := header + begin + rule +
		"      # category C, expires 2026-10-29, #141\n" +
		"      - path: ^metadata/b_test\\.go$\n        linters:\n          - funlen\n" +
		end + "issues:\n  uniq-by-line: false\n"
	rules, err := Check(config, day(t, "2026-10-19"))
	if err != nil || rules != 2 {
		t.Fatalf("Check = %d, %v; want 2, nil", rules, err)
	}
}

func TestCheckAcceptsNoGeneratedBlock(t *testing.T) {
	t.Parallel()
	rules, err := Check(header+"formatters:\n", day(t, "2026-10-19"))
	if err != nil || rules != 0 {
		t.Fatalf("Check = %d, %v; want 0, nil", rules, err)
	}
}

func TestCheckReadsCRLFLines(t *testing.T) {
	t.Parallel()
	config := strings.ReplaceAll(header+begin+rule+end, "\n", "\r\n")
	rules, err := Check(config, day(t, "2026-10-19"))
	if err != nil || rules != 1 {
		t.Fatalf("Check = %d, %v; want 1, nil", rules, err)
	}
}

func TestCheckRefusesPlantedDefects(t *testing.T) {
	t.Parallel()
	path := "      - path: ^arr/a_test\\.go$\n        linters:\n          - gocyclo\n"
	cases := map[string]string{
		"expired":         begin + "      # category C, entered 2026-10-01, expires 2026-10-11, #141\n" + path + end,
		"untagged":        begin + path + end,
		"no category":     begin + "      # expires 2026-10-29, #141\n" + path + end,
		"bad date":        begin + "      # category C, expires 2026-13-40, #141\n" + path + end,
		"production file": begin + strings.Replace(rule, "a_test", "a", 1) + end,
		"other linter":    begin + strings.Replace(rule, "gocognit", "errcheck", 1) + end,
		"text rule":       begin + strings.Replace(rule, "        linters:\n          - gocognit\n", "        text: anything\n", 1) + end,
		"no end marker":   begin + rule,
		"tier outside":    "      - path: (.+)_test\\.go\n        linters:\n          - funlen\n",
	}
	for name, rules := range cases {
		if _, err := Check(header+rules, day(t, "2026-10-19")); err == nil {
			t.Errorf("%s: Check accepted %q", name, rules)
		}
	}
	if _, err := Check("linters:\n  enable:\n    - gocyclo\n", day(t, "2026-10-19")); err == nil {
		t.Error("a configuration without a rules block was accepted")
	}
}

func TestExpiredHoldsThroughTheExpiryDay(t *testing.T) {
	t.Parallel()
	cases := []struct {
		today string
		want  bool
	}{
		{"2026-10-28", false},
		{"2026-10-29", false},
		{"2026-10-30", true},
	}
	for _, c := range cases {
		got, err := Expired("2026-10-29", day(t, c.today))
		if err != nil || got != c.want {
			t.Errorf("Expired(2026-10-29, %s) = %v, %v; want %v", c.today, got, err, c.want)
		}
	}
}
