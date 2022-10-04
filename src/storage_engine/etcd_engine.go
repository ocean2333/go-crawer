package storage_engine

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/ocean2333/go-crawer/src/common"
	"github.com/ocean2333/go-crawer/src/config"
	"github.com/ocean2333/go-crawer/src/logger"
)

var (
	etcdInstance *etcdEngine
)

type etcdEngine struct {
	client      *clientv3.Client
	mutex       sync.RWMutex
	leaderFlag  bool
	rpcTimeout  uint32
	dialTimeout uint32
}

func newEtcdEngine(cfg *config.StorageEngineConfig) (StorageEngine, error) {
	var err error

	etcdInstance = &etcdEngine{
		leaderFlag:  false,
		rpcTimeout:  cfg.RpcTimeout,
		dialTimeout: cfg.DialTimeout,
	}

	etcdAddresses := strings.Split(cfg.Addr, ",")
	etcdCfg := clientv3.Config{
		Endpoints:   etcdAddresses,
		DialTimeout: time.Duration(cfg.DialTimeout) * time.Second,
	}

	etcdInstance.client, err = clientv3.New(etcdCfg)
	if err != nil {
		logger.Log.Errorf("newEtcdEngine addr(%#v), new clientv3 err: %v", etcdAddresses, err)
		return nil, common.NewErrorCode(common.ReturnCode_StorageEngineInternalError, err)
	}

	if err = etcdInstance.check(); err != nil {
		logger.Log.Errorf("netEtcdEngine addr(%#v) check status err: %v", etcdAddresses, err)
		return nil, common.NewErrorCode(common.ReturnCode_StorageEngineInternalError, err)
	}

	logger.Log.Infof("newEtcdEngine success, cfg(%#v), instance(%#v)", cfg, etcdInstance)

	return etcdInstance, nil
}

func (e *etcdEngine) Submit(ops []*KvOperation) ([]*KeyValue, error) {
	if !e.isLeader() {
		return nil, common.NewErrorCode(common.ReturnCode_StorageEngineInternalError, errors.New("submit by slave err"))
	}

	kv := clientv3.NewKV(e.client)
	txn := kv.Txn(context.TODO())
	txnCs := make([]clientv3.Cmp, 0)
	txnOps := make([]clientv3.Op, 0)

	for _, op := range ops {
		if len(op.Kv.Key) == 0 && op.Op != KvOpcodePrefixDel {
			return nil, common.NewErrorCode(common.ReturnCode_StorageEngineInternalError, errors.New("invalid operation key"))
		}
		switch op.Op {
		case KvOpcodePut:
			if op.Kv.Version == 0 {
				txnCs = append(txnCs, clientv3.Compare(clientv3.CreateRevision(op.Kv.Key), "=", 0))
			} else {
				txnCs = append(txnCs, clientv3.Compare(clientv3.ModRevision(op.Kv.Key), "=", op.Kv.Version))
			}
			txnOps = append(txnOps, clientv3.OpPut(op.Kv.Key, string(op.Kv.Value)))
		case KvOpcodeDel:
			txnCs = append(txnCs, clientv3.Compare(clientv3.ModRevision(op.Kv.Key), "=", op.Kv.Version))
			txnOps = append(txnOps, clientv3.OpDelete(op.Kv.Key))
		case KvOpcodePrefixDel:
			txnOps = append(txnOps, clientv3.OpDelete(op.Kv.Key, clientv3.WithPrefix()))
		}
		logger.Log.Debugf("op(0x%x), key(%s), val(%s), version(0x%x)", op.Op, op.Kv.Key, string(op.Kv.Value), op.Kv.Version)
	}

	txnResp, err := txn.If(txnCs...).Then(txnOps...).Commit()
	if err != nil {
		logger.Log.Errorf("Submit err: %v", err)
		return nil, common.NewErrorCode(common.ReturnCode_StorageEngineInternalError, err)
	}

	if !txnResp.Succeeded {
		logger.Log.Errorf("Submit txn commit failed")
		return nil, common.NewErrorCode(common.ReturnCode_StorageEngineSubmitFailed, errors.New("txn commit err"))
	}

	ret := make([]*KeyValue, len(ops))

	for i, op := range ops {
		ret[i] = &KeyValue{Key: op.Kv.Key, Value: op.Kv.Value}

		switch op.Op {
		case KvOpcodePut:
			ret[i].Version = txnResp.Responses[i].GetResponsePut().Header.Revision
		case KvOpcodeDel:
			ret[i].Version = txnResp.Responses[i].GetResponseDeleteRange().Header.Revision
		case KvOpcodePrefixDel:
			ret[i].Version = txnResp.Responses[i].GetResponseDeleteRange().Header.Revision
		}
	}

	// update memcache after submit
	memcache := GetMemcacheInstance()
	for _, op := range ops {
		switch op.Op {
		case KvOpcodePut:
			memcache.Set(string(op.Kv.Key), op.Kv.Value)
		case KvOpcodeDel:
			memcache.Delete(string(op.Kv.Key))
		case KvOpcodePrefixDel:
			memcache.Reset()
		}
	}

	return ret, nil
}

