package impl

import (
	"context"
	"fmt"
	"os"
	"sort"
	"time"

	"go.uber.org/fx"

	"github.com/filecoin-project/dagstore"
	"github.com/filecoin-project/dagstore/shard"
	"github.com/filecoin-project/go-address"
	datatransfer "github.com/filecoin-project/go-data-transfer"
	"github.com/filecoin-project/go-fil-markets/piecestore"
	"github.com/filecoin-project/go-fil-markets/retrievalmarket"
	"github.com/filecoin-project/go-fil-markets/storagemarket"
	"github.com/filecoin-project/go-fil-markets/stores"
	"github.com/filecoin-project/go-jsonrpc/auth"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/ipfs/go-cid"
	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/pkg/errors"

	"github.com/filecoin-project/venus-auth/jwtclient"

	clients2 "github.com/filecoin-project/venus-market/v2/api/clients"
	"github.com/filecoin-project/venus-market/v2/config"
	"github.com/filecoin-project/venus-market/v2/minermgr"
	"github.com/filecoin-project/venus-market/v2/models/repo"
	"github.com/filecoin-project/venus-market/v2/network"
	"github.com/filecoin-project/venus-market/v2/paychmgr"
	"github.com/filecoin-project/venus-market/v2/piecestorage"
	"github.com/filecoin-project/venus-market/v2/retrievalprovider"
	"github.com/filecoin-project/venus-market/v2/storageprovider"
	"github.com/filecoin-project/venus-market/v2/version"

	"github.com/filecoin-project/venus/pkg/constants"
	v1api "github.com/filecoin-project/venus/venus-shared/api/chain/v1"
	gatewayAPIV2 "github.com/filecoin-project/venus/venus-shared/api/gateway/v2"
	marketAPI "github.com/filecoin-project/venus/venus-shared/api/market"
	vTypes "github.com/filecoin-project/venus/venus-shared/types"
	gatewayTypes "github.com/filecoin-project/venus/venus-shared/types/gateway"
	types "github.com/filecoin-project/venus/venus-shared/types/market"
)

var (
	_   marketAPI.IMarket = (*MarketNodeImpl)(nil)
	log                   = logging.Logger("market_api")
)

type MarketNodeImpl struct {
	fx.In

	FundAPI
	gatewayAPIV2.IMarketServiceProvider

	FullNode          v1api.FullNode
	Host              host.Host
	StorageProvider   storageprovider.StorageProvider
	RetrievalProvider retrievalprovider.IRetrievalProvider
	DataTransfer      network.ProviderDataTransfer
	DealPublisher     *storageprovider.DealPublisher
	DealAssigner      storageprovider.DealAssiger

	AuthClient *jwtclient.AuthClient

	Messager                                    clients2.IMixMessage
	StorageAsk                                  storageprovider.IStorageAsk
	DAGStore                                    *dagstore.DAGStore
	DAGStoreWrapper                             stores.DAGStoreWrapper
	PieceStorageMgr                             *piecestorage.PieceStorageManager
	UserMgr                                     minermgr.IMinerMgr
	PaychAPI                                    *paychmgr.PaychAPI
	Repo                                        repo.Repo
	Config                                      *config.MarketConfig
	ConsiderOnlineStorageDealsConfigFunc        config.ConsiderOnlineStorageDealsConfigFunc
	SetConsiderOnlineStorageDealsConfigFunc     config.SetConsiderOnlineStorageDealsConfigFunc
	ConsiderOnlineRetrievalDealsConfigFunc      config.ConsiderOnlineRetrievalDealsConfigFunc
	SetConsiderOnlineRetrievalDealsConfigFunc   config.SetConsiderOnlineRetrievalDealsConfigFunc
	StorageDealPieceCidBlocklistConfigFunc      config.StorageDealPieceCidBlocklistConfigFunc
	SetStorageDealPieceCidBlocklistConfigFunc   config.SetStorageDealPieceCidBlocklistConfigFunc
	ConsiderOfflineStorageDealsConfigFunc       config.ConsiderOfflineStorageDealsConfigFunc
	SetConsiderOfflineStorageDealsConfigFunc    config.SetConsiderOfflineStorageDealsConfigFunc
	ConsiderOfflineRetrievalDealsConfigFunc     config.ConsiderOfflineRetrievalDealsConfigFunc
	SetConsiderOfflineRetrievalDealsConfigFunc  config.SetConsiderOfflineRetrievalDealsConfigFunc
	ConsiderVerifiedStorageDealsConfigFunc      config.ConsiderVerifiedStorageDealsConfigFunc
	SetConsiderVerifiedStorageDealsConfigFunc   config.SetConsiderVerifiedStorageDealsConfigFunc
	ConsiderUnverifiedStorageDealsConfigFunc    config.ConsiderUnverifiedStorageDealsConfigFunc
	SetConsiderUnverifiedStorageDealsConfigFunc config.SetConsiderUnverifiedStorageDealsConfigFunc

	GetExpectedSealDurationFunc config.GetExpectedSealDurationFunc
	SetExpectedSealDurationFunc config.SetExpectedSealDurationFunc

	GetMaxDealStartDelayFunc config.GetMaxDealStartDelayFunc
	SetMaxDealStartDelayFunc config.SetMaxDealStartDelayFunc

	TransferPathFunc    config.TransferPathFunc
	SetTransferPathFunc config.SetTransferPathFunc

	PublishMsgPeriodConfigFunc             config.PublishMsgPeriodConfigFunc
	SetPublishMsgPeriodConfigFunc          config.SetPublishMsgPeriodConfigFunc
	MaxDealsPerPublishMsgFunc              config.MaxDealsPerPublishMsgFunc
	SetMaxDealsPerPublishMsgFunc           config.SetMaxDealsPerPublishMsgFunc
	MaxProviderCollateralMultiplierFunc    config.MaxProviderCollateralMultiplierFunc
	SetMaxProviderCollateralMultiplierFunc config.SetMaxProviderCollateralMultiplierFunc

	MaxPublishDealsFeeFunc        config.MaxPublishDealsFeeFunc
	SetMaxPublishDealsFeeFunc     config.SetMaxPublishDealsFeeFunc
	MaxMarketBalanceAddFeeFunc    config.MaxMarketBalanceAddFeeFunc
	SetMaxMarketBalanceAddFeeFunc config.SetMaxMarketBalanceAddFeeFunc
}

