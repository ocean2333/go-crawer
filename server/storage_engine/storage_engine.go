package storage_engine

import (
	"errors"
	"sync"

	"github.com/ocean2333/go-crawer/server/config"
	"github.com/ocean2333/go-crawer/server/logger"
)

const (
	KvOpcodePut       uint8 = 1
	KvOpcodeDel       uint8 = 2
	KvOpcodePrefixDel uint8 = 3
)

var (
	seInstance StorageEngine
	startOnce  sync.Once
	creators   map[string]storageEngineCreator
	mutex      sync.Mutex
)

type KeyValue struct {
	Key     string
	Value   []byte
	Version int64
}

type KvOperation struct {
	Op uint8
	Kv KeyValue
}

type CampaignOptions struct {
	Cfg               *config.LeaderElectionConfig
	CampaignKeyPrefix string
	CampaignId        string
}

type StorageEngine interface {
	Submit(ops []*KvOperation) ([]*KeyValue, error)
	Get(key string) (*KeyValue, error)
	GetByPrefix(prefix string) ([]*KeyValue, error)
	WatchByPrefix(prefix string, handler func(*KeyValue, bool))
	Campaign(opt *CampaignOptions, roleChangedLeaderAddrNotifyCh chan string)
}

type storageEngineCreator func(cfg *config.StorageEngineConfig) (StorageEngine, error)

func Start(cfg *config.StorageEngineConfig) error {
	var err error

	mutex.Lock()
	defer mutex.Unlock()

	if seInstance != nil {
		logger.Log.Infof("storage engine already Started")
		return nil
	}

	creator, ok := creators[cfg.Name]
	if !ok {
		logger.Log.Errorf("Start storage engine with invalid config(%#v)", cfg)
		return errors.New("invalid storage engine name")
	}

	seInstance, err = creator(cfg)
	if err != nil {
		logger.Log.Errorf("Start storage engine err: %v", err)
	}

	return err
}

func InitEtcdEngine() {
	startOnce.Do(func() {
		Start(&config.Get().EtcdCfg.StorageEngineCfg)
		go func() {
			GetInstance().Campaign(&CampaignOptions{
				Cfg:               &config.Get().EtcdCfg.LeaderElectionCfg,
				CampaignKeyPrefix: "campaign",
				CampaignId:        "1",
			}, make(chan string))
		}()
	})
}

func GetInstance() StorageEngine {
	InitEtcdEngine()
	return seInstance
}

func ClearForUnitTest() {
	prefixDelOps := []*KvOperation{{Op: KvOpcodePrefixDel, Kv: KeyValue{Key: "/"}}}
	if _, err := GetInstance().Submit(prefixDelOps); err != nil {
		panic(err)
	}
}

func init() {
	mutex.Lock()
	defer mutex.Unlock()

	if creators == nil {
		creators = make(map[string]storageEngineCreator)
	}

	creators["etcd"] = newEtcdEngine
	creators["mock"] = newMockEngine
}