func (e *etcdEngine) Get(key string) (*KeyValue, error) {
	// try to get from memcache first
	memcache := GetMemcacheInstance()
	if v, err := memcache.Get(key); err == nil {
		return &KeyValue{Key: key, Value: v, Version: 1}, nil
	}

	// else, get from etcd
	kv := clientv3.NewKV(e.client)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(e.rpcTimeout)*time.Second)
	defer cancel()
	getResp, err := kv.Get(ctx, key)
	if err != nil {
		logger.Log.Errorf("Get key(%s) err: %v", key, err)
		return nil, common.NewErrorCode(common.ReturnCode_StorageEngineInternalError, err)
	}
	if len(getResp.Kvs) == 0 {
		logger.Log.Debugf("Get key(%s) not exist", key)
		return nil, nil
	}

	ret := &KeyValue{
		Key:     key,
		Value:   getResp.Kvs[0].Value,
		Version: getResp.Kvs[0].ModRevision,
	}

	// put to memcache
	memcache.Set(key, ret.Value)
	return ret, nil
}

func (e *etcdEngine) GetByPrefix(prefix string) ([]*KeyValue, error) {
	// prefix not support get from memcache
	kv := clientv3.NewKV(e.client)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(e.rpcTimeout)*time.Second)
	defer cancel()
	getResp, err := kv.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		logger.Log.Errorf("GetByPrefix prefix(%s) err: %v", prefix, err)
		return nil, common.NewErrorCode(common.ReturnCode_StorageEngineInternalError, err)
	}

	if len(getResp.Kvs) == 0 {
		logger.Log.Debugf("GetByPrefix prefix(%s) kvs not exist", prefix)
		return nil, nil
	}

	ret := make([]*KeyValue, len(getResp.Kvs))
	for i, item := range getResp.Kvs {
		ret[i] = &KeyValue{
			Key:     string(item.Key),
			Value:   item.Value,
			Version: item.ModRevision,
		}
	}

	// put to memcache
	memcache := GetMemcacheInstance()
	for _, item := range ret {
		memcache.Set(item.Key, item.Value)
	}
	return ret, nil
}

func (e *etcdEngine) WatchByPrefix(prefix string, handler func(*KeyValue, bool)) {
	go func() {
		rch := clientv3.NewWatcher(e.client).Watch(context.TODO(), prefix, clientv3.WithPrefix(), clientv3.WithPrefix())
		for wresp := range rch {
			for _, ev := range wresp.Events {
				kv := &KeyValue{
					Key:     string(ev.Kv.Key),
					Value:   ev.Kv.Value,
					Version: ev.Kv.ModRevision,
				}
				handler(kv, ev.Type == clientv3.EventTypeDelete)
			}
		}
	}()
}