func (m MarketNodeImpl) ResponseMarketEvent(ctx context.Context, resp *gatewayTypes.ResponseEvent) error {
	return m.IMarketServiceProvider.ResponseMarketEvent(ctx, resp)
}

func (m MarketNodeImpl) ListenMarketEvent(ctx context.Context, policy *gatewayTypes.MarketRegisterPolicy) (<-chan *gatewayTypes.RequestEvent, error) {
	return m.IMarketServiceProvider.ListenMarketEvent(ctx, policy)
}

func (m MarketNodeImpl) ActorList(ctx context.Context) ([]types.User, error) {
	return m.UserMgr.ActorList(ctx)
}

func (m MarketNodeImpl) ActorExist(ctx context.Context, addr address.Address) (bool, error) {
	return m.UserMgr.Has(ctx, addr), nil
}

func (m MarketNodeImpl) ActorSectorSize(ctx context.Context, addr address.Address) (abi.SectorSize, error) {
	if bHas := m.UserMgr.Has(ctx, addr); bHas {
		minerInfo, err := m.FullNode.StateMinerInfo(ctx, addr, vTypes.EmptyTSK)
		if err != nil {
			return 0, err
		}

		return minerInfo.SectorSize, nil
	}

	return 0, errors.New("not found")
}

func (m *MarketNodeImpl) MarketImportDealData(ctx context.Context, propCid cid.Cid, path string) error {
	fi, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer fi.Close() //nolint:errcheck

	return m.StorageProvider.ImportDataForDeal(ctx, propCid, fi)
}

func (m *MarketNodeImpl) MarketImportPublishedDeal(ctx context.Context, deal types.MinerDeal) error {
	return m.StorageProvider.ImportPublishedDeal(ctx, deal)
}

func (m *MarketNodeImpl) MarketListDeals(ctx context.Context, addrs []address.Address) ([]*vTypes.MarketDeal, error) {
	return m.listDeals(ctx, addrs)
}

// 检索订单没法按 `miner address` 过滤
func (m *MarketNodeImpl) MarketListRetrievalDeals(ctx context.Context) ([]types.ProviderDealState, error) {
	var out []types.ProviderDealState
	deals, err := m.RetrievalProvider.ListDeals(ctx)
	if err != nil {
		return nil, err
	}

	for _, deal := range deals {
		if deal.ChannelID != nil {
			if deal.ChannelID.Initiator == "" || deal.ChannelID.Responder == "" {
				deal.ChannelID = nil // don't try to push unparsable peer IDs over jsonrpc
			}
		}
		out = append(out, *deal)
	}
	return out, nil
}

func (m *MarketNodeImpl) MarketGetDealUpdates(ctx context.Context) (<-chan types.MinerDeal, error) {
	results := make(chan types.MinerDeal)
	unsub := m.StorageProvider.SubscribeToEvents(func(evt storagemarket.ProviderEvent, deal *types.MinerDeal) {
		select {
		case results <- *deal:
		case <-ctx.Done():
		}
	})
	go func() {
		<-ctx.Done()
		unsub()
		close(results)
	}()
	return results, nil
}

func (m *MarketNodeImpl) MarketListIncompleteDeals(ctx context.Context, mAddr address.Address) ([]types.MinerDeal, error) {
	var deals []*types.MinerDeal
	var err error
	if mAddr == address.Undef {
		deals, err = m.Repo.StorageDealRepo().ListDeal(ctx)
		if err != nil {
			return nil, fmt.Errorf("get deal: %s", err)
		}
	} else {
		deals, err = m.Repo.StorageDealRepo().ListDealByAddr(ctx, mAddr)
		if err != nil {
			return nil, fmt.Errorf("get deal for %s: %s", mAddr.String(), err)
		}
	}

	resDeals := make([]types.MinerDeal, len(deals))
	for idx, deal := range deals {
		resDeals[idx] = *deal
	}

	return resDeals, nil
}

