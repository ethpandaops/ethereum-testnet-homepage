package ethereum

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/sirupsen/logrus"
)

type InventoryWatcher struct {
	log        logrus.FieldLogger
	config     *InventoryConfig
	nodesMutex sync.RWMutex
	nodes      Nodes
}

type Nodes map[string]*Node

func NewInventoryWatcher(log logrus.FieldLogger, config *InventoryConfig) *InventoryWatcher {
	return &InventoryWatcher{
		log:        log.WithField("module", "ethereum/nodes_watcher"),
		config:     config,
		nodes:      make(Nodes),
		nodesMutex: sync.RWMutex{},
	}
}

func (w *InventoryWatcher) Start(ctx context.Context) error {
	w.log.Info("starting nodes watcher")

	s := gocron.NewScheduler(time.Local)

	if _, err := s.Every("60s").Do(func() {
		if err := w.fetchInventory(ctx); err != nil {
			w.log.WithError(err).Error("failed to fetch inventory")
		}
	}); err != nil {
		return err
	}

	s.StartAsync()

	if err := w.fetchInventory(ctx); err != nil {
		w.log.WithError(err).Error("failed to fetch inventory")
	}

	return nil
}

func (w *InventoryWatcher) fetchInventory(ctx context.Context) error {
	w.log.Debug("fetching inventory")

	rsp, err := http.Get(w.config.URL)
	if err != nil {
		return err
	}

	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", rsp.StatusCode)
	}

	var inventory *AnsibleInventory
	if err := json.NewDecoder(rsp.Body).Decode(&inventory); err != nil {
		return err
	}

	w.log.WithFields(logrus.Fields{
		"ethereum_pairs": len(inventory.EthereumPairs),
	}).Debug("fetched inventory")

	if err := w.handleNewInventory(ctx, inventory); err != nil {
		return err
	}

	return nil
}

func (w *InventoryWatcher) handleNewInventory(ctx context.Context, inventory *AnsibleInventory) error {
	// Determine which nodes are new.
	newNodes := make(map[string]NodeInfo)

	//nolint:gocritic // Not concerned about this amount of data
	for name, pair := range inventory.EthereumPairs {
		if _, ok := w.nodes[name]; !ok {
			newNodes[name] = pair
		}
	}

	// Determine which nodes are no longer in the inventory.
	removedNodes := make(map[string]struct{})

	for name := range w.nodes {
		if _, ok := inventory.EthereumPairs[name]; !ok {
			removedNodes[name] = struct{}{}
		}
	}

	// Add new nodes.
	//nolint:gocritic // Not concerned about this amount of data
	for name, pair := range newNodes {
		// If we have a basic auth username/password then we need to add it to the
		// URL.
		beaconURL := pair.Consensus.BeaconURI

		if w.config.Username != "" && w.config.Password != "" {
			beac, err := url.Parse(pair.Consensus.BeaconURI)
			if err != nil {
				return err
			}

			beac.User = url.UserPassword(w.config.Username, w.config.Password)

			beaconURL = beac.String()
		}

		node, err := w.addNode(ctx, name, beaconURL, pair.Execution.RPCURL, pair)
		if err != nil {
			w.log.WithError(err).Error("failed to add node")
		} else {
			if err := node.Start(ctx); err != nil {
				w.log.WithError(err).Error("failed to start node")
			}
		}
	}

	// Update the info of every node.
	//nolint:gocritic // Not concerned about this amount of data
	for node, info := range inventory.EthereumPairs {
		if n, ok := w.nodes[node]; ok {
			n.UpdateInfo(info)
		}
	}

	// Remove nodes that we are tracking that the inventory is no longer tracking.
	for name := range removedNodes {
		if err := w.removeNode(ctx, name); err != nil {
			w.log.WithError(err).Error("failed to remove node")
		}
	}

	return nil
}

func (w *InventoryWatcher) Nodes() Nodes {
	w.nodesMutex.RLock()
	defer w.nodesMutex.RUnlock()

	return w.nodes
}

//nolint:gocritic // Not concerned about this amount of data
func (w *InventoryWatcher) addNode(ctx context.Context, name, beaconURL, rpcURL string, info NodeInfo) (*Node, error) {
	w.nodesMutex.Lock()
	defer w.nodesMutex.Unlock()

	w.nodes[name] = NewNode(ctx, w.log, name, beaconURL, rpcURL, info)

	w.log.WithFields(logrus.Fields{
		"name": name,
	}).Info("added node")

	return w.nodes[name], nil
}

func (w *InventoryWatcher) removeNode(ctx context.Context, name string) error {
	w.nodesMutex.Lock()
	defer w.nodesMutex.Unlock()

	if _, ok := w.nodes[name]; !ok {
		return fmt.Errorf("unable to remove node: node %s does not exist", name)
	}

	if node, ok := w.nodes[name]; ok {
		if err := node.Stop(ctx); err != nil {
			w.log.WithFields(logrus.Fields{
				"name": name,
				"err":  err,
			}).Error("failed to stop node")
		}
	}

	delete(w.nodes, name)

	w.log.WithFields(logrus.Fields{
		"name": name,
	}).Info("removed node")

	return nil
}
