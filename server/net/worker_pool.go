package net

import (
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/ocean2333/go-crawer/server/config"
	"github.com/ocean2333/go-crawer/server/logger"
	"github.com/ocean2333/go-crawer/server/storage_engine"
)

var (
	defaultWorkerPoolManager *workerPoolManager
	initOnce                 sync.Once
)

type workerPoolManager struct {
	pools map[string]*workerPool
}

func GetWorkerPoolManager() *workerPoolManager {
	initOnce.Do(func() {
		defaultWorkerPoolManager = &workerPoolManager{
			pools: make(map[string]*workerPool),
		}
	})
	return defaultWorkerPoolManager
}

func (wm *workerPoolManager) RegisterPool(name string, size int) {
	wm.pools[name] = NewWorkerPool(size)
}

func (wm *workerPoolManager) GetWorkerPool(name string) *workerPool {
	if wp, ok := wm.pools[name]; ok {
		return wp
	}
	return nil
}

type workerPool struct {
	workers []*worker
}

func NewWorkerPool(size int) *workerPool {
	wp := &workerPool{
		workers: make([]*worker, 0),
	}
	for i := 0; i < size; i++ {
		wp.workers = append(wp.workers, NewWorker())
	}
	return wp
}

func (wp *workerPool) GetPicWithProxyAndSaveSync(url string, rid string, aid string, pid string, wg *sync.WaitGroup) {
	defer wg.Done()
	wp.GetPicWithProxyAndSave(url, rid, aid, pid)
}

func (wp *workerPool) GetThumbnailWithProxyAndSaveSync(url string, rid string, aid string, pid string, wg *sync.WaitGroup) {
	defer wg.Done()
	wp.GetThumbnailWithProxyAndSave(url, rid, aid, pid)
}

func (wp *workerPool) GetPicWithProxyAndSave(url string, rid string, aid string, pid string) {
	var w *worker
	for _, worker := range wp.workers {
		if worker.UseIfFree() {
			w = worker
			break
		}
	}
	if w == nil {
		logger.Log.Errorf("worker pool is full")
		return
	}
	w.getPicWithProxyAndSave(url, rid, aid, pid)
	w.Free()
}

func (wp *workerPool) GetThumbnailWithProxyAndSave(url string, rid string, aid string, pid string) {
	var w *worker
	for _, worker := range wp.workers {
		if worker.UseIfFree() {
			w = worker
			break
		}
	}
	if w == nil {
		logger.Log.Errorf("worker pool is full")
		return
	}
	w.getThumbnailWithProxyAndSave(url, rid, aid, pid)
	w.Free()
}

type worker struct {
	sync.Mutex
	Busy      bool
	transport *http.Transport
}

func NewWorker() *worker {
	return &worker{
		transport: Transport,
	}
}

func (w *worker) UseIfFree() bool {
	w.Lock()
	defer w.Unlock()
	if w.Busy {
		return false
	}
	w.Busy = true
	return true
}

func (w *worker) Free() {
	w.Lock()
	defer w.Unlock()
	w.Busy = false
}

func (w *worker) getThumbnailWithProxyAndSave(url string, rid string, aid string, pid string) {
	client := http.Client{
		Transport: w.transport,
		Timeout:   10 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		logger.Log.Errorf("get %s err: %v", url, err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		logger.Log.Errorf("get %s err: status code: %d", url, resp.StatusCode)
		return
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Log.Errorf("read resp body err: %v", err)
		return
	}
	ops := make([]storage_engine.FileOperation, 0)
	fs := storage_engine.GetFsInstance()
	ops = append(ops, fs.MkDir(config.Get().GlobalCfg.ThumbnailPath))
	ops = append(ops, fs.ChDir(config.Get().GlobalCfg.ThumbnailPath))
	ops = append(ops, fs.MkDir(rid))
	ops = append(ops, fs.ChDir(rid))
	ops = append(ops, fs.MkDir(aid))
	ops = append(ops, fs.ChDir(aid))
	ops = append(ops, fs.CreateAndWrtie(pid+".jpg", data))
	err = fs.Submit(ops...)
	if err != nil {
		logger.Log.Errorf("error when submit file ops: %s", err)
	}
}

func (w *worker) getPicWithProxyAndSave(url string, rid string, aid string, pid string) {
	client := http.Client{
		Transport: w.transport,
		Timeout:   10 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		logger.Log.Errorf("get %s err: %v", url, err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		logger.Log.Errorf("get %s err: status code: %d", url, resp.StatusCode)
		return
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Log.Errorf("read resp body err: %v", err)
		return
	}
	ops := make([]storage_engine.FileOperation, 0)
	fs := storage_engine.GetFsInstance()
	ops = append(ops, fs.MkDir(config.Get().GlobalCfg.HighResPath))
	ops = append(ops, fs.ChDir(config.Get().GlobalCfg.HighResPath))
	ops = append(ops, fs.MkDir(rid))
	ops = append(ops, fs.ChDir(rid))
	ops = append(ops, fs.MkDir(aid))
	ops = append(ops, fs.ChDir(aid))
	ops = append(ops, fs.CreateAndWrtie(pid+".jpg", data))
	err = fs.Submit(ops...)
	if err != nil {
		logger.Log.Errorf("error when submit file ops: %s", err)
	}
}