func (m *MarketNodeImpl) UpdateStorageDealStatus(ctx context.Context, dealProposal cid.Cid, state storagemarket.StorageDealStatus, pieceState types.PieceStatus) error {
	return m.Repo.StorageDealRepo().UpdateDealStatus(ctx, dealProposal, state, pieceState)
}

func (m *MarketNodeImpl) MarketSetAsk(ctx context.Context, mAddr address.Address, price vTypes.BigInt, verifiedPrice vTypes.BigInt, duration abi.ChainEpoch, minPieceSize abi.PaddedPieceSize, maxPieceSize abi.PaddedPieceSize) error {
	options := []storagemarket.StorageAskOption{
		storagemarket.MinPieceSize(minPieceSize),
		storagemarket.MaxPieceSize(maxPieceSize),
	}

	return m.StorageAsk.SetAsk(ctx, mAddr, price, verifiedPrice, duration, options...)
}

func (m *MarketNodeImpl) MarketListStorageAsk(ctx context.Context) ([]*types.SignedStorageAsk, error) {
	return m.StorageAsk.ListAsk(ctx)
}

func (m *MarketNodeImpl) MarketGetAsk(ctx context.Context, mAddr address.Address) (*types.SignedStorageAsk, error) {
	return m.StorageAsk.GetAsk(ctx, mAddr)
}

func (m *MarketNodeImpl) MarketSetRetrievalAsk(ctx context.Context, mAddr address.Address, ask *retrievalmarket.Ask) error {
	return m.Repo.RetrievalAskRepo().SetAsk(ctx, &types.RetrievalAsk{
		Miner:                   mAddr,
		PricePerByte:            ask.PricePerByte,
		UnsealPrice:             ask.UnsealPrice,
		PaymentInterval:         ask.PaymentInterval,
		PaymentIntervalIncrease: ask.PaymentIntervalIncrease,
	})
}

func (m *MarketNodeImpl) MarketListRetrievalAsk(ctx context.Context) ([]*types.RetrievalAsk, error) {
	return m.Repo.RetrievalAskRepo().ListAsk(ctx)
}

func (m *MarketNodeImpl) MarketGetRetrievalAsk(ctx context.Context, mAddr address.Address) (*retrievalmarket.Ask, error) {
	ask, err := m.Repo.RetrievalAskRepo().GetAsk(ctx, mAddr)
	if err != nil {
		return nil, err
	}
	return &retrievalmarket.Ask{
		PricePerByte:            ask.PricePerByte,
		UnsealPrice:             ask.UnsealPrice,
		PaymentInterval:         ask.PaymentInterval,
		PaymentIntervalIncrease: ask.PaymentIntervalIncrease,
	}, nil
}

func (m *MarketNodeImpl) MarketListDataTransfers(ctx context.Context) ([]types.DataTransferChannel, error) {
	inProgressChannels, err := m.DataTransfer.InProgressChannels(ctx)
	if err != nil {
		return nil, err
	}

	apiChannels := make([]types.DataTransferChannel, 0, len(inProgressChannels))
	for _, channelState := range inProgressChannels {
		apiChannels = append(apiChannels, types.NewDataTransferChannel(m.Host.ID(), channelState))
	}

	return apiChannels, nil
}

func (m *MarketNodeImpl) MarketDataTransferUpdates(ctx context.Context) (<-chan types.DataTransferChannel, error) {
	channels := make(chan types.DataTransferChannel)

	unsub := m.DataTransfer.SubscribeToEvents(func(evt datatransfer.Event, channelState datatransfer.ChannelState) {
		channel := types.NewDataTransferChannel(m.Host.ID(), channelState)
		select {
		case <-ctx.Done():
		case channels <- channel:
		}
	})

	go func() {
		defer unsub()
		<-ctx.Done()
	}()

	return channels, nil
}

func (m *MarketNodeImpl) MarketRestartDataTransfer(ctx context.Context, transferID datatransfer.TransferID, otherPeer peer.ID, isInitiator bool) error {
	selfPeer := m.Host.ID()
	if isInitiator {
		return m.DataTransfer.RestartDataTransferChannel(ctx, datatransfer.ChannelID{Initiator: selfPeer, Responder: otherPeer, ID: transferID})
	}
	return m.DataTransfer.RestartDataTransferChannel(ctx, datatransfer.ChannelID{Initiator: otherPeer, Responder: selfPeer, ID: transferID})
}

func (m *MarketNodeImpl) MarketCancelDataTransfer(ctx context.Context, transferID datatransfer.TransferID, otherPeer peer.ID, isInitiator bool) error {
	selfPeer := m.Host.ID()
	if isInitiator {
		return m.DataTransfer.CloseDataTransferChannel(ctx, datatransfer.ChannelID{Initiator: selfPeer, Responder: otherPeer, ID: transferID})
	}
	return m.DataTransfer.CloseDataTransferChannel(ctx, datatransfer.ChannelID{Initiator: otherPeer, Responder: selfPeer, ID: transferID})
}

