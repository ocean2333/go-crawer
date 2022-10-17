package parser

import (
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/ocean2333/go-crawer/server/config"
	"github.com/ocean2333/go-crawer/server/logger"
	"github.com/ocean2333/go-crawer/server/model"
	"github.com/ocean2333/go-crawer/server/net"
	"github.com/ocean2333/go-crawer/server/storage_engine"
	"github.com/stretchr/testify/assert"
)

func TestParseWnacgRules(t *testing.T) {
	LoadRules()
	r := GetRule("1")
	assert.Equal(t, "wnacg", r.Title)
	logger.Log.Infof("%#v\n", r)
	logger.Log.Infof("%#v\n", r.HomepageRules.Selectors)
	logger.Log.Infof("%#v\n", r.SearchRules)
	logger.Log.Infof("%#v\n", r.AlbumRules)
}

func TestParseHomepage(t *testing.T) {
	LoadRules()
	r := GetRule("1")
	resp, err := net.GetWithDefaultProxy(r.HomepageUrl)
	assert.Nil(t, err)
	doc, err := goquery.NewDocumentFromResponse(resp)
	assert.Nil(t, err)
	ParseHtml(r.Rid, "", r.HomepageRules, doc)
	kvValues, err := storage_engine.GetInstance().GetByPrefix(r.Rid)
	assert.Nil(t, err)
	for _, kvValue := range kvValues {
		storeAlbumMetadata := new(model.StoreAlbumMetadata)
		err := storeAlbumMetadata.Decode(kvValue.Value)
		assert.Nil(t, err)
		t.Logf("title: %s", storeAlbumMetadata.Title)
	}
}

func TestParseAlbumpage(t *testing.T) {
	LoadRules()
	r := GetRule("1")
	kvValues, err := storage_engine.GetInstance().GetByPrefix(r.Rid)
	assert.Nil(t, err)
	for _, kvValue := range kvValues {
		storeAlbumMetadata := new(model.StoreAlbumMetadata)
		err := storeAlbumMetadata.Decode(kvValue.Value)
		assert.Nil(t, err)
		resp, err := net.GetWithDefaultProxy("https://wnacg.net" + storeAlbumMetadata.Aid)
		assert.Nil(t, err)
		doc, err := goquery.NewDocumentFromResponse(resp)
		assert.Nil(t, err)
		ParseHtml(r.Rid, storeAlbumMetadata.Aid, r.AlbumRules, doc)
	}
}

func init() {
	go func() {
		storage_engine.Start(&config.Get().EtcdCfg.StorageEngineCfg)
		storage_engine.GetInstance().Campaign(&storage_engine.CampaignOptions{
			Cfg:               &config.Get().EtcdCfg.LeaderElectionCfg,
			CampaignKeyPrefix: "campaign",
			CampaignId:        "1",
		}, make(chan string))
	}()
}
