// Package config provides configuration loading for the LocalAgent.
package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// BrokerConfig holds the broker API connection settings.
type BrokerConfig struct {
	Name      string `yaml:"name"`
	APIKey    string `yaml:"api_key"`
	SecretKey string `yaml:"secret_key"`
	Simulated bool   `yaml:"simulated"`
}

// AgentConfig holds all configuration for the LocalAgent process.
type AgentConfig struct {
	SaaSURL  string       `yaml:"saas_url"`
	Email    string       `yaml:"email"`
	Password string       `yaml:"password"`
	Broker   BrokerConfig `yaml:"broker"`
}

// LoadAgentConfig reads and parses an agent config YAML file.
func LoadAgentConfig(path string) (*AgentConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg AgentConfig
	err = yaml.Unmarshal(data, &cfg)
	return &cfg, err
}