func (m *MarketNodeImpl) MarketPendingDeals(ctx context.Context) ([]types.PendingDealInfo, error) {
	return m.DealPublisher.PendingDeals(), nil
}

func (m *MarketNodeImpl) MarketPublishPendingDeals(ctx context.Context) error {
	m.DealPublisher.ForcePublishPendingDeals()
	return nil
}

func (m *MarketNodeImpl) PiecesListPieces(ctx context.Context) ([]cid.Cid, error) {
	return m.Repo.StorageDealRepo().ListPieceInfoKeys(ctx)
}

func (m *MarketNodeImpl) PiecesListCidInfos(ctx context.Context) ([]cid.Cid, error) {
	return m.Repo.CidInfoRepo().ListCidInfoKeys(ctx)
}

func (m *MarketNodeImpl) PiecesGetPieceInfo(ctx context.Context, pieceCid cid.Cid) (*piecestore.PieceInfo, error) {
	pi, err := m.Repo.StorageDealRepo().GetPieceInfo(ctx, pieceCid)
	if err != nil {
		return nil, err
	}
	return pi, nil
}

func (m *MarketNodeImpl) PiecesGetCIDInfo(ctx context.Context, payloadCid cid.Cid) (*piecestore.CIDInfo, error) {
	ci, err := m.Repo.CidInfoRepo().GetCIDInfo(ctx, payloadCid)
	if err != nil {
		return nil, err
	}

	return &ci, nil
}

func (m *MarketNodeImpl) permissionVerify(ctx context.Context, mAddr address.Address) error {
	if bHas := auth.HasPerm(ctx, []auth.Permission{}, "admin"); bHas {
		return nil
	}

	name, bExist := jwtclient.CtxGetName(ctx)
	if !bExist {
		return fmt.Errorf("token illegal")
	}

	bExist, err := m.AuthClient.MinerExistInUser(name, mAddr.String())
	if err != nil {
		return err
	}

	if !bExist {
		return fmt.Errorf("the token has no permissions to operate this miner")
	}

	return nil
}

func (m *MarketNodeImpl) DealsConsiderOnlineStorageDeals(ctx context.Context, mAddr address.Address) (bool, error) {
	return m.ConsiderOnlineStorageDealsConfigFunc(mAddr)
}

func (m *MarketNodeImpl) DealsSetConsiderOnlineStorageDeals(ctx context.Context, mAddr address.Address, b bool) error {
	if err := m.permissionVerify(ctx, mAddr); err != nil {
		return err
	}
	return m.SetConsiderOnlineStorageDealsConfigFunc(mAddr, b)
}

func (m *MarketNodeImpl) DealsConsiderOnlineRetrievalDeals(ctx context.Context, mAddr address.Address) (bool, error) {
	return m.ConsiderOnlineRetrievalDealsConfigFunc(mAddr)
}

func (m *MarketNodeImpl) DealsSetConsiderOnlineRetrievalDeals(ctx context.Context, mAddr address.Address, b bool) error {
	if err := m.permissionVerify(ctx, mAddr); err != nil {
		return err
	}
	return m.SetConsiderOnlineRetrievalDealsConfigFunc(mAddr, b)
}

func (m *MarketNodeImpl) DealsPieceCidBlocklist(ctx context.Context, mAddr address.Address) ([]cid.Cid, error) {
	return m.StorageDealPieceCidBlocklistConfigFunc(mAddr)
}

func (m *MarketNodeImpl) DealsSetPieceCidBlocklist(ctx context.Context, mAddr address.Address, cids []cid.Cid) error {
	if err := m.permissionVerify(ctx, mAddr); err != nil {
		return err
	}
	return m.SetStorageDealPieceCidBlocklistConfigFunc(mAddr, cids)
}

func (m *MarketNodeImpl) DealsConsiderOfflineStorageDeals(ctx context.Context, mAddr address.Address) (bool, error) {
	return m.ConsiderOfflineStorageDealsConfigFunc(mAddr)
}

func (m *MarketNodeImpl) DealsSetConsiderOfflineStorageDeals(ctx context.Context, mAddr address.Address, b bool) error {
	if err := m.permissionVerify(ctx, mAddr); err != nil {
		return err
	}
	return m.SetConsiderOfflineStorageDealsConfigFunc(mAddr, b)
}

func (m *MarketNodeImpl) DealsConsiderOfflineRetrievalDeals(ctx context.Context, mAddr address.Address) (bool, error) {
	return m.ConsiderOfflineRetrievalDealsConfigFunc(mAddr)
}

func (m *MarketNodeImpl) DealsSetConsiderOfflineRetrievalDeals(ctx context.Context, mAddr address.Address, b bool) error {
	if err := m.permissionVerify(ctx, mAddr); err != nil {
		return err
	}
	return m.SetConsiderOfflineRetrievalDealsConfigFunc(mAddr, b)
}

func (m *MarketNodeImpl) DealsConsiderVerifiedStorageDeals(ctx context.Context, mAddr address.Address) (bool, error) {
	return m.ConsiderVerifiedStorageDealsConfigFunc(mAddr)
}

