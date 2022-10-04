package storage_engine

import (
	"errors"
	"sort"
	"strings"
	"sync"

	"github.com/ocean2333/go-crawer/src/common"
	"github.com/ocean2333/go-crawer/src/config"
)

var (
	mockInstance *mockEngine
)

type mockEngine struct {
	mutex   sync.RWMutex
	kvs     map[string]*KeyValue
	version int64
}

func (e *mockEngine) Submit(ops []*KvOperation) ([]*KeyValue, error) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	// check version
	for _, op := range ops {
		if op.Op == KvOpcodePrefixDel {
			continue
		}

		oldKv, ok := e.kvs[op.Kv.Key]
		if (!ok && op.Kv.Version != 0) ||
			(ok && oldKv.Version != op.Kv.Version) {
			rc := common.ReturnCode_StorageEngineSubmitFailed
			return nil, common.NewErrorCode(rc, errors.New(rc.String()))
		}
	}

	ret := []*KeyValue{}
	for _, op := range ops {
		switch op.Op {
		case KvOpcodePut:
			e.version++
			e.kvs[op.Kv.Key] = newMockKv(op.Kv.Key, op.Kv.Value, e.version)
			ret = append(ret, newMockKv(op.Kv.Key, op.Kv.Value, e.version))
		case KvOpcodeDel:
			e.version++
			delete(e.kvs, op.Kv.Key)
			ret = append(ret, newMockKv(op.Kv.Key, op.Kv.Value, e.version))
		case KvOpcodePrefixDel:
			deleteKeys := []string{}
			for k, _ := range e.kvs {
				if strings.Index(k, op.Kv.Key) == 0 {
					deleteKeys = append(deleteKeys, k)
				}
			}
			for _, k := range deleteKeys {
				delete(e.kvs, k)
			}
		}
	}

	return ret, nil
}

func (e *mockEngine) Get(key string) (*KeyValue, error) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	if kv, ok := e.kvs[key]; ok {
		return newMockKv(kv.Key, kv.Value, kv.Version), nil
	}

	return nil, nil
}

func (e *mockEngine) GetByPrefix(prefix string) ([]*KeyValue, error) {
	e.mutex.RLock()
	defer e.mutex.RUnlock()

	ret := []*KeyValue{}
	for k, v := range e.kvs {
		if strings.Index(k, prefix) == 0 {
			ret = append(ret, newMockKv(v.Key, v.Value, v.Version))
		}
	}

	return ret, nil
}

func (e *mockEngine) WatchByPrefix(prefix string, handler func(*KeyValue, bool)) {}

func (e *mockEngine) Campaign(opt *CampaignOptions, roleChangedLeaderAddrNotifyCh chan string) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	campInfos := []*KeyValue{}
	for k, v := range e.kvs {
		if strings.Index(k, opt.CampaignKeyPrefix) == 0 {
			campInfos = append(campInfos, v)
		}
	}

	sort.Slice(campInfos, func(i, j int) bool {
		return campInfos[i].Version < campInfos[j].Version
	})

	exist := false
	for _, info := range campInfos {
		if string(info.Value) == opt.CampaignId {
			exist = true
			break
		}
	}

	if !exist {
		key := opt.CampaignKeyPrefix + "/" + opt.CampaignId
		e.version++
		e.kvs[key] = newMockKv(key, []byte(opt.CampaignId), e.version)
		campInfos = append(campInfos, e.kvs[key])
	}

	roleChangedLeaderAddrNotifyCh <- string(campInfos[0].Value)
}

func newMockEngine(cfg *config.StorageEngineConfig) (StorageEngine, error) {
	mockInstance = &mockEngine{
		kvs:     make(map[string]*KeyValue),
		version: 0,
	}
	return mockInstance, nil
}

func newMockKv(key string, value []byte, version int64) *KeyValue {
	ret := &KeyValue{Key: key, Version: version}
	ret.Value = append(ret.Value, value...)
	return ret
}
