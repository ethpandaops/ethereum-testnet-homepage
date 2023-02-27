package ethereum

import (
	"context"
	"sync"
	"time"

	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/ethpandaops/beacon/pkg/beacon"
	"github.com/ethpandaops/beacon/pkg/beacon/api/types"
	"github.com/go-co-op/gocron"
	"github.com/sirupsen/logrus"
)

type Node struct {
	log       logrus.FieldLogger
	infoMutex sync.RWMutex
	info      NodeInfo
	beacon    beacon.Node

	ConsensusHead                *v1.HeadEvent
	ConsensusFinalizedCheckpoint *v1.Finality
	ConsensusPeers               *types.Peers

	scheduler *gocron.Scheduler
}

//nolint:gocritic // Not concerned about this amount of data
func NewNode(ctx context.Context, log logrus.FieldLogger, name, beaconURL, rpcURL string, info NodeInfo) *Node {
	opts := *beacon.DefaultOptions()
	opts.BeaconSubscription.
		Enable().
		Topics = []string{"block", "head"}

	log = log.WithField("node", name)

	return &Node{
		log:       log.WithField("module", "service/ethereum/node"),
		infoMutex: sync.RWMutex{},
		info:      info,
		scheduler: gocron.NewScheduler(time.Local),

		beacon: beacon.NewNode(log, &beacon.Config{
			Addr: beaconURL,
			Name: name,
		}, "ethereum_testnet_homepage", opts),
	}
}

func (n *Node) Info() NodeInfo {
	n.infoMutex.RLock()
	defer n.infoMutex.RUnlock()

	return n.info
}

//nolint:gocritic // Not concerned about this amount of data
func (n *Node) UpdateInfo(info NodeInfo) {
	n.infoMutex.Lock()
	defer n.infoMutex.Unlock()

	n.info = info
}

func (n *Node) Start(ctx context.Context) error {
	n.beacon.StartAsync(ctx)

	n.beacon.OnHead(ctx, func(ctx context.Context, head *v1.HeadEvent) error {
		n.ConsensusHead = head

		return nil
	})

	n.beacon.OnFinalityCheckpointUpdated(ctx, func(ctx context.Context, event *beacon.FinalityCheckpointUpdated) error {
		n.ConsensusFinalizedCheckpoint = event.Finality

		return nil
	})

	n.beacon.OnPeersUpdated(ctx, func(ctx context.Context, event *beacon.PeersUpdatedEvent) error {
		n.ConsensusPeers = &event.Peers

		return nil
	})

	if _, err := n.scheduler.Every("60s").Do(func() {
		peers, err := n.beacon.FetchPeers(ctx)
		if err != nil {
			n.log.WithError(err).Error("Failed to get peer count")

			return
		}

		n.ConsensusPeers = peers
	}); err != nil {
		return err
	}

	return nil
}

func (n *Node) Stop(ctx context.Context) error {
	if err := n.beacon.Stop(ctx); err != nil {
		return err
	}

	return nil
}
