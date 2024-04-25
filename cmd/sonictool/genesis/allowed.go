package genesis

import (
	"github.com/Fantom-foundation/go-opera/opera"
	"github.com/Fantom-foundation/go-opera/opera/genesis"
	"github.com/Fantom-foundation/lachesis-base/hash"
)

type GenesisTemplate struct {
	Name   string
	Header genesis.Header
	Hashes genesis.Hashes
}

var (
	testnetHeader = genesis.Header{
		GenesisID:   hash.HexToHash("0x4c4fdf6346c8851355eb305399d05036a0512aea15d1cb9b364a353704d5fbcb"),
		NetworkID:   opera.TestNetworkID,
		NetworkName: "x1-testnet",
	}

	allowedLegacyGenesis = []GenesisTemplate{}
)
