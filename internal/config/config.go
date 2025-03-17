package config

import (
	"fmt"
	"os"

	gitconfig "github.com/go-git/go-git/v5/config"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type Config struct {
	GitRoot     string                  `yaml:"git_root"`
	LogFilePath string                  `yaml:"log_file_path"`
	Threads     int                     `yaml:"threads"`
	Remotes     map[string]RemoteConfig `yaml:"remotes"`
}

type RemoteConfig struct {
	URL         string              `yaml:"url"`
	LocalPrefix string              `yaml:"local_prefix"`
	RefSpec     []gitconfig.RefSpec `yaml:"refspec"`
	Repos       []string            `yaml:"repos"`
}

func ReadConfig(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %v", err)
	}

	logrus.Info("config successfuly loaded!")
	return &config, nil
}
