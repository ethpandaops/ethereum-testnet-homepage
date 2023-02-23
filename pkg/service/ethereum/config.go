package ethereum

import "errors"

type Config struct {
	NetworkName string `yaml:"network_name"`

	Inventory InventoryConfig `yaml:"inventory"`
}

type InventoryConfig struct {
	Enabled bool `yaml:"enabled"`

	URL string `yaml:"url"`

	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

func (c *Config) Validate() error {
	if c.NetworkName == "" {
		return errors.New("no network name specified")
	}

	if c.Inventory.Enabled {
		if c.Inventory.URL == "" {
			return errors.New("no inventory URL specified")
		}
	}

	return nil
}
