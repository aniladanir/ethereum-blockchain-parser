package config

type Configurator interface {
	// GetAppVersion returns application version
	GetAppVersion() string
	// GetLogFile returns path to the log file
	GetLogFile() string
	// GetLogLevel returns log level
	GetLogLevel() string
	// GetHttpServerIP return server ip address
	GetHttpServerIP() string
	// GetHttpServerPort returns server port
	GetHttpServerPort() int
	// GetChainProcessInterval returns internal for chain process in milliseconds
	GetChainProcessInterval() int
}

type Config struct {
	Version              string `json:"version"`
	ChainProcessInterval int    `json:"chainProcessInterval"`
	Log                  struct {
		File  string `json:"file"`
		Level string `json:"level"`
	} `json:"log"`
	Http struct {
		Server struct {
			IpAddress string `json:"ipAddress"`
			Port      int    `json:"port"`
		} `json:"server"`
	} `json:"http"`
}