func (m *MarketNodeImpl) DealsSetConsiderVerifiedStorageDeals(ctx context.Context, mAddr address.Address, b bool) error {
	if err := m.permissionVerify(ctx, mAddr); err != nil {
		return err
	}
	return m.SetConsiderVerifiedStorageDealsConfigFunc(mAddr, b)
}

func (m *MarketNodeImpl) DealsConsiderUnverifiedStorageDeals(ctx context.Context, mAddr address.Address) (bool, error) {
	return m.ConsiderUnverifiedStorageDealsConfigFunc(mAddr)
}

func (m *MarketNodeImpl) DealsSetConsiderUnverifiedStorageDeals(ctx context.Context, mAddr address.Address, b bool) error {
	if err := m.permissionVerify(ctx, mAddr); err != nil {
		return err
	}
	return m.SetConsiderUnverifiedStorageDealsConfigFunc(mAddr, b)
}

func (m *MarketNodeImpl) SectorGetExpectedSealDuration(ctx context.Context, mAddr address.Address) (time.Duration, error) {
	return m.GetExpectedSealDurationFunc(mAddr)
}

func (m *MarketNodeImpl) SectorSetExpectedSealDuration(ctx context.Context, mAddr address.Address, duration time.Duration) error {
	if err := m.permissionVerify(ctx, mAddr); err != nil {
		return err
	}
	return m.SetExpectedSealDurationFunc(mAddr, duration)
}

func (m *MarketNodeImpl) DealsMaxStartDelay(ctx context.Context, mAddr address.Address) (time.Duration, error) {
	return m.GetMaxDealStartDelayFunc(mAddr)
}

func (m *MarketNodeImpl) DealsSetMaxStartDelay(ctx context.Context, mAddr address.Address, duration time.Duration) error {
	if err := m.permissionVerify(ctx, mAddr); err != nil {
		return err
	}
	return m.SetMaxDealStartDelayFunc(mAddr, duration)
}

func (m *MarketNodeImpl) DealsPublishMsgPeriod(ctx context.Context, mAddr address.Address) (time.Duration, error) {
	return m.PublishMsgPeriodConfigFunc(mAddr)
}

func (m *MarketNodeImpl) DealsSetPublishMsgPeriod(ctx context.Context, mAddr address.Address, duration time.Duration) error {
	if err := m.permissionVerify(ctx, mAddr); err != nil {
		return err
	}
	return m.SetPublishMsgPeriodConfigFunc(mAddr, duration)
}

func (m *MarketNodeImpl) MarketDataTransferPath(ctx context.Context, mAddr address.Address) (string, error) {
	return m.TransferPathFunc(mAddr)
}

func (m *MarketNodeImpl) MarketSetDataTransferPath(ctx context.Context, mAddr address.Address, path string) error {
	return m.SetTransferPathFunc(mAddr, path)
}

func (m *MarketNodeImpl) MarketMaxDealsPerPublishMsg(ctx context.Context, mAddr address.Address) (uint64, error) {
	return m.MaxDealsPerPublishMsgFunc(mAddr)
}

func (m *MarketNodeImpl) MarketSetMaxDealsPerPublishMsg(ctx context.Context, mAddr address.Address, num uint64) error {
	if err := m.permissionVerify(ctx, mAddr); err != nil {
		return err
	}
	return m.SetMaxDealsPerPublishMsgFunc(mAddr, num)
}

func (m *MarketNodeImpl) DealsMaxProviderCollateralMultiplier(ctx context.Context, mAddr address.Address) (uint64, error) {
	return m.MaxProviderCollateralMultiplierFunc(mAddr)
}

func (m *MarketNodeImpl) DealsSetMaxProviderCollateralMultiplier(ctx context.Context, mAddr address.Address, c uint64) error {
	if err := m.permissionVerify(ctx, mAddr); err != nil {
		return err
	}
	return m.SetMaxProviderCollateralMultiplierFunc(mAddr, c)
}

func (m *MarketNodeImpl) DealsMaxPublishFee(ctx context.Context, mAddr address.Address) (vTypes.FIL, error) {
	return m.MaxPublishDealsFeeFunc(mAddr)
}

func (m *MarketNodeImpl) DealsSetMaxPublishFee(ctx context.Context, mAddr address.Address, fee vTypes.FIL) error {
	if err := m.permissionVerify(ctx, mAddr); err != nil {
		return err
	}
	return m.SetMaxPublishDealsFeeFunc(mAddr, fee)
}

func (m *MarketNodeImpl) MarketMaxBalanceAddFee(ctx context.Context, mAddr address.Address) (vTypes.FIL, error) {
	return m.MaxMarketBalanceAddFeeFunc(mAddr)
}

func (m *MarketNodeImpl) MarketSetMaxBalanceAddFee(ctx context.Context, mAddr address.Address, fee vTypes.FIL) error {
	if err := m.permissionVerify(ctx, mAddr); err != nil {
		return err
	}
	return m.SetMaxMarketBalanceAddFeeFunc(mAddr, fee)
}

