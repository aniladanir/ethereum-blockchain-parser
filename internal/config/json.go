package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type jsonConfiguration struct {
	cfg *Config
}

func NewJsonConfig(filePath string) (Configurator, error) {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading configuration file: %w", err)
	}

	cfg := new(Config)
	if err := json.Unmarshal(fileContent, cfg); err != nil {
		return nil, err
	}

	return &jsonConfiguration{cfg: cfg}, nil
}

func (jc *jsonConfiguration) GetAppVersion() string {
	return jc.cfg.Version
}

func (jc *jsonConfiguration) GetLogFile() string {
	return jc.cfg.Log.File
}

func (jc *jsonConfiguration) GetLogLevel() string {
	return jc.cfg.Log.Level
}

func (jc *jsonConfiguration) GetHttpServerIP() string {
	return jc.cfg.Http.Server.IpAddress
}

func (jc *jsonConfiguration) GetHttpServerPort() int {
	return jc.cfg.Http.Server.Port
}

func (jc *jsonConfiguration) GetChainProcessInterval() int {
	return jc.cfg.ChainProcessInterval
}
