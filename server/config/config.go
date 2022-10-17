package config

import (
	"io/ioutil"

	"github.com/ocean2333/go-crawer/server/logger"
	"gopkg.in/yaml.v2"
)

type Config struct {
	GlobalCfg GlobalConfig `yaml:"global_config"`
	EtcdCfg   EtcdConfig   `yaml:"etcd_config"`
}

type GlobalConfig struct {
	BasePath      string `yaml:"save_path"`
	ThumbnailPath string `yaml:"thumbnail_path"`
	HighResPath   string `yaml:"highres_path"`
}

type EtcdConfig struct {
	LeaderElectionCfg LeaderElectionConfig `yaml:"leader_election_config"`
	StorageEngineCfg  StorageEngineConfig  `yaml:"storage_engine_config"`
}

type LeaderElectionConfig struct {
	LeaderLeasePeriod int32 `yaml:"leader_lease_period"`
	ElectionPeriod    int32 `yaml:"election_period"`
}

type StorageEngineConfig struct {
	Name        string `yaml:"name"`
	Addr        string `yaml:"addr"`
	DialTimeout uint32 `yaml:"dial_timeout"`
	RpcTimeout  uint32 `yaml:"rpc_timeout"`
}

var cfg *Config

func init() {
	var err error
	cfg, err = LoadConfig("config.yaml")
	if err != nil {
		logger.Log.Errorf("load config error, use default config")
		cfg = defaultConfig()
		return
	}
}

func defaultConfig() *Config {
	return &Config{
		GlobalCfg: GlobalConfig{
			BasePath:      "./base",
			ThumbnailPath: "./thumbnail",
			HighResPath:   "hisg_res",
		},
		EtcdCfg: EtcdConfig{
			LeaderElectionCfg: LeaderElectionConfig{
				LeaderLeasePeriod: 10,
				ElectionPeriod:    10,
			},
			StorageEngineCfg: StorageEngineConfig{
				Name:        "etcd",
				Addr:        "127.0.0.1:2379",
				DialTimeout: 5,
				RpcTimeout:  30,
			},
		},
	}
}

func LoadConfig(file string) (*Config, error) {
	if file == "" {
		logger.Log.Infof("Without config file, default cfg(%#v)", cfg)
		return cfg, nil
	}

	content, err := ioutil.ReadFile(file)
	if err != nil {
		logger.Log.Errorf("LoadConfig file(%s) err: %v", file, err)
		return nil, err
	}

	//logs.Debug("content: %s", string(content))

	if err = yaml.Unmarshal(content, &cfg); err != nil {
		logger.Log.Errorf("LoadConfig unmarshal file(%s) err: %v", file, err)
		return nil, err
	}

	logger.Log.Infof("LoadConfig cfg(%#v)", cfg)

	return cfg, nil
}

func Get() *Config {
	return cfg
}
