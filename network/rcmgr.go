package network

import (
	"context"
	"errors"
	"fmt"
	"math/bits"
	"os"
	"path/filepath"

	"github.com/ipfs-force-community/droplet/v2/config"

	"go.uber.org/fx"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/network"
	rcmgr "github.com/libp2p/go-libp2p/p2p/host/resource-manager"
)

func ResourceManager(lc fx.Lifecycle, homeDir *config.HomeDir) (network.ResourceManager, error) {
	repoPath := string(*homeDir)
	// Adjust default defaultLimits
	// - give it more memory, up to 4G, min of 1G
	// - if maxconns are too high, adjust Conn/FD/Stream defaultLimits
	defaultLimits := rcmgr.DefaultLimits

	// TODO: also set appropriate default limits for lotus protocols
	libp2p.SetDefaultServiceLimits(&defaultLimits)

	// Minimum 1GB of memory
	defaultLimits.SystemBaseLimit.Memory = 1 << 30
	// For every extra 1GB of memory we have available, increase our limit by 1GiB
	defaultLimits.SystemLimitIncrease.Memory = 1 << 30
	defaultLimitConfig := defaultLimits.AutoScale()

	changes := rcmgr.PartialLimitConfig{}

	if defaultLimitConfig.ToPartialLimitConfig().System.Memory > 4<<30 {
		// Cap our memory limit
		changes.System.Memory = 4 << 30
	}

	maxconns := int(200) // make config
	if rcmgr.LimitVal(2*maxconns) > defaultLimitConfig.ToPartialLimitConfig().System.ConnsInbound {
		// adjust conns to 2x to allow for two conns per peer (TCP+QUIC)
		changes.System.ConnsInbound = rcmgr.LimitVal(logScale(2 * maxconns))
		changes.System.ConnsOutbound = rcmgr.LimitVal(logScale(2 * maxconns))
		changes.System.Conns = rcmgr.LimitVal(logScale(4 * maxconns))

		changes.System.StreamsInbound = rcmgr.LimitVal(logScale(16 * maxconns))
		changes.System.StreamsOutbound = rcmgr.LimitVal(logScale(64 * maxconns))
		changes.System.Streams = rcmgr.LimitVal(logScale(64 * maxconns))

		if rcmgr.LimitVal(2*maxconns) > defaultLimitConfig.ToPartialLimitConfig().System.FD {
			changes.System.FD = rcmgr.LimitVal(logScale(2 * maxconns))
		}

		changes.ServiceDefault.StreamsInbound = rcmgr.LimitVal(logScale(8 * maxconns))
		changes.ServiceDefault.StreamsOutbound = rcmgr.LimitVal(logScale(32 * maxconns))
		changes.ServiceDefault.Streams = rcmgr.LimitVal(logScale(32 * maxconns))

		changes.ProtocolDefault.StreamsInbound = rcmgr.LimitVal(logScale(8 * maxconns))
		changes.ProtocolDefault.StreamsOutbound = rcmgr.LimitVal(logScale(32 * maxconns))
		changes.ProtocolDefault.Streams = rcmgr.LimitVal(logScale(32 * maxconns))

		log.Info("adjusted default resource manager limits")
	}

	changedLimitConfig := changes.Build(defaultLimitConfig)
	// initialize
	var limiter rcmgr.Limiter
	var opts []rcmgr.Option

	// create limiter -- parse $repo/limits.json if exists
	limitsFile := filepath.Join(repoPath, "limits.json")
	limitsIn, err := os.Open(limitsFile)
	switch {
	case err == nil:
		defer limitsIn.Close() //nolint:errcheck
		limiter, err = rcmgr.NewLimiterFromJSON(limitsIn, changedLimitConfig)
		if err != nil {
			return nil, fmt.Errorf("error parsing limit file: %w", err)
		}

	case errors.Is(err, os.ErrNotExist):
		limiter = rcmgr.NewFixedLimiter(changedLimitConfig)

	default:
		return nil, err
	}

	if os.Getenv("MARKET_DEBUG_RCMGR") != "" {
		debugPath := filepath.Join(repoPath, "debug")
		if err := os.MkdirAll(debugPath, 0o755); err != nil {
			return nil, fmt.Errorf("error creating debug directory: %w", err)
		}
		traceFile := filepath.Join(debugPath, "rcmgr.json.gz")
		opts = append(opts, rcmgr.WithTrace(traceFile))
	}

	mgr, err := rcmgr.NewResourceManager(limiter, opts...)
	if err != nil {
		return nil, fmt.Errorf("error creating resource manager: %w", err)
	}

	lc.Append(fx.Hook{
		OnStop: func(_ context.Context) error {
			return mgr.Close()
		},
	})

	return mgr, nil
}

func ResourceManagerOption(mgr network.ResourceManager) Libp2pOpts {
	return Libp2pOpts{
		Opts: []libp2p.Option{libp2p.ResourceManager(mgr)},
	}
}

func logScale(val int) int {
	bitlen := bits.Len(uint(val))
	return 1 << bitlen
}
