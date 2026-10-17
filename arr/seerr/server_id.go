package seerr

import (
	"encoding/json"
	"fmt"
)

// UnmarshalJSON decodes r and sets ServerAssigned when serverId is present and not null.
// Seerr gives its first Sonarr or Radarr server id 0 and stores null for a request without a
// server, so ServerID alone cannot tell the two apart.
func (r *MediaRequest) UnmarshalJSON(data []byte) error {
	type plain MediaRequest
	aux := struct {
		*plain
		ServerID *int `json:"serverId"`
	}{plain: (*plain)(r)}
	if err := json.Unmarshal(data, &aux); err != nil {
		return fmt.Errorf("seerr: decode media request: %w", err)
	}
	r.ServerID, r.ServerAssigned = 0, aux.ServerID != nil
	if aux.ServerID != nil {
		r.ServerID = *aux.ServerID
	}
	return nil
}

// MarshalJSON encodes r with serverId when ServerAssigned is set or ServerID is not 0.
func (r MediaRequest) MarshalJSON() ([]byte, error) {
	type plain MediaRequest
	out, err := json.Marshal(struct {
		plain
		ServerID *int `json:"serverId,omitempty"`
	}{plain: plain(r), ServerID: serverIDOf(r.ServerID, r.ServerAssigned)})
	if err != nil {
		return nil, fmt.Errorf("seerr: encode media request: %w", err)
	}
	return out, nil
}

// MarshalJSON encodes b with serverId when ServerAssigned is set or ServerID is not 0, so a
// request can target server 0.
func (b CreateRequestBody) MarshalJSON() ([]byte, error) {
	type plain CreateRequestBody
	out, err := json.Marshal(struct {
		plain
		ServerID *int `json:"serverId,omitempty"`
	}{plain: plain(b), ServerID: serverIDOf(b.ServerID, b.ServerAssigned)})
	if err != nil {
		return nil, fmt.Errorf("seerr: encode request body: %w", err)
	}
	return out, nil
}

// serverIDOf returns the serverId to encode, or nil to leave it out.
func serverIDOf(id int, assigned bool) *int {
	if !assigned && id == 0 {
		return nil
	}
	return &id
}