func (m *MarketNodeImpl) MessagerWaitMessage(ctx context.Context, mid cid.Cid) (*vTypes.MsgLookup, error) {
	// WaitMsg method has been replace in messager mode
	return m.Messager.WaitMsg(ctx, mid, constants.MessageConfidence, constants.LookbackNoLimit, false)
}

func (m *MarketNodeImpl) MessagerPushMessage(ctx context.Context, msg *vTypes.Message, meta *vTypes.MessageSendSpec) (cid.Cid, error) {
	var spec *vTypes.MessageSendSpec
	if meta != nil {
		spec = &vTypes.MessageSendSpec{
			MaxFee:            meta.MaxFee,
			GasOverEstimation: meta.GasOverEstimation,
		}
	}
	return m.Messager.PushMessage(ctx, msg, spec)
}

func (m *MarketNodeImpl) MessagerGetMessage(ctx context.Context, mid cid.Cid) (*vTypes.Message, error) {
	return m.Messager.GetMessage(ctx, mid)
}

func (m *MarketNodeImpl) listDeals(ctx context.Context, addrs []address.Address) ([]*vTypes.MarketDeal, error) {
	ts, err := m.FullNode.ChainHead(ctx)
	if err != nil {
		return nil, err
	}

	allDeals, err := m.FullNode.StateMarketDeals(ctx, ts.Key())
	if err != nil {
		return nil, err
	}

	var out []*vTypes.MarketDeal

	has := func(addr address.Address) bool {
		for _, a := range addrs {
			if a == addr {
				return true
			}
		}

		return false
	}

	for _, deal := range allDeals {
		if m.UserMgr.Has(ctx, deal.Proposal.Provider) && has(deal.Proposal.Provider) {
			out = append(out, deal)
		}
	}

	return out, nil
}

func (m *MarketNodeImpl) NetAddrsListen(context.Context) (peer.AddrInfo, error) {
	return peer.AddrInfo{
		ID:    m.Host.ID(),
		Addrs: m.Host.Addrs(),
	}, nil
}

func (m *MarketNodeImpl) ID(context.Context) (peer.ID, error) {
	return m.Host.ID(), nil
}

func (m *MarketNodeImpl) DagstoreListShards(ctx context.Context) ([]types.DagstoreShardInfo, error) {
	info := m.DAGStore.AllShardsInfo()
	ret := make([]types.DagstoreShardInfo, 0, len(info))
	for k, i := range info {
		ret = append(ret, types.DagstoreShardInfo{
			Key:   k.String(),
			State: i.ShardState.String(),
			Error: func() string {
				if i.Error == nil {
					return ""
				}
				return i.Error.Error()
			}(),
		})
	}

	// order by key.
	sort.SliceStable(ret, func(i, j int) bool {
		return ret[i].Key < ret[j].Key
	})

	return ret, nil
}

func (m *MarketNodeImpl) DagstoreInitializeShard(ctx context.Context, key string) error {
	// check whether key valid
	cidKey, err := cid.Decode(key)
	if err != nil {
		return err
	}
	_, err = m.Repo.StorageDealRepo().GetPieceInfo(ctx, cidKey)
	if err != nil {
		return err
	}

	// check whether shard info exit
	k := shard.KeyFromString(key)
	info, err := m.DAGStore.GetShardInfo(k)
	if err != nil && err != dagstore.ErrShardUnknown {
		return fmt.Errorf("failed to get shard info: %w", err)
	}

	if st := info.ShardState; st != dagstore.ShardStateNew {
		return fmt.Errorf("cannot initialize shard; expected state ShardStateNew, was: %s", st.String())
	}

	bs, err := m.DAGStoreWrapper.LoadShard(ctx, cidKey)
	if err != nil {
		return err
	}
	return bs.Close()
}

func (m *MarketNodeImpl) DagstoreInitializeAll(ctx context.Context, params types.DagstoreInitializeAllParams) (<-chan types.DagstoreInitializeAllEvent, error) {
	deals, err := m.Repo.StorageDealRepo().GetDealByAddrAndStatus(ctx, address.Undef, storageprovider.ReadyRetrievalDealStatus...)
	if err != nil {
		return nil, err
	}
	// are we initializing only unsealed pieces?
	onlyUnsealed := !params.IncludeSealed

	var toInitialize []string
	for _, deal := range deals {
		pieceCid := deal.ClientDealProposal.Proposal.PieceCID
		info, err := m.DAGStore.GetShardInfo(shard.KeyFromCID(pieceCid))
		if err != nil && err != dagstore.ErrShardUnknown {
			return nil, err
		}

		if info.ShardState != dagstore.ShardStateNew {
			continue
		}

		// if we're initializing only unsealed pieces, check if there's an
		// unsealed deal for this piece available.
		if onlyUnsealed {
			_, err = m.PieceStorageMgr.FindStorageForRead(ctx, pieceCid.String())
			if err != nil {
				// todo unseal
				log.Warnw("DagstoreInitializeAll: failed to get unsealed status; skipping deal", "piece cid", pieceCid, "error", err)
				continue
			}
		}
		// todo trigger unseal
		// yes, we're initializing this shard.
		toInitialize = append(toInitialize, pieceCid.String())
	}

	return m.dagstoreLoadShards(ctx, toInitialize, params.MaxConcurrency)
}

