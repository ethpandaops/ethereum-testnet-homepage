package ethereum

import (
	"context"
	"sync"
	"time"

	"github.com/ethpandaops/beacon/pkg/beacon"
	"github.com/go-co-op/gocron"
	"github.com/sirupsen/logrus"
)

type Node struct {
	infoMutex sync.RWMutex
	info      NodeInfo
	beacon    beacon.Node

	scheduler *gocron.Scheduler
}

//nolint:gocritic // Not concerned about this amount of data
func NewNode(ctx context.Context, log logrus.FieldLogger, name, beaconURL, rpcURL string, info NodeInfo) *Node {
	opts := *beacon.DefaultOptions()
	opts.BeaconSubscription.Topics = []string{"block"}

	return &Node{
		infoMutex: sync.RWMutex{},
		info:      info,
		scheduler: gocron.NewScheduler(time.Local),

		beacon: beacon.NewNode(log.WithField("node", name), &beacon.Config{
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

	return nil
}

func (n *Node) Stop(ctx context.Context) error {
	if err := n.beacon.Stop(ctx); err != nil {
		return err
	}

	return nil
}
