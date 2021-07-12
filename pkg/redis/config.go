package redis

import (
	"encoding/json"
	"fmt"
)

const (
	configModeSingle  ConfigMode = "single"
	configModeCluster ConfigMode = "cluster"
)

type ConfigMode string

func (c *ConfigMode) UnmarshalJSON(b []byte) error {
	if b == nil || len(b) == 0 {
		return nil
	}

	var dst string
	if err := json.Unmarshal(b, &dst); err != nil {
		return err
	}

	switch ConfigMode(dst) {
	case "", configModeCluster, configModeSingle:
		*c = ConfigMode(dst)
		return nil
	}

	return fmt.Errorf("unsupported redis mode: %q", dst)
}

type Config struct {
	Mode ConfigMode `yaml:"mode" env:"REDIS_MODE"`
	Addr string     `yaml:"addr" env:"REDIS_ADDR"`
}
