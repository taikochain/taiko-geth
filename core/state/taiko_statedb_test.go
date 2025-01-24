package state

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/holiman/uint256"
)

// TestCopy tests that copying a StateDB object indeed makes the original and
// the copy independent of each other. This test is a regression test against
// https://github.com/ethereum/go-ethereum/pull/15549.
func TestTouchedAccounts(t *testing.T) {
	// Create a random state test to copy and modify "independently"
	orig, _ := New(types.EmptyRootHash, NewDatabaseForTesting())

	accountCounts := 10

	for i := byte(0); i < byte(accountCounts); i++ {
		obj := orig.getOrNewStateObject(common.BytesToAddress([]byte{i}))
		obj.AddBalance(uint256.NewInt(uint64(i)), tracing.BalanceChangeUnspecified)
		orig.updateStateObject(obj)
	}
	orig.Finalise(false)

	touchedAccs := orig.TouchedAccounts()
	if have, want := len(touchedAccs), accountCounts; have != want {
		t.Fatalf("have %d touched accounts, want %d", have, want)
	}

	// destroy accounts
	selfDestruct1 := common.BytesToAddress([]byte{1})
	selfDestruct2 := common.BytesToAddress([]byte{2})
	selfDestruct3 := common.BytesToAddress([]byte{3})

	orig.SelfDestruct(selfDestruct1)
	orig.SelfDestruct(selfDestruct2)
	orig.SelfDestruct(selfDestruct3)
	orig.Finalise(false)

	touchedAccs = orig.TouchedAccounts()
	if have, want := len(touchedAccs), accountCounts; have != want {
		t.Fatalf("have %d touched accounts, want %d", have, want)
	}

	// create empty accounts
	for i := byte(10); i < 12; i++ {
		orig.getOrNewStateObject(common.BytesToAddress([]byte{i}))
	}
	orig.Finalise(true)

	touchedAccs = orig.TouchedAccounts()
	if have, want := len(touchedAccs), accountCounts; have != want {
		t.Fatalf("have %d touched accounts, want %d", have, want)
	}
}
