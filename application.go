package kvstore

import (
	"bytes"

	"github.com/dgraph-io/badger/v3"
	abcitypes "github.com/tendermint/tendermint/abci/types"
)

type Application struct {
	db           *badger.DB
	currentBatch *badger.Txn
}

var _ abcitypes.Application = (*Application)(nil)

func NewApplication(db *badger.DB) *Application {
	return &Application{db: db}
}

func (Application) Info(req abcitypes.RequestInfo) abcitypes.ResponseInfo {
	return abcitypes.ResponseInfo{}
}

func (app *Application) DeliverTx(req abcitypes.RequestDeliverTx) abcitypes.ResponseDeliverTx {
	code := app.isValid(req.Tx)
	if code != 0 {
		return abcitypes.ResponseDeliverTx{Code: code}
	}

	parts := bytes.Split(req.Tx, []byte("="))
	key, value := parts[0], parts[1]

	err := app.currentBatch.Set(key, value)
	if err != nil {
		panic(err)
	}

	return abcitypes.ResponseDeliverTx{Code: 0}
}

func (app *Application) CheckTx(req abcitypes.RequestCheckTx) abcitypes.ResponseCheckTx {
	code := app.isValid(req.Tx)
	return abcitypes.ResponseCheckTx{Code: code, GasWanted: 1}
}

func (app *Application) isValid(tx []byte) (code uint32) {
	// check format
	parts := bytes.Split(tx, []byte("="))
	if len(parts) != 2 {
		return 1
	}

	key, value := parts[0], parts[1]

	// check if the same key=value already exists
	err := app.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil && err != badger.ErrKeyNotFound {
			return err
		}
		if err == nil {
			return item.Value(func(val []byte) error {
				if bytes.Equal(val, value) {
					code = 2
				}
				return nil
			})
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	return code
}

func (app *Application) Commit() abcitypes.ResponseCommit {
	err := app.currentBatch.Commit()
	if err != nil {
		panic(err)
	}

	return abcitypes.ResponseCommit{Data: []byte{}}
}

func (app *Application) Query(req abcitypes.RequestQuery) (res abcitypes.ResponseQuery) {
	res.Key = req.Data
	err := app.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(req.Data)
		if err != nil && err != badger.ErrKeyNotFound {
			return err
		}
		if err == badger.ErrKeyNotFound {
			res.Log = "does not exist"
		} else {
			return item.Value(func(val []byte) error {
				res.Log = "exists"
				res.Value = val
				return nil
			})
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	return
}

func (Application) InitChain(req abcitypes.RequestInitChain) abcitypes.ResponseInitChain {
	return abcitypes.ResponseInitChain{}
}

func (app *Application) BeginBlock(req abcitypes.RequestBeginBlock) abcitypes.ResponseBeginBlock {
	app.currentBatch = app.db.NewTransaction(true)
	return abcitypes.ResponseBeginBlock{}
}

func (Application) EndBlock(req abcitypes.RequestEndBlock) abcitypes.ResponseEndBlock {
	return abcitypes.ResponseEndBlock{}
}

func (Application) ListSnapshots(abcitypes.RequestListSnapshots) abcitypes.ResponseListSnapshots {
	return abcitypes.ResponseListSnapshots{}
}

func (Application) OfferSnapshot(abcitypes.RequestOfferSnapshot) abcitypes.ResponseOfferSnapshot {
	return abcitypes.ResponseOfferSnapshot{}
}

func (Application) LoadSnapshotChunk(abcitypes.RequestLoadSnapshotChunk) abcitypes.ResponseLoadSnapshotChunk {
	return abcitypes.ResponseLoadSnapshotChunk{}
}

func (Application) ApplySnapshotChunk(abcitypes.RequestApplySnapshotChunk) abcitypes.ResponseApplySnapshotChunk {
	return abcitypes.ResponseApplySnapshotChunk{}
}
