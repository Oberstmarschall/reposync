package main

import (
	"flag"
	"reposync/internal/config"
	"reposync/internal/logging"
	"reposync/internal/repo"
)

func main() {
	confFilePath := parseFlags()
	cfg, err := config.ReadConfig(confFilePath)
	if err != nil {
		panic(err)
	}

	err = logging.InitLogging(cfg.LogFilePath)
	if err != nil {
		panic(err)
	}

	repo.SyncReposInParallel(cfg.Threads, cfg.GitRoot, &cfg.Remotes)
}

func parseFlags() string {
	configPath := flag.String("config", `/etc/reposync/config.yaml`, "Path to the configuration yml file")

	flag.Parse()

	return *configPath
}
