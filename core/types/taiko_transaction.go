package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// To make taiko-geth compatible with op-service.
const DepositTxType = 0x7E

// To make taiko-geth compatible with op-service.
type DepositTx struct {
	// SourceHash uniquely identifies the source of the deposit
	SourceHash common.Hash
	// From is exposed through the types.Signer, not through TxData
	From common.Address
	// nil means contract creation
	To *common.Address `rlp:"nil"`
	// Mint is minted on L2, locked on L1, nil if no minting.
	Mint *big.Int `rlp:"nil"`
	// Value is transferred from L2 balance, executed after Mint (if any)
	Value *big.Int
	// gas limit
	Gas uint64
	// Field indicating if this transaction is exempt from the L2 gas limit.
	IsSystemTransaction bool
	// Normal Tx data
	Data []byte
}

func (tx *Transaction) MarkAsAnchor() error {
	return tx.inner.markAsAnchor()
}

func (tx *Transaction) IsAnchor() bool {
	return tx.inner.isAnchor()
}

func (tx *DynamicFeeTx) isAnchor() bool {
	return tx.isAnhcor
}

func (tx *LegacyTx) isAnchor() bool {
	return false
}

func (tx *AccessListTx) isAnchor() bool {
	return false
}

func (tx *BlobTx) isAnchor() bool {
	return false
}

func (tx *DynamicFeeTx) markAsAnchor() error {
	tx.isAnhcor = true
	return nil
}

func (tx *LegacyTx) markAsAnchor() error {
	return ErrInvalidTxType
}

func (tx *AccessListTx) markAsAnchor() error {
	return ErrInvalidTxType
}

func (tx *BlobTx) markAsAnchor() error {
	return ErrInvalidTxType
}
