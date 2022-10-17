package model

import (
	"encoding/json"
	"fmt"

	"github.com/ocean2333/go-crawer/src/storage_engine"
)

type StoreItem struct {
	Item     string            `json:"item"`
	Id       string            `json:"id"`
	Property map[string]string `json:"property"`
}

func (s *StoreItem) String() string {
	return fmt.Sprintf("item: %s, id: %s, property: %v", s.Item, s.Id, s.Property)
}

func (s *StoreItem) Set(name string, value string) {
	if s.Property == nil {
		s.Property = make(map[string]string)
	}
	s.Property[name] = value
}

func (s *StoreItem) Get(name string) string {
	return s.Property[name]
}

type StoreAlbumMetadata struct {
	Version     uint64 `json:"version"`
	TimeStamp   uint64 `json:"timestamp"`
	Rid         string `json:"rid"`
	Aid         string `json:"aid"`
	Title       string `json:"title"`
	Datetime    string `json:"datetime"`
	Cover       string `json:"cover"`
	Author      string `json:"author"`
	Uploader    string `json:"uploader"`
	Rating      string `json:"rating"`
	Tag         string `json:"tag"`
	Description string `json:"description"`
	Url         string `json:"url"`
}

func (s *StoreAlbumMetadata) Key() string {
	return fmt.Sprintf("%s-%s", s.Rid, s.Aid)
}

func (s *StoreAlbumMetadata) Encode() ([]byte, error) {
	return json.Marshal(s)
}

func (s *StoreAlbumMetadata) Decode(data []byte) error {
	return json.Unmarshal(data, s)
}

func (s *StoreAlbumMetadata) ToKvValue() (storage_engine.KeyValue, error) {
	v, err := s.Encode()
	if err != nil {
		return storage_engine.KeyValue{}, err
	}
	return storage_engine.KeyValue{
		Key:     s.Key(),
		Value:   v,
		Version: int64(s.Version),
	}, nil
}

type StorePictureMetadata struct {
	Version   uint64 `json:"version"`
	TimeStamp uint64 `json:"timestamp"`
	Rid       string `json:"rid"`
	Aid       string `json:"aid"`
	Pid       string `json:"pic"`
	Thumbnail string `json:"thumbnail"`
	Url       string `json:"url"`
}

func (s *StorePictureMetadata) Key() string {
	return fmt.Sprintf("%s-%s-%s", s.Rid, s.Aid, s.Pid)
}

func (s *StorePictureMetadata) Encode() ([]byte, error) {
	return json.Marshal(s)
}

func (s *StorePictureMetadata) Decode(data []byte) error {
	return json.Unmarshal(data, s)
}

func (s *StorePictureMetadata) ToKvValue() (storage_engine.KeyValue, error) {
	v, err := s.Encode()
	if err != nil {
		return storage_engine.KeyValue{}, err
	}
	return storage_engine.KeyValue{
		Key:     s.Key(),
		Value:   v,
		Version: int64(s.Version),
	}, nil
}

type AdminTitlesResponse struct {
	Titles []string `json:"titles"`
}