func (m *MarketNodeImpl) DagstoreInitializeStorage(ctx context.Context, storageName string, params types.DagstoreInitializeAllParams) (<-chan types.DagstoreInitializeAllEvent, error) {
	storage, err := m.PieceStorageMgr.GetPieceStorageByName(storageName)
	if err != nil {
		return nil, err
	}
	resourceIds, err := storage.ListResourceIds(ctx)
	if err != nil {
		return nil, err
	}

	var toInitialize []string
	for _, resource := range resourceIds {
		pieceCid, err := cid.Decode(resource)
		if err != nil {
			log.Warnf("resource name (%s) was not a valid piece cid %v", resource, err)
			continue
		}
		pieceInfo, err := m.Repo.StorageDealRepo().GetPieceInfo(ctx, pieceCid)
		if err != nil || (pieceInfo != nil && len(pieceInfo.Deals) == 0) {
			log.Warnf("piece cid %s not in storage deals", pieceCid)
			continue
		}

		_, err = m.DAGStore.GetShardInfo(shard.KeyFromString(resource))
		if err != nil && !errors.Is(err, dagstore.ErrShardUnknown) {
			return nil, err
		}

		toInitialize = append(toInitialize, resource)
	}

	return m.dagstoreLoadShards(ctx, toInitialize, params.MaxConcurrency)
}

