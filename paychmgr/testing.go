package paychmgr

import (
	"github.com/ipfs-force-community/droplet/v2/models/badger"
	"github.com/ipfs-force-community/droplet/v2/models/repo"
	ds "github.com/ipfs/go-datastore"
	ds_sync "github.com/ipfs/go-datastore/sync"
)

func newRepo() repo.Repo {
	paychDs := ds_sync.MutexWrap(ds.NewMapDatastore())
	params := badger.BadgerDSParams{
		PaychInfoDS: badger.NewPayChanDS(paychDs),
		PaychMsgDS:  badger.NewPayChanMsgDs(paychDs),
	}
	return badger.NewBadgerRepo(params)
}
