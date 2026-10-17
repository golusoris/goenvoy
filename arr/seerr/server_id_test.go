package seerr_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/golusoris/goenvoy/arr/seerr"
)

func TestMediaRequestUnmarshalServerID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		in           string
		wantID       int
		wantAssigned bool
		wantErr      bool
	}{
		{name: "server 0 is assigned", in: `{"id":1,"serverId":0}`, wantID: 0, wantAssigned: true},
		{name: "server 1 is assigned", in: `{"id":1,"serverId":1}`, wantID: 1, wantAssigned: true},
		{name: "null is no server", in: `{"id":1,"serverId":null}`},
		{name: "absent is no server", in: `{"id":1}`},
		{name: "a string id fails", in: `{"id":1,"serverId":"0"}`, wantErr: true},
		{name: "invalid JSON fails", in: `{"id":`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var r seerr.MediaRequest
			err := json.Unmarshal([]byte(tt.in), &r)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if r.ID != 1 || r.ServerID != tt.wantID || r.ServerAssigned != tt.wantAssigned {
				t.Errorf("got ID=%d ServerID=%d ServerAssigned=%v, want ID=1 ServerID=%d ServerAssigned=%v",
					r.ID, r.ServerID, r.ServerAssigned, tt.wantID, tt.wantAssigned)
			}
		})
	}
}

func TestMediaRequestUnmarshalNested(t *testing.T) {
	t.Parallel()

	var m seerr.MediaInfo
	in := `{"id":7,"requests":[{"id":1,"serverId":0,"type":"tv","is4k":true},{"id":2,"serverId":null}]}`
	if err := json.Unmarshal([]byte(in), &m); err != nil {
		t.Fatal(err)
	}
	if len(m.Requests) != 2 {
		t.Fatalf("requests = %d, want 2", len(m.Requests))
	}
	first, second := m.Requests[0], m.Requests[1]
	if !first.ServerAssigned || first.Type != "tv" || !first.Is4k {
		t.Errorf("first = %+v, want server 0 assigned, type tv, 4K", first)
	}
	if second.ServerAssigned {
		t.Errorf("second.ServerAssigned = true, want false for a null serverId")
	}
}

func TestServerIDMarshal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   any
		want string // substring the encoding must hold; "" means serverId must be absent
	}{
		{name: "request server 0 assigned", in: seerr.MediaRequest{ID: 1, ServerAssigned: true}, want: `"serverId":0`},
		{name: "request server 2", in: seerr.MediaRequest{ID: 1, ServerID: 2}, want: `"serverId":2`},
		{name: "request without server", in: seerr.MediaRequest{ID: 1}},
		{name: "body server 0 assigned", in: seerr.CreateRequestBody{MediaType: "movie", MediaID: 5, ServerAssigned: true}, want: `"serverId":0`},
		{name: "body server 3 via pointer", in: &seerr.CreateRequestBody{MediaType: "movie", MediaID: 5, ServerID: 3}, want: `"serverId":3`},
		{name: "body without server", in: seerr.CreateRequestBody{MediaType: "movie", MediaID: 5}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, err := json.Marshal(tt.in)
			if err != nil {
				t.Fatal(err)
			}
			got := string(out)
			if strings.Contains(got, "ServerAssigned") {
				t.Errorf("encoding %s leaks ServerAssigned", got)
			}
			if tt.want == "" {
				if strings.Contains(got, "serverId") {
					t.Errorf("encoding %s holds serverId, want it absent", got)
				}
				return
			}
			if !strings.Contains(got, tt.want) {
				t.Errorf("encoding %s lacks %s", got, tt.want)
			}
		})
	}
}

func TestMediaRequestRoundTrip(t *testing.T) {
	t.Parallel()

	for _, in := range []string{`{"id":1,"serverId":0}`, `{"id":1,"serverId":null}`} {
		var first, second seerr.MediaRequest
		if err := json.Unmarshal([]byte(in), &first); err != nil {
			t.Fatal(err)
		}
		out, err := json.Marshal(first)
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(out, &second); err != nil {
			t.Fatal(err)
		}
		if first.ServerID != second.ServerID || first.ServerAssigned != second.ServerAssigned {
			t.Errorf("%s: round trip %+v -> %s -> %+v", in, first, out, second)
		}
	}
}
