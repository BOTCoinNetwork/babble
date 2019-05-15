package peers

import (
	"bytes"
	"encoding/json"
	"math"

	"github.com/mosaicnetworks/babble/src/common"
	"github.com/mosaicnetworks/babble/src/crypto"
)

//PeerSet is a set of Peers forming a consensus network
type PeerSet struct {
	Peers    []*Peer          `json:"peers"`
	ByPubKey map[string]*Peer `json:"-"`
	ByID     map[uint32]*Peer `json:"-"`

	//cached values
	hash          []byte
	hex           string
	superMajority *int
	trustCount    *int
}

/* Constructors */

//NewPeerSet creates a new PeerSet from a list of Peers
func NewPeerSet(peers []*Peer) *PeerSet {
	peerSet := &PeerSet{
		ByPubKey: make(map[string]*Peer),
		ByID:     make(map[uint32]*Peer),
	}

	for _, peer := range peers {
		peerSet.ByPubKey[peer.PubKeyString()] = peer
		peerSet.ByID[peer.ID()] = peer
	}

	peerSet.Peers = peers

	return peerSet
}

func NewPeerSetFromPeerSliceBytes(peerSliceBytes []byte) (*PeerSet, error) {
	//Decode Peer slice
	peers := []*Peer{}

	b := bytes.NewBuffer(peerSliceBytes)
	dec := json.NewDecoder(b) //will read from b

	err := dec.Decode(&peers)
	if err != nil {
		return nil, err
	}
	//create new PeerSet
	return NewPeerSet(peers), nil
}

//WithNewPeer returns a new PeerSet with a list of peers including the new one.
func (peerSet *PeerSet) WithNewPeer(peer *Peer) *PeerSet {
	peers := append(peerSet.Peers, peer)
	newPeerSet := NewPeerSet(peers)
	return newPeerSet
}

//WithRemovedPeer returns a new PeerSet with a list of peers exluding the
//provided one
func (peerSet *PeerSet) WithRemovedPeer(peer *Peer) *PeerSet {
	peers := []*Peer{}
	for _, p := range peerSet.Peers {
		if p.PubKeyHex != peer.PubKeyHex {
			peers = append(peers, p)
		}
	}
	newPeerSet := NewPeerSet(peers)
	return newPeerSet
}

/* ToSlice Methods */

//PubKeys returns the PeerSet's slice of public keys
func (c *PeerSet) PubKeys() []string {
	res := []string{}

	for _, peer := range c.Peers {
		res = append(res, peer.PubKeyString())
	}

	return res
}

//IDs returns the PeerSet's slice of IDs
func (c *PeerSet) IDs() []uint32 {
	res := []uint32{}

	for _, peer := range c.Peers {
		res = append(res, peer.ID())
	}

	return res
}

/* Utilities */

//Len returns the number of Peers in the PeerSet
func (c *PeerSet) Len() int {
	return len(c.ByPubKey)
}

//Hash uniquely identifies a PeerSet. It is computed by sorting the peers set
//by ID, and hashing (SHA256) their public keys together, one by one.
func (c *PeerSet) Hash() ([]byte, error) {
	if len(c.hash) == 0 {
		hash := []byte{}
		for _, p := range c.Peers {
			pk := p.PubKeyBytes()
			hash = crypto.SimpleHashFromTwoHashes(hash, pk)
		}
		c.hash = hash
	}
	return c.hash, nil
}

//Hex is the hexadecimal representation of Hash
func (c *PeerSet) Hex() string {
	if len(c.hex) == 0 {
		hash, _ := c.Hash()
		c.hex = common.EncodeToString(hash)
	}
	return c.hex
}

func (c *PeerSet) Marshal() ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if err := enc.Encode(c.Peers); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

//SuperMajority return the number of peers that forms a strong majortiy (+2/3)
//in the PeerSet
func (c *PeerSet) SuperMajority() int {
	if c.superMajority == nil {
		val := 2*c.Len()/3 + 1
		c.superMajority = &val
	}
	return *c.superMajority
}

func (c *PeerSet) TrustCount() int {
	if c.trustCount == nil {
		val := 0
		if len(c.Peers) > 1 {
			val = int(math.Ceil(float64(c.Len()) / float64(3)))
		}
		c.trustCount = &val
	}
	return *c.trustCount
}

func (c *PeerSet) clearCache() {
	c.hash = []byte{}
	c.hex = ""
	c.superMajority = nil
}
