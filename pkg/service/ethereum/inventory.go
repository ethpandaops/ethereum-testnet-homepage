package ethereum

type NodeInfo struct {
	Consensus struct {
		Client    string `json:"client"`
		Image     string `json:"image"`
		ENR       string `json:"enr"`
		PeerID    string `json:"peer_id"`
		BeaconURI string `json:"beacon_uri"`
	} `json:"consensus"`
	Execution struct {
		Client string `json:"client"`
		Image  string `json:"image"`
		ENode  string `json:"enode"`
		RPCURL string `json:"rpc_url"`
	} `json:"execution"`
}

type AnsibleInventory struct {
	EthereumPairs map[string]NodeInfo `json:"ethereum_pairs"`
}
