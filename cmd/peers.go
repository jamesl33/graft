package cmd

import (
	"encoding/json"
	"net/url"

	"github.com/jamesl33/graft/pkg/raft"
	"github.com/pkg/errors"
)

// peers is a slice of peer definitions.
type peers []peer

// newPeers parses peers from the given JSON encdoed data.
func newPeers(data []byte) (peers, error) {
	var peers peers

	if len(data) == 0 {
		return peers, nil
	}

	err := json.Unmarshal(data, &peers)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal peers")
	}

	return peers, nil
}

type peer struct {
	// ID of the peer.
	ID raft.Peer `json:"id"`

	// Address of the peer.
	Address *url.URL `json:"address"`
}

// UnmarshalJSON implements the 'json.Unmarshaler' interface.
func (p *peer) UnmarshalJSON(data []byte) error {
	var overlay struct {
		ID      raft.Peer `json:"id"`
		Address string    `json:"address"`
	}

	err := json.Unmarshal(data, &overlay)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal into overlay")
	}

	parsed, err := url.Parse(overlay.Address)
	if err != nil {
		return errors.Wrap(err, "failed to parse address")
	}

	p.ID = overlay.ID
	p.Address = parsed

	return nil
}