func (e *etcdEngine) Campaign(opt *CampaignOptions, roleChangedLeaderAddrNotifyCh chan string) {
	for {
		if err := e.check(); err != nil {
			logger.Log.Errorf("Campaign client check err: %v", err)
			time.Sleep(time.Duration(opt.Cfg.ElectionPeriod) * time.Second)
			continue
		}

		session, err := concurrency.NewSession(e.client, concurrency.WithTTL(int(opt.Cfg.LeaderLeasePeriod)))
		if err != nil {
			logger.Log.Errorf("Campaign NewSession err: %v", err)
			time.Sleep(time.Duration(opt.Cfg.ElectionPeriod) * time.Second)
			continue
		}
		election := concurrency.NewElection(session, opt.CampaignKeyPrefix)
		observeRoleChangeCh := election.Observe(context.Background())
		stopObserveCh := make(chan struct{})
		reElectCh := make(chan struct{})

		go func() { // watch for role change
			for {
				select {
				case <-stopObserveCh:
					logger.Log.Errorf("Campaign self(%s) watch role change, stop observe", opt.CampaignId)
					e.setLeaderFlag(false)
					e.onRoleChange(roleChangedLeaderAddrNotifyCh, "")
					return
				case resp, ok := <-observeRoleChangeCh:
					if !ok { // observed channel closed
						logger.Log.Infof("Campaign self(%s) watch role change, the channel is closed", opt.CampaignId)
						e.setLeaderFlag(false)
						e.onRoleChange(roleChangedLeaderAddrNotifyCh, "")
						close(reElectCh) // trigger re election
						return
					}

					if len(resp.Kvs) == 0 {
						logger.Log.Infof("Campaign self(%s) watch role change, the value is empty", opt.CampaignId)
						e.setLeaderFlag(false)
						e.onRoleChange(roleChangedLeaderAddrNotifyCh, "")
						close(reElectCh) // trigger re election
						return
					}

					leader := string(resp.Kvs[0].Value)
					logger.Log.Infof("Campaign self(%s) watch role change, leader(%s)", opt.CampaignId, leader)
					e.setLeaderFlag(leader == opt.CampaignId)
					e.onRoleChange(roleChangedLeaderAddrNotifyCh, leader)
				}
			}
		}()

		logger.Log.Infof("Campaign self(%s) start Campaign", opt.CampaignId)
		if err := election.Campaign(context.Background(), opt.CampaignId); err != nil {
			close(stopObserveCh)
			logger.Log.Errorf("Campaign self(%s) err: %v", opt.CampaignId, err)
			time.Sleep(time.Duration(opt.Cfg.ElectionPeriod) * time.Second)
			continue
		}
		logger.Log.Infof("Campaign self(%s) end Campaign, I should be the leader", opt.CampaignId)

		select {
		case <-session.Done(): // lose leader, stop task
			logger.Log.Infof("Campaign self(%s) lose the leader", opt.CampaignId)
			close(stopObserveCh)
		case <-reElectCh: // observed channel close, re-election
			logger.Log.Infof("Campaign self(%s) re-election", opt.CampaignId)
			close(stopObserveCh)
			ctxTmp, cancel := context.WithTimeout(context.Background(), time.Second*1)
			election.Resign(ctxTmp)
			session.Close()
			cancel()
		}
	}
}

func (e *etcdEngine) isLeader() bool {
	e.mutex.RLock()
	defer e.mutex.RUnlock()
	return e.leaderFlag
}

func (e *etcdEngine) setLeaderFlag(leaderFlag bool) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	e.leaderFlag = leaderFlag
}

func (e *etcdEngine) check() error {
	kv := clientv3.NewKV(e.client)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(e.dialTimeout)*time.Second)
	defer cancel()
	if _, err := kv.Get(ctx, "check_status"); err != nil {
		return common.NewErrorCode(common.ReturnCode_StorageEngineInternalError, err)
	}
	return nil
}

func (e *etcdEngine) onRoleChange(roleChangeLeaderValNotifyCh chan string, leaderVal string) {
	go func() {
		roleChangeLeaderValNotifyCh <- leaderVal
	}()
}
