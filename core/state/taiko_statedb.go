package state

import (
	"github.com/ethereum/go-ethereum/common"
)

func (s Storage) Keys() []common.Hash {
	keys := make([]common.Hash, 0, len(s))
	for key := range s {
		keys = append(keys, key)
	}
	return keys
}

// TouchedAccounts represents the storage of an account at a specific point in time.
type TouchedAccounts map[common.Address][]common.Hash

// TouchedAccounts returns a map of all touched accounts and their storage.
// Incudes self-destructed accounts, loaded accounts and new accounts exclude empty account.
func (s *StateDB) TouchedAccounts() TouchedAccounts {
	touched := make(TouchedAccounts, len(s.stateObjects))
	for addr, obj := range s.stateObjects {
		touched[addr] = obj.originStorage.Keys()
	}
	for addr, obj := range s.stateObjectsDestruct {
		// ignore empty account because it won't affect the state
		if !obj.empty() {
			touched[addr] = obj.originStorage.Keys()
		}
	}
	return touched
}
