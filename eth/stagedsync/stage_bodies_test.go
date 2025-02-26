package stagedsync

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/ledgerwatch/erigon-lib/common/u256"
	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon-lib/kv/memdb"
	"github.com/ledgerwatch/erigon/common"
	"github.com/ledgerwatch/erigon/core/rawdb"
	"github.com/ledgerwatch/erigon/core/types"
	"github.com/stretchr/testify/require"
)

func TestBodiesUnwind(t *testing.T) {
	require, ctx := require.New(t), context.Background()
	_, tx := memdb.NewTestTx(t)
	txn := &types.DynamicFeeTransaction{Tip: u256.N1, FeeCap: u256.N1, CommonTx: types.CommonTx{ChainID: u256.N1, Value: u256.N1, Gas: 1, Nonce: 1}}
	buf := bytes.NewBuffer(nil)
	err := txn.MarshalBinary(buf)
	require.NoError(err)
	rlpTxn := buf.Bytes()

	logEvery := time.NewTicker(logInterval)
	defer logEvery.Stop()

	b := &types.RawBody{Transactions: [][]byte{rlpTxn, rlpTxn, rlpTxn}}
	for i := uint64(1); i <= 10; i++ {
		err = rawdb.WriteRawBody(tx, common.Hash{byte(i)}, i, b)
		require.NoError(err)
		err = rawdb.WriteCanonicalHash(tx, common.Hash{byte(i)}, i)
		require.NoError(err)
	}
	{
		err = rawdb.MakeBodiesNonCanonical(tx, 5+1, ctx, "test", logEvery) // block 5 already canonical, start from next one
		require.NoError(err)

		n, err := tx.ReadSequence(kv.EthTx)
		require.NoError(err)
		require.Equal(5*(3+2), int(n)) // from 0, 5 block with 3 txn in each
	}
	{
		err = rawdb.MakeBodiesCanonical(tx, 5+1, ctx, "test", logEvery) // block 5 already canonical, start from next one
		require.NoError(err)
		n, err := tx.ReadSequence(kv.EthTx)
		require.NoError(err)
		require.Equal(10*(3+2), int(n))

		err = rawdb.WriteRawBody(tx, common.Hash{11}, 11, b)
		require.NoError(err)
		err = rawdb.WriteCanonicalHash(tx, common.Hash{11}, 11)
		require.NoError(err)

		n, err = tx.ReadSequence(kv.EthTx)
		require.NoError(err)
		require.Equal(11*(3+2), int(n))
	}

	{
		// unwind to block 5, means mark blocks >= 6 as non-canonical
		err = rawdb.MakeBodiesNonCanonical(tx, 5+1, ctx, "test", logEvery)
		require.NoError(err)

		n, err := tx.ReadSequence(kv.EthTx)
		require.NoError(err)
		require.Equal(5*(3+2), int(n)) // from 0, 5 block with 3 txn in each

		err = rawdb.MakeBodiesCanonical(tx, 5+1, ctx, "test", logEvery) // block 5 already canonical, start from next one
		require.NoError(err)
		n, err = tx.ReadSequence(kv.EthTx)
		require.NoError(err)
		require.Equal(11*(3+2), int(n))
	}
}
