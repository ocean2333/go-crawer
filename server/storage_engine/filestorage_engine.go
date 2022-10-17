package storage_engine

type FileOperation interface {
	Type() string
	Params() []string
	GetData() []byte
	SetData([]byte)
}

type FileStorageEngine interface {
	Submit(...FileOperation) error
	ChDir(path string) FileOperation
	MkDir(path string) FileOperation
	Read(path string) FileOperation
	Overwrite(path string, data []byte) FileOperation
	CreateAndWrtie(path string, data []byte) FileOperation
}
