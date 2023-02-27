package ethereum

import v1 "github.com/attestantio/go-eth2-client/api/v1"

type NodeSummary struct {
	// The name of the node.
	Name string `json:"name"`
	// Info about the node.
	Info NodeInfo `json:"info"`
	// Status is the status of the node.
	Status SummaryStatus `json:"status"`
}

type SummaryStatus struct {
	Consensus ConsensusSummaryStatus `json:"consensus"`
	Execution ExecutionSummaryStatus `json:"execution"`
}

type ConsensusSummaryStatus struct {
	Healthy        bool          `json:"healthy"`
	Version        string        `json:"version"`
	ConfigName     string        `json:"config_name"`
	DepositChainID uint64        `json:"deposit_chain_id"`
	Genesis        *v1.Genesis   `json:"genesis"`
	Finality       *v1.Finality  `json:"finality"`
	Head           *v1.HeadEvent `json:"head"`
}

type ExecutionSummaryStatus struct {
}
