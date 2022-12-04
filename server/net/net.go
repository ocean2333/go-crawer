package net

import (
	"io/ioutil"
	"net/http"
	"time"

	"github.com/ocean2333/go-crawer/server/config"
	"github.com/ocean2333/go-crawer/server/logger"
	"github.com/ocean2333/go-crawer/server/storage_engine"
)

func GetWithoutProxy(url string) (resp *http.Response, err error) {
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	return client.Get(url)
}

func GetWithDefaultProxy(url string) (resp *http.Response, err error) {
	client := http.Client{
		Transport: Transport,
		Timeout:   10 * time.Second,
	}
	return client.Get(url)
}

func GetThumbnailWithProxyAndSave(url string, rid string, aid string, pid string) error {
	client := http.Client{
		Transport: Transport,
		Timeout:   10 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		logger.Log.Errorf("get %s err: %v", url, err)
		return err
	}
	if resp.StatusCode != http.StatusOK {
		logger.Log.Errorf("get %s err: status code: %d", url, resp.StatusCode)
		return err
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Log.Errorf("read resp body err: %v", err)
		return err
	}
	ops := make([]storage_engine.FileOperation, 0)
	fs := storage_engine.GetFsInstance()
	ops = append(ops, fs.MkDir(config.Get().GlobalCfg.ThumbnailPath))
	ops = append(ops, fs.ChDir(config.Get().GlobalCfg.ThumbnailPath))
	ops = append(ops, fs.MkDir(rid))
	ops = append(ops, fs.ChDir(rid))
	ops = append(ops, fs.MkDir(aid))
	ops = append(ops, fs.ChDir(aid))
	ops = append(ops, fs.CreateAndWrtie(pid, data))
	err = fs.Submit(ops...)
	if err != nil {
		logger.Log.Errorf("error when submit file ops: %s", err)
	}
	return nil
}

func GetPicWithProxyAndSave(url string, rid string, aid string, pid string) error {
	client := http.Client{
		Transport: Transport,
		Timeout:   10 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		logger.Log.Errorf("get %s err: %v", url, err)
		return err
	}
	if resp.StatusCode != http.StatusOK {
		logger.Log.Errorf("get %s err: status code: %d", url, resp.StatusCode)
		return err
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Log.Errorf("read resp body err: %v", err)
		return err
	}
	ops := make([]storage_engine.FileOperation, 0)
	fs := storage_engine.GetFsInstance()
	ops = append(ops, fs.MkDir(config.Get().GlobalCfg.HighResPath))
	ops = append(ops, fs.ChDir(config.Get().GlobalCfg.HighResPath))
	ops = append(ops, fs.MkDir(rid))
	ops = append(ops, fs.ChDir(rid))
	ops = append(ops, fs.MkDir(aid))
	ops = append(ops, fs.ChDir(aid))
	ops = append(ops, fs.CreateAndWrtie(pid, data))
	err = fs.Submit(ops...)
	if err != nil {
		logger.Log.Errorf("error when submit file ops: %s", err)
	}
	return nil
}
