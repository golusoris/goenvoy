package launchbox

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

type fakeCloseErrRC struct {
	io.Reader
	closeErr error
}

func (f *fakeCloseErrRC) Close() error {
	return f.closeErr
}

type testXMLPayload struct {
	Value int `xml:"Value"`
}

func runDecodeEntryTest(t *testing.T, open func() (io.ReadCloser, error), name, wantErrSub string, wantErrIs error, wantVal int) {
	t.Helper()
	var v testXMLPayload
	err := decodeEntry(open, name, &v)

	if wantErrSub != "" {
		if err == nil {
			t.Fatalf("expected error containing %q, got nil", wantErrSub)
		}
		if !strings.Contains(err.Error(), wantErrSub) {
			t.Fatalf("err = %q, want substring %q", err.Error(), wantErrSub)
		}
	} else if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if wantErrIs != nil && !errors.Is(err, wantErrIs) {
		t.Fatalf("errors.Is(err, %v) = false; err: %v", wantErrIs, err)
	}

	if wantVal != 0 && v.Value != wantVal {
		t.Fatalf("decoded value = %d, want %d", v.Value, wantVal)
	}
}

func TestDecodeEntry(t *testing.T) {
	t.Parallel()

	errBoomClose := errors.New("boom close")
	errBoomOpen := errors.New("boom open")

	t.Run("open_error", func(t *testing.T) {
		t.Parallel()
		runDecodeEntryTest(t, func() (io.ReadCloser, error) { return nil, errBoomOpen }, "OpenEntry.xml", "launchbox: open OpenEntry.xml:", errBoomOpen, 0)
	})
	t.Run("decode_error", func(t *testing.T) {
		t.Parallel()
		runDecodeEntryTest(t, func() (io.ReadCloser, error) { return io.NopCloser(strings.NewReader("invalid < xml")), nil }, "DecodeEntry.xml", "launchbox: parse DecodeEntry.xml:", nil, 0)
	})
	t.Run("close_error", func(t *testing.T) {
		t.Parallel()
		runDecodeEntryTest(t, func() (io.ReadCloser, error) {
			return &fakeCloseErrRC{
				Reader:   strings.NewReader("<testXMLPayload><Value>42</Value></testXMLPayload>"),
				closeErr: errBoomClose,
			}, nil
		}, "CloseEntry.xml", "launchbox: close CloseEntry.xml:", errBoomClose, 42)
	})
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		runDecodeEntryTest(t, func() (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader("<testXMLPayload><Value>100</Value></testXMLPayload>")), nil
		}, "SuccessEntry.xml", "", nil, 100)
	})
}

func makeCorruptPlatformsZip(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	mw, err := zw.Create("Metadata.xml")
	if err != nil {
		t.Fatal(err)
	}
	_, err = mw.Write([]byte(`<?xml version="1.0" standalone="yes"?>
<LaunchBox>
  <Game>
    <Name>New Game</Name>
    <DatabaseID>99</DatabaseID>
  </Game>
</LaunchBox>`))
	if err != nil {
		t.Fatal(err)
	}

	pw, err := zw.Create("Platforms.xml")
	if err != nil {
		t.Fatal(err)
	}
	_, err = pw.Write([]byte(`corrupted xml data <<`))
	if err != nil {
		t.Fatal(err)
	}

	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestParseZip_AtomicPublish(t *testing.T) {
	t.Parallel()

	c := &Client{
		games: []Game{
			{DatabaseID: 10, Name: "Initial Game"},
		},
		gamesByID: make(map[int]*Game),
	}
	c.gamesByID[10] = &c.games[0]

	corruptData := makeCorruptPlatformsZip(t)
	err := c.parseZip(corruptData)
	if err == nil {
		t.Fatal("expected error from corrupt zip, got nil")
	}

	if len(c.games) != 1 || c.games[0].DatabaseID != 10 {
		t.Fatalf("games was mutated on error: %v", c.games)
	}
	if c.gamesByID[10] == nil || c.gamesByID[10].DatabaseID != 10 {
		t.Fatalf("gamesByID was corrupted on error: %v", c.gamesByID)
	}
}
