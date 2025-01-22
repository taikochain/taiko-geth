package preflight

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/stateless"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/trie/utils"
	"github.com/holiman/uint256"
)

type PreflightStateDB struct {
	*state.StateDB
	touchedAccounts map[common.Address]*types.StateAccount
	ancestors       []common.Hash
}

func (s *PreflightStateDB) CreateAccount(addr common.Address) {
	s.StateDB.CreateAccount(addr)
}

func (s *PreflightStateDB) CreateContract(addr common.Address) {
	s.StateDB.CreateContract(addr)
}

func (s *PreflightStateDB) SubBalance(addr common.Address, amount *uint256.Int, reason tracing.BalanceChangeReason) {
	s.StateDB.SubBalance(addr, amount, reason)
}

func (s *PreflightStateDB) AddBalance(addr common.Address, amount *uint256.Int, reason tracing.BalanceChangeReason) {
	s.StateDB.AddBalance(addr, amount, reason)
}

func (s *PreflightStateDB) GetBalance(addr common.Address) *uint256.Int {
	return s.StateDB.GetBalance(addr)
}

func (s *PreflightStateDB) GetNonce(addr common.Address) uint64 {
	return s.StateDB.GetNonce(addr)
}

func (s *PreflightStateDB) SetNonce(addr common.Address, nonce uint64) {
	s.StateDB.SetNonce(addr, nonce)
}

func (s *PreflightStateDB) GetCodeHash(addr common.Address) common.Hash {
	return s.StateDB.GetCodeHash(addr)
}

func (s *PreflightStateDB) GetCode(addr common.Address) []byte {
	return s.StateDB.GetCode(addr)
}

func (s *PreflightStateDB) SetCode(addr common.Address, code []byte) {
	s.StateDB.SetCode(addr, code)
}

func (s *PreflightStateDB) GetCodeSize(addr common.Address) int {
	return s.StateDB.GetCodeSize(addr)
}

func (s *PreflightStateDB) AddRefund(gas uint64) {
	s.StateDB.AddRefund(gas)
}

func (s *PreflightStateDB) SubRefund(gas uint64) {
	s.StateDB.SubRefund(gas)
}

func (s *PreflightStateDB) GetRefund() uint64 {
	return s.StateDB.GetRefund()
}

func (s *PreflightStateDB) GetCommittedState(addr common.Address, hash common.Hash) common.Hash {
	return s.StateDB.GetCommittedState(addr, hash)
}

func (s *PreflightStateDB) GetState(addr common.Address, hash common.Hash) common.Hash {
	return s.StateDB.GetState(addr, hash)
}

func (s *PreflightStateDB) SetState(addr common.Address, key common.Hash, value common.Hash) {
	s.StateDB.SetState(addr, key, value)
}

func (s *PreflightStateDB) GetStorageRoot(addr common.Address) common.Hash {
	return s.StateDB.GetStorageRoot(addr)
}

func (s *PreflightStateDB) GetTransientState(addr common.Address, key common.Hash) common.Hash {
	return s.StateDB.GetTransientState(addr, key)
}

func (s *PreflightStateDB) SetTransientState(addr common.Address, key common.Hash, value common.Hash) {
	s.StateDB.SetTransientState(addr, key, value)
}

func (s *PreflightStateDB) SelfDestruct(addr common.Address) {
	s.StateDB.SelfDestruct(addr)
}

func (s *PreflightStateDB) HasSelfDestructed(addr common.Address) bool {
	return s.StateDB.HasSelfDestructed(addr)
}

func (s *PreflightStateDB) Selfdestruct6780(addr common.Address) {
	s.StateDB.Selfdestruct6780(addr)
}

// Exist reports whether the given account exists in state.
// Notably this should also return true for self-destructed accounts.
func (s *PreflightStateDB) Exist(addr common.Address) bool {
	return s.StateDB.Exist(addr)
}

// Empty returns whether the given account is empty. Empty
// is defined according to EIP161 (balance = nonce = code = 0).
func (s *PreflightStateDB) Empty(addr common.Address) bool {
	return s.StateDB.Empty(addr)
}

func (s *PreflightStateDB) AddressInAccessList(addr common.Address) bool {
	return s.StateDB.AddressInAccessList(addr)
}

func (s *PreflightStateDB) SlotInAccessList(addr common.Address, slot common.Hash) (addressOk bool, slotOk bool) {
	return s.StateDB.SlotInAccessList(addr, slot)
}

// AddAddressToAccessList adds the given address to the access list. This operation is safe to perform
// even if the feature/fork is not active yet
func (s *PreflightStateDB) AddAddressToAccessList(addr common.Address) {
	s.StateDB.AddAddressToAccessList(addr)
}

// AddSlotToAccessList adds the given (address,slot) to the access list. This operation is safe to perform
// even if the feature/fork is not active yet
func (s *PreflightStateDB) AddSlotToAccessList(addr common.Address, slot common.Hash) {
	s.StateDB.AddSlotToAccessList(addr, slot)
}

// PointCache returns the point cache used in computations
func (s *PreflightStateDB) PointCache() *utils.PointCache {
	return s.StateDB.PointCache()
}

func (s *PreflightStateDB) Prepare(rules params.Rules, sender common.Address, coinbase common.Address, dest *common.Address, precompiles []common.Address, txAccesses types.AccessList) {
	s.StateDB.Prepare(rules, sender, coinbase, dest, precompiles, txAccesses)
}

func (s *PreflightStateDB) RevertToSnapshot(revid int) {
	s.StateDB.RevertToSnapshot(0)
}

func (s *PreflightStateDB) Snapshot() int {
	return s.StateDB.Snapshot()
}

func (s *PreflightStateDB) AddLog(log *types.Log) {
	s.StateDB.AddLog(log)
}

func (s *PreflightStateDB) AddPreimage(hash common.Hash, preimage []byte) {
	s.StateDB.AddPreimage(hash, preimage)
}

func (s *PreflightStateDB) Witness() *stateless.Witness {
	return s.StateDB.Witness()
}
