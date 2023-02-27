package ethereum

import (
	"context"

	"github.com/sirupsen/logrus"
)

// Service is the Ethereum Service. HTTP-level concerns should NOT be contained in this package,
// they should be handled and reasoned with at a higher level.
type Service struct {
	log logrus.FieldLogger

	inventory *InventoryWatcher

	metrics *Metrics
}

// NewService returns a new Service instance.
func NewService(log logrus.FieldLogger, namespace string, config *Config) *Service {
	return &Service{
		log: log.WithField("module", "service/ethereum"),

		inventory: NewInventoryWatcher(log, &config.Inventory),

		metrics: NewMetrics(namespace),
	}
}

func (s *Service) Start(ctx context.Context) error {
	s.log.Info("Starting Ethereum service")

	if err := s.inventory.Start(ctx); err != nil {
		return err
	}

	return nil
}

// Nodes returns basic information about the configured nodes.
func (s *Service) Nodes(ctx context.Context, req *NodesRequest) (*NodesResponse, error) {
	var err error

	const call = "nodes"

	s.metrics.ObserveCall(call, "")

	defer func() {
		if err != nil {
			s.metrics.ObserveErrorCall(call, "")
		}
	}()

	rsp := &NodesResponse{}

	s.inventory.nodesMutex.RLock()

	defer s.inventory.nodesMutex.RUnlock()

	for name, node := range s.inventory.Nodes() {
		summary := NodeSummary{
			Name: name,
			Info: node.Info(),
			Status: SummaryStatus{
				Consensus: ConsensusSummaryStatus{
					Healthy: node.beacon.Healthy(),
				},
				Execution: ExecutionSummaryStatus{},
			},
		}

		if node.ConsensusFinalizedCheckpoint != nil {
			summary.Status.Consensus.Finality = node.ConsensusFinalizedCheckpoint
		}

		if node.ConsensusHead != nil {
			summary.Status.Consensus.Head = node.ConsensusHead
		}

		genesis, err := node.beacon.Genesis()
		if err == nil {
			summary.Status.Consensus.Genesis = genesis
		}

		beaconConfig, err := node.beacon.Spec()
		if err == nil {
			summary.Status.Consensus.ConfigName = beaconConfig.ConfigName
			summary.Status.Consensus.DepositChainID = beaconConfig.DepositChainID
		}

		beaconVersion, err := node.beacon.NodeVersion()
		if err == nil {
			summary.Status.Consensus.Version = beaconVersion
		}

		rsp.Nodes = append(rsp.Nodes, summary)
	}

	return &NodesResponse{
		Nodes: rsp.Nodes,
	}, nil
}
