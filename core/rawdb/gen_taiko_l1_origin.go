// Code generated by github.com/fjl/gencodec. DO NOT EDIT.

package rawdb

import (
	"encoding/json"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
)

var _ = (*l1OriginMarshaling)(nil)

// MarshalJSON marshals as JSON.
func (l L1Origin) MarshalJSON() ([]byte, error) {
	type L1Origin struct {
		BlockID       *math.HexOrDecimal256 `json:"blockID" gencodec:"required"`
		L2BlockHash   common.Hash           `json:"l2BlockHash"`
		L1BlockHeight *math.HexOrDecimal256 `json:"l1BlockHeight" rlp:"optional"`
		L1BlockHash   common.Hash           `json:"l1BlockHash" rlp:"optional"`
	}
	var enc L1Origin
	enc.BlockID = (*math.HexOrDecimal256)(l.BlockID)
	enc.L2BlockHash = l.L2BlockHash
	enc.L1BlockHeight = (*math.HexOrDecimal256)(l.L1BlockHeight)
	enc.L1BlockHash = l.L1BlockHash
	return json.Marshal(&enc)
}

// UnmarshalJSON unmarshals from JSON.
func (l *L1Origin) UnmarshalJSON(input []byte) error {
	type L1Origin struct {
		BlockID       *math.HexOrDecimal256 `json:"blockID" gencodec:"required"`
		L2BlockHash   *common.Hash          `json:"l2BlockHash"`
		L1BlockHeight *math.HexOrDecimal256 `json:"l1BlockHeight" rlp:"optional"`
		L1BlockHash   *common.Hash          `json:"l1BlockHash" rlp:"optional"`
	}
	var dec L1Origin
	if err := json.Unmarshal(input, &dec); err != nil {
		return err
	}
	if dec.BlockID == nil {
		return errors.New("missing required field 'blockID' for L1Origin")
	}
	l.BlockID = (*big.Int)(dec.BlockID)
	if dec.L2BlockHash != nil {
		l.L2BlockHash = *dec.L2BlockHash
	}
	if dec.L1BlockHeight != nil {
		l.L1BlockHeight = (*big.Int)(dec.L1BlockHeight)
	}
	if dec.L1BlockHash != nil {
		l.L1BlockHash = *dec.L1BlockHash
	}
	return nil
}