func (m *MarketNodeImpl) dagstoreLoadShards(ctx context.Context, toInitialize []string, concurrency int) (<-chan types.DagstoreInitializeAllEvent, error) {
	// prepare the thottler tokens.
	var throttle chan struct{}
	if c := concurrency; c > 0 {
		throttle = make(chan struct{}, c)
		for i := 0; i < c; i++ {
			throttle <- struct{}{}
		}
	}

	total := len(toInitialize)
	if total == 0 {
		out := make(chan types.DagstoreInitializeAllEvent)
		close(out)
		return out, nil
	}

	// response channel must be closed when we're done, or the context is cancelled.
	// this buffering is necessary to prevent inflight children goroutines from
	// publishing to a closed channel (res) when the context is cancelled.
	out := make(chan types.DagstoreInitializeAllEvent, 32) // internal buffer.
	res := make(chan types.DagstoreInitializeAllEvent, 32) // returned to caller.

	// pump events back to caller.
	// two events per shard.
	go func() {
		defer close(res)

		for i := 0; i < total*2; i++ {
			select {
			case res <- <-out:
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		for i, k := range toInitialize {
			if throttle != nil {
				select {
				case <-throttle:
					// acquired a throttle token, proceed.
				case <-ctx.Done():
					return
				}
			}

			go func(k string, i int) {
				r := types.DagstoreInitializeAllEvent{
					Key:     k,
					Event:   "start",
					Total:   total,
					Current: i + 1, // start with 1
				}
				select {
				case out <- r:
				case <-ctx.Done():
					return
				}

				err := m.DagstoreInitializeShard(ctx, k)

				if throttle != nil {
					throttle <- struct{}{}
				}

				r.Event = "end"
				if err == nil {
					r.Success = true
				} else {
					r.Success = false
					r.Error = err.Error()
				}

				select {
				case out <- r:
				case <-ctx.Done():
				}
			}(k, i)
		}
	}()

	return res, nil
}

func (m *MarketNodeImpl) DagstoreRecoverShard(ctx context.Context, key string) error {
	k := shard.KeyFromString(key)

	info, err := m.DAGStore.GetShardInfo(k)
	if err != nil {
		return fmt.Errorf("failed to get shard info: %w", err)
	}
	if st := info.ShardState; st != dagstore.ShardStateErrored {
		return fmt.Errorf("cannot recover shard; expected state ShardStateErrored, was: %s", st.String())
	}

	ch := make(chan dagstore.ShardResult, 1)
	if err = m.DAGStore.RecoverShard(ctx, k, ch, dagstore.RecoverOpts{}); err != nil {
		return fmt.Errorf("failed to recover shard: %w", err)
	}

	var res dagstore.ShardResult
	select {
	case res = <-ch:
	case <-ctx.Done():
		return ctx.Err()
	}

	return res.Error
}

func (m *MarketNodeImpl) DagstoreGC(ctx context.Context) ([]types.DagstoreShardResult, error) {
	if m.DAGStore == nil {
		return nil, fmt.Errorf("dagstore not available on this node")
	}

	res, err := m.DAGStore.GC(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to gc: %w", err)
	}

	ret := make([]types.DagstoreShardResult, 0, len(res.Shards))
	for k, err := range res.Shards {
		r := types.DagstoreShardResult{Key: k.String()}
		if err == nil {
			r.Success = true
		} else {
			r.Success = false
			r.Error = err.Error()
		}
		ret = append(ret, r)
	}

	return ret, nil
}

func (m *MarketNodeImpl) GetUnPackedDeals(ctx context.Context, miner address.Address, spec *types.GetDealSpec) ([]*types.DealInfoIncludePath, error) {
	return m.DealAssigner.GetUnPackedDeals(ctx, miner, spec)
}

func (m *MarketNodeImpl) AssignUnPackedDeals(ctx context.Context, sid abi.SectorID, ssize abi.SectorSize, spec *types.GetDealSpec) ([]*types.DealInfoIncludePath, error) {
	return m.DealAssigner.AssignUnPackedDeals(ctx, sid, ssize, spec)
}

func (m *MarketNodeImpl) MarkDealsAsPacking(ctx context.Context, miner address.Address, deals []abi.DealID) error {
	return m.DealAssigner.MarkDealsAsPacking(ctx, miner, deals)
}

func (m *MarketNodeImpl) UpdateDealOnPacking(ctx context.Context, miner address.Address, dealId abi.DealID, sectorid abi.SectorNumber, offset abi.PaddedPieceSize) error {
	return m.DealAssigner.UpdateDealOnPacking(ctx, miner, dealId, sectorid, offset)
}

func (m *MarketNodeImpl) UpdateDealStatus(ctx context.Context, miner address.Address, dealId abi.DealID, status types.PieceStatus) error {
	return m.DealAssigner.UpdateDealStatus(ctx, miner, dealId, status)
}

func (m *MarketNodeImpl) DealsImportData(ctx context.Context, dealPropCid cid.Cid, fname string) error {
	fi, err := os.Open(fname)
	if err != nil {
		return fmt.Errorf("failed to open given file: %w", err)
	}
	defer fi.Close() //nolint:errcheck

	return m.StorageProvider.ImportDataForDeal(ctx, dealPropCid, fi)
}

func (m *MarketNodeImpl) GetDeals(ctx context.Context, miner address.Address, pageIndex, pageSize int) ([]*types.DealInfo, error) {
	return m.DealAssigner.GetDeals(ctx, miner, pageIndex, pageSize)
}

func (m *MarketNodeImpl) PaychVoucherList(ctx context.Context, pch address.Address) ([]*vTypes.SignedVoucher, error) {
	return m.PaychAPI.PaychVoucherList(ctx, pch)
}

func (m *MarketNodeImpl) AddFsPieceStorage(ctx context.Context, name string, path string, readonly bool) error {
	ifs := &config.FsPieceStorage{ReadOnly: readonly, Path: path, Name: name}
	fsps, err := piecestorage.NewFsPieceStorage(ifs)
	if err != nil {
		return err
	}
	// add in memory
	err = m.PieceStorageMgr.AddPieceStorage(fsps)
	if err != nil {
		return err
	}

	// add to config
	return m.Config.AddFsPieceStorage(ifs)
}

func (m *MarketNodeImpl) AddS3PieceStorage(ctx context.Context, name, endpoit, bucket, subdir, accessKeyID, secretAccessKey, token string, readonly bool) error {
	ifs := &config.S3PieceStorage{
		ReadOnly:  readonly,
		EndPoint:  endpoit,
		Name:      name,
		Bucket:    bucket,
		SubDir:    subdir,
		AccessKey: accessKeyID,
		SecretKey: secretAccessKey,
		Token:     token,
	}
	s3ps, err := piecestorage.NewS3PieceStorage(ifs)
	if err != nil {
		return err
	}
	// add in memory
	err = m.PieceStorageMgr.AddPieceStorage(s3ps)
	if err != nil {
		return err
	}

	// add to config
	return m.Config.AddS3PieceStorage(ifs)
}

func (m *MarketNodeImpl) ListPieceStorageInfos(ctx context.Context) types.PieceStorageInfos {
	return m.PieceStorageMgr.ListStorageInfos()
}

func (m *MarketNodeImpl) RemovePieceStorage(ctx context.Context, name string) error {
	err := m.PieceStorageMgr.RemovePieceStorage(name)
	if err != nil {
		return err
	}

	return m.Config.RemovePieceStorage(name)
}

func (m *MarketNodeImpl) OfflineDealImport(ctx context.Context, deal types.MinerDeal) error {
	return m.StorageProvider.ImportOfflineDeal(ctx, deal)
}

func (m *MarketNodeImpl) Version(ctx context.Context) (vTypes.Version, error) {
	return vTypes.Version{Version: version.UserVersion()}, nil
}

func (m *MarketNodeImpl) GetStorageDealStatistic(ctx context.Context, miner address.Address) (*types.StorageDealStatistic, error) {
	statistic, err := m.Repo.StorageDealRepo().GroupStorageDealNumberByStatus(ctx, miner)
	if err != nil {
		return nil, err
	}
	return &types.StorageDealStatistic{DealsStatus: statistic}, nil
}

func (m *MarketNodeImpl) GetRetrievalDealStatistic(ctx context.Context, miner address.Address) (*types.RetrievalDealStatistic, error) {
	statistic, err := m.Repo.RetrievalDealRepo().GroupRetrievalDealNumberByStatus(ctx, miner)
	if err != nil {
		return nil, err
	}
	return &types.RetrievalDealStatistic{DealsStatus: statistic}, nil
}
