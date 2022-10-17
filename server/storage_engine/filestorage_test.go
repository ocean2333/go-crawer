package storage_engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMkdir(t *testing.T) {
	fs := GetFsInstance()
	fo := fs.MkDir("testDir")
	err := fs.Submit(fo)
	assert.Nil(t, err)
}

func TestCreateAndWrtie(t *testing.T) {
	fs := GetFsInstance()
	fo := fs.CreateAndWrtie("testfile", []byte("123456"))
	err := fs.Submit(fo)
	assert.Nil(t, err)
}

func TestReadFile(t *testing.T) {
	fs := GetFsInstance()
	fo := fs.Read("testfile")
	err := fs.Submit(fo)
	assert.Nil(t, err)
	data := fo.GetData()
	assert.Equal(t, string(data), "123456")
}
