package statedb

import (
	"bytes"
	"fmt"
	cc "github.com/Fantom-foundation/Carmen/go/common"
	carmen "github.com/Fantom-foundation/Carmen/go/state"
	io2 "github.com/Fantom-foundation/Carmen/go/state/mpt/io"
	"github.com/Fantom-foundation/go-opera/opera/genesis"
	"github.com/Fantom-foundation/go-opera/utils/adapters/kvdb2ethdb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/nokeyiserr"
	"github.com/Fantom-foundation/lachesis-base/kvdb/pebble"
	"github.com/Fantom-foundation/lachesis-base/kvdb/table"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
	"io"
	"os"
	"path/filepath"
)

var emptyCodeHash = crypto.Keccak256(nil)

// IsAlreadyImported checks, if there is already a Carmen directory filled with state data,
// so the EVM data import into it should be skipped. Calling CheckLiveStateHash should follow
// to make sure the directory contains the state with the expected hash.
func (m *StateDbManager) IsAlreadyImported() bool {
	stats, err := os.Stat(m.parameters.Directory)
	return err == nil && stats.IsDir()
}

// ImportWorldState imports Fantom World State data from the genesis file into the Carmen state.
// Must be called before the first StateDbManager.Open call.
func (m *StateDbManager) ImportWorldState(liveReader io.Reader, archiveReader io.Reader, blockNum uint64) error {
	if m.carmenState != nil {
		return fmt.Errorf("carmen state must be closed before the FWS data import")
	}

	liveDir := filepath.Join(m.parameters.Directory, "live")
	if err := os.MkdirAll(liveDir, 0700); err != nil {
		return fmt.Errorf("failed to create carmen dir during FWS import; %v", err)
	}
	if err := io2.ImportLiveDb(liveDir, liveReader); err != nil {
		return fmt.Errorf("failed to import LiveDB; %v", err)
	}

	if m.parameters.Archive == carmen.S5Archive {
		archiveDir := filepath.Join(m.parameters.Directory, "archive")
		if err := os.MkdirAll(archiveDir, 0700); err != nil {
			return fmt.Errorf("failed to create carmen archive dir during FWS import; %v", err)
		}
		if err := io2.InitializeArchive(archiveDir, archiveReader, blockNum); err != nil {
			return fmt.Errorf("failed to initialize Archive; %v", err)
		}
	} else if m.parameters.Archive != carmen.NoArchive {
		return fmt.Errorf("archive is used, but cannot be initialized from FWS genesis section")
	}
	return nil
}

