package homepage

import (
	"fmt"

	"github.com/ethpandaops/ethereum-testnet-homepage/pkg/service/ethereum"
	"github.com/ethpandaops/ethereum-testnet-homepage/pkg/service/homepage"
)

type Config struct {
	Global   GlobalConfig    `yaml:"global"`
	Ethereum ethereum.Config `yaml:"ethereum"`
	Homepage homepage.Config `yaml:"homepage"`
}

type GlobalConfig struct {
	ListenAddr   string `yaml:"listenAddr" default:":5555"`
	LoggingLevel string `yaml:"logging" default:"warn"`
	MetricsAddr  string `yaml:"metricsAddr" default:":9090"`
}

func (c *GlobalConfig) Validate() error {
	return nil
}

func (c *Config) Validate() error {
	if err := c.Global.Validate(); err != nil {
		return fmt.Errorf("global config: %w", err)
	}

	if err := c.Ethereum.Validate(); err != nil {
		return fmt.Errorf("ethereum config: %w", err)
	}

	if err := c.Homepage.Validate(); err != nil {
		return fmt.Errorf("homepage config: %w", err)
	}

	return nil
}
