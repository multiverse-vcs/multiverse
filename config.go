package ipfs

import (
	"github.com/go-git/go-git/v5/config"
)

type ConfigStorage struct {
	config *config.Config
}

func (c *ConfigStorage) SetConfig(cfg *config.Config) error {
	if err := cfg.Validate(); err != nil {
		return err
	}

	c.config = cfg
	return nil
}

func (c *ConfigStorage) Config() (*config.Config, error) {
	if c.config == nil {
		c.config = config.NewConfig()
	}

	return c.config, nil
}
