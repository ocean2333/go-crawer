package storage_engine

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"sync"
)

var (
	fsInstance *fsEngine
	newOnce    = &sync.Once{}
)

type fsEngine struct {
	BasePath   string
	globalLock sync.Mutex // use global lock to prevent deadlock
	locks      map[string]*sync.RWMutex
}

type fsOperation struct {
	opType string
	params []string
	data   []byte
}

func (op *fsOperation) Type() string {
	return op.opType
}

func (op *fsOperation) Params() []string {
	return op.params
}

func (op *fsOperation) GetData() []byte {
	return op.data
}

func (op *fsOperation) SetData(data []byte) {
	op.data = data
}

func (fs *fsEngine) ChDir(p string) FileOperation {
	return &fsOperation{
		opType: "chdir",
		params: []string{p},
	}
}

func (fs *fsEngine) MkDir(p string) FileOperation {
	return &fsOperation{
		opType: "mkdir",
		params: []string{p},
	}
}

func (fs *fsEngine) Read(p string) FileOperation {
	return &fsOperation{
		opType: "read",
		params: []string{p},
	}
}

func (fs *fsEngine) Overwrite(p string, data []byte) FileOperation {
	return &fsOperation{
		opType: "overwrite",
		params: []string{p},
		data:   data,
	}
}

func (fs *fsEngine) CreateAndWrtie(p string, data []byte) FileOperation {
	return &fsOperation{
		opType: "create_and_write",
		params: []string{p},
		data:   data,
	}
}

func (fsEngine *fsEngine) Submit(ops ...FileOperation) error {
	nowPath := fsEngine.BasePath
	fsEngine.globalLock.Lock()
	// peek all ops to lock relative file or path
	for _, op := range ops {
		if op.Type() == "read" {
			lock := fsEngine.locks[path.Join(nowPath, op.Params()[0])]
			if lock == nil {
				fsEngine.locks[path.Join(nowPath, op.Params()[0])] = &sync.RWMutex{}
			}
			fsEngine.locks[path.Join(nowPath, op.Params()[0])].RLock()
		} else if op.Type() == "overwrite" || op.Type() == "create_and_write" {
			lock := fsEngine.locks[path.Join(nowPath, op.Params()[0])]
			if lock == nil {
				fsEngine.locks[path.Join(nowPath, op.Params()[0])] = &sync.RWMutex{}
			}
			fsEngine.locks[path.Join(nowPath, op.Params()[0])].Lock()
		} else if op.Type() == "chdir" {
			nowPath = path.Join(nowPath, op.Params()[0])
		} else if op.Type() == "mkdir" {
		} else {
			return fmt.Errorf("unknown operation: %s", op.Type())
		}
	}
	fsEngine.globalLock.Unlock()

	nowPath = fsEngine.BasePath
	// do all ops
	for _, op := range ops {
		if op.Type() == "read" {
			data, err := ioutil.ReadFile(path.Join(nowPath, op.Params()[0]))
			if err != nil {
				return err
			}
			op.SetData(data)
			lock := fsEngine.locks[path.Join(nowPath, op.Params()[0])]
			lock.RUnlock()
		} else if op.Type() == "overwrite" {
			// TODO
		} else if op.Type() == "create_and_write" {
			_, err := os.Create(path.Join(nowPath, op.Params()[0]))
			if err != nil && !os.IsExist(err) {
				return err
			}
			data := op.GetData()
			ioutil.WriteFile(path.Join(nowPath, op.Params()[0]), data, fs.ModeAppend)
			lock := fsEngine.locks[path.Join(nowPath, op.Params()[0])]
			lock.Unlock()
		} else if op.Type() == "chdir" {
			nowPath = path.Join(nowPath, op.Params()[0])
		} else if op.Type() == "mkdir" {
			exist, err := dirExist(path.Join(nowPath, op.Params()[0]))
			if err != nil {
				return err
			}
			if !exist {
				err := os.Mkdir(path.Join(nowPath, op.Params()[0]), os.ModePerm)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func InitFileStorageEngine(basePath string) {
	newOnce.Do(func() {
		fsInstance = &fsEngine{
			BasePath: ".",
			locks:    make(map[string]*sync.RWMutex),
		}
	})
	fo := fsInstance.MkDir("./base")
	err := fsInstance.Submit(fo)
	if err != nil {
		panic(err)
	}
	fsInstance.BasePath = "./base"
}

func GetFsInstance() FileStorageEngine {
	if fsInstance == nil {
		InitFileStorageEngine("./base")
	}
	return fsInstance
}

func dirExist(path string) (bool, error) {
	stat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if stat.IsDir() {
		return true, nil
	}
	return false, nil
}