// ImportLegacyEvmData reads legacy EVM trie database and imports one state (for one given block) into Carmen state.
func (m *StateDbManager) ImportLegacyEvmData(evmItems genesis.EvmItems, blockNum uint64, root common.Hash) error {
	if m.carmenState != nil {
		return fmt.Errorf("carmen state must be closed before the legacy EVM data import")
	}
	if err := m.Open(); err != nil {
		return fmt.Errorf("failed to open StateDbManager for legacy EVM data import; %v", err)
	}
	defer m.Close()

	carmenDir, err := os.MkdirTemp("", "opera-tmp-import-legacy-genesis")
	if err != nil {
		panic(fmt.Errorf("failed to create temporary dir for legacy EVM data import: %v", err))
	}
	defer os.RemoveAll(carmenDir)

	m.logger.Log.Info("Unpacking legacy EVM data into a temporary directory", "dir", carmenDir)
	db, err := pebble.New(carmenDir, 1024, 100, nil, nil)
	if err != nil {
		panic(fmt.Errorf("failed to open temporary database for legacy EVM data import: %v", err))
	}
	evmItems.ForEach(func(key, value []byte) bool {
		err := db.Put(key, value)
		if err != nil {
			return false
		}
		return true
	})

	m.logger.Log.Info("Importing legacy EVM data into Carmen", "index", blockNum, "root", root)

	var currentBlock uint64 = 1
	var accountsCount, slotsCount uint64 = 0, 0
	bulk := m.liveStateDb.StartBulkLoad(currentBlock)

	restartBulkIfNeeded := func () error {
		if (accountsCount + slotsCount) % 1_000_000 == 0 && currentBlock < blockNum {
			if err := bulk.Close(); err != nil {
				return err
			}
			currentBlock++
			bulk = m.liveStateDb.StartBulkLoad(currentBlock)
		}
		return nil
	}

	chaindb := rawdb.NewDatabase(kvdb2ethdb.Wrap(nokeyiserr.Wrap(db)))
	triedb := trie.NewDatabase(chaindb)
	t, err := trie.NewSecure(root, triedb)
	if err != nil {
		return fmt.Errorf("failed to open trie; %v", err)
	}
	preimages := table.New(db, []byte("secure-key-"))

	accIter := t.NodeIterator(nil)
	for accIter.Next(true) {
		if accIter.Leaf() {

			addressBytes, err := preimages.Get(accIter.LeafKey())
			if err != nil || addressBytes == nil {
				return fmt.Errorf("missing preimage for account address hash %v; %v", accIter.LeafKey(), err)
			}
			address := cc.Address(common.BytesToAddress(addressBytes))

			var acc state.Account
			if err := rlp.DecodeBytes(accIter.LeafBlob(), &acc); err != nil {
				return fmt.Errorf("invalid account encountered during traversal; %v", err)
			}

			bulk.CreateAccount(address)
			bulk.SetNonce(address, acc.Nonce)
			bulk.SetBalance(address, acc.Balance)


			if !bytes.Equal(acc.CodeHash, emptyCodeHash) {
				code := rawdb.ReadCode(chaindb, common.BytesToHash(acc.CodeHash))
				if len(code) == 0 {
					return fmt.Errorf("missing code for account %v", address)
				}
				bulk.SetCode(address, code)
			}

			if acc.Root != types.EmptyRootHash {
				storageTrie, err := trie.NewSecure(acc.Root, triedb)
				if err != nil {
					return fmt.Errorf("failed to open storage trie for account %v; %v", address, err)
				}
				storageIt := storageTrie.NodeIterator(nil)
				for storageIt.Next(true) {
					if storageIt.Leaf() {
						keyBytes, err := preimages.Get(storageIt.LeafKey())
						if err != nil || keyBytes == nil {
							return fmt.Errorf("missing preimage for storage key hash %v; %v", storageIt.LeafKey(), err)
						}
						key := cc.Key(common.BytesToHash(keyBytes))

						_, valueBytes, _, err := rlp.Split(storageIt.LeafBlob())
						if err != nil {
							return fmt.Errorf("failed to decode storage; %v", err)
						}
						value := cc.Value(common.BytesToHash(valueBytes))

						bulk.SetState(address, key, value)
						slotsCount++
						if err := restartBulkIfNeeded(); err != nil {
							return err
						}
					}
				}
				if storageIt.Error() != nil {
					return fmt.Errorf("failed to iterate storage trie of account %v; %v", address, storageIt.Error())
				}
			}

			accountsCount++
			if err := restartBulkIfNeeded(); err != nil {
				return err
			}
		}
	}
	if accIter.Error() != nil {
		return fmt.Errorf("failed to iterate accounts trie; %v", accIter.Error())
	}

	if err := bulk.Close(); err != nil {
		return err
	}
	// add the empty genesis block into archive
	if currentBlock < blockNum {
		bulk = m.liveStateDb.StartBulkLoad(blockNum)
		if err := bulk.Close(); err != nil {
			return err
		}
	}
	return nil
}

// CheckLiveStateHash reads hash of the Carmen state and compare it with given expected state hash.
// If it does not match, it returns an error.
func (m *StateDbManager) CheckLiveStateHash(blockNum uint64, root common.Hash) error {
	if m.carmenState == nil {
		if err := m.Open(); err != nil {
			return fmt.Errorf("failed to open StateDbManager for live state hash checking; %v", err)
		}
		defer m.Close()
	}
	stateHash := m.liveStateDb.GetHash()
	if cc.Hash(root) != stateHash {
		return fmt.Errorf("hash of the EVM state is incorrect: blockNum: %d expected: %x reproducedHash: %x", blockNum, root, stateHash)
	}
	return nil
}