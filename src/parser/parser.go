package parser

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/ocean2333/go-crawer/src/logger"
	"github.com/ocean2333/go-crawer/src/model"
	"github.com/ocean2333/go-crawer/src/net"
	"github.com/ocean2333/go-crawer/src/storage_engine"
)

type findable interface {
	Find(selector string) *goquery.Selection
	Each(f func(int, *goquery.Selection)) *goquery.Selection
}

// parse html and save metadata in etcd, load pic into local path. For interactive experience, return all items' title
func ParseHtml(rid string, aid string, rules *Rules, html findable) (albumMetadatas []*model.StoreAlbumMetadata, pictureMetaDatas []*model.StorePictureMetadata) {
	items, err := parseItems(getSelectorByName("item", rules.Selectors), html)
	parsedId := make(map[string]bool)
	if err != nil {
		logger.Log.Errorf("parse item error: %v", err)
	}
	ops := make([]*storage_engine.KvOperation, 0)
	items.Each(func(i int, item *goquery.Selection) {
		storeItem := new(model.StoreItem)

		id, err := parseFromItem(getSelectorByName("id", rules.Selectors), item)
		if err != nil {
			logger.Log.Errorf("parse id error: %v", err)
		} else {
			storeItem.Id = id
		}

		// check if item is already stored
		var etcdId string
		if aid == "" {
			etcdId = fmt.Sprintf("%s-%s", rid, storeItem.Id)
			if data := peekEtcd(etcdId); data != nil {
				album := new(model.StoreAlbumMetadata)
				err := json.Unmarshal(data.Value, album)
				if err != nil {
					logger.Log.Errorf("unmarshal album error: %v", err)
				} else {
					albumMetadatas = append(albumMetadatas, album)
					return
				}
			}
		} else {
			etcdId = fmt.Sprintf("%s-%s-%s", rid, aid, storeItem.Id)
			if data := peekEtcd(etcdId); data != nil {
				pic := new(model.StorePictureMetadata)
				err := json.Unmarshal(data.Value, pic)
				if err != nil {
					logger.Log.Errorf("unmarshal pic error: %v", err)
				} else {
					pictureMetaDatas = append(pictureMetaDatas, pic)
					return
				}
			}
		}

		forAllSelectorExceptItem(rules.Selectors, func(selector *Selector) {
			if selector.Name == "id" {
				return
			}
			if selector.Name == "high_res" {
				str, err := parseFromItem(getSelectorByName("url", rules.Selectors), item)
				if err != nil {
					logger.Log.Errorf("parse url error: %v", err)
				} else {
					storeItem.Set("url", str)
				}
				resp, err := net.GetWithDefaultProxy(storeItem.Get("url"))
				if err != nil {
					logger.Log.Errorf("get %s error: %v", storeItem.Get("url"), err)
				} else {
					doc, err := goquery.NewDocumentFromReader(resp.Body)
					if err != nil {
						logger.Log.Errorf("new document error: %v", err)
					} else {
						highRes, err := parseFromItem(getSelectorByName("high_res", rules.Selectors), doc)
						if err != nil {
							logger.Log.Errorf("parse highRes error: %v", err)
						} else {
							storeItem.Set("high_res", highRes)
						}
					}
				}
				return
			}
			str, err := parseFromItem(selector, item)
			if err != nil {
				logger.Log.Errorf("parse %s error: %v", selector.Name, err)
			} else {
				storeItem.Set(selector.Name, str)
			}
		})

		if aid == "" {
			storeMetadata := &model.StoreAlbumMetadata{
				Version:     0,
				TimeStamp:   uint64(time.Now().UnixNano()),
				Rid:         rid,
				Aid:         storeItem.Id,
				Title:       storeItem.Get("title"),
				Datetime:    storeItem.Get("datetime"),
				Cover:       storeItem.Get("cover"),
				Author:      storeItem.Get("author"),
				Uploader:    storeItem.Get("uploader"),
				Rating:      storeItem.Get("rating"),
				Tag:         storeItem.Get("tag"),
				Description: storeItem.Get("description"),
				Url:         storeItem.Get("url"),
			}
			kv, err := storeMetadata.ToKvValue()
			if err != nil {
				logger.Log.Errorf("error while marshal (%#v) to kvValue: err %v", storeMetadata, err)
			}
			if parsedId[kv.Key] {
				return
			}
			ops = append(ops, &storage_engine.KvOperation{Op: storage_engine.KvOpcodePut, Kv: kv})
			parsedId[kv.Key] = true
			albumMetadatas = append(albumMetadatas, storeMetadata)
		} else {
			if storeItem.Get("high_res") != "" {
				storeItem.Set("url", storeItem.Get("high_res"))
			}
			storeMetadata := &model.StorePictureMetadata{
				Version:   0,
				TimeStamp: uint64(time.Now().UnixNano()),
				Rid:       rid,
				Aid:       aid,
				Pid:       storeItem.Id,
				Thumbnail: storeItem.Get("thumbnail"),
				Url:       storeItem.Get("url"),
			}
			kv, err := storeMetadata.ToKvValue()
			if err != nil {
				logger.Log.Errorf("error while marshal (%#v) to kvValue: err %v", storeMetadata, err)
			}
			if parsedId[kv.Key] {
				return
			}
			ops = append(ops, &storage_engine.KvOperation{Op: storage_engine.KvOpcodePut, Kv: kv})
			parsedId[kv.Key] = true
			pictureMetaDatas = append(pictureMetaDatas, storeMetadata)
		}
	})

	if _, err := storage_engine.GetInstance().Submit(ops); err != nil {
		logger.Log.Errorf("submit failed: err %v", err)
	}
	return
}

// parse certain page
func ParseAlbumPage(rid string, aid string, page int) ([]*model.StoreAlbumMetadata, []*model.StorePictureMetadata, error) {
	rule := GetRule(rid)
	patten := map[string]string{
		"pid":  rid,
		"aid":  aid,
		"page": strconv.Itoa(page),
	}
	url := RepalcePatten(patten, rule.AlbumUrl)
	resp, err := net.GetWithDefaultProxy(url)
	if err != nil {
		logger.Log.Errorf("get %s error: %v", url, err)
		return nil, nil, errors.New("failed to get url")
	}
	if resp.StatusCode != 200 {
		logger.Log.Errorf("get %s error: status code %d", url, resp.StatusCode)
		return nil, nil, fmt.Errorf("failed to get url, status code %d", resp.StatusCode)
	}
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		logger.Log.Errorf("new document err: %v", err)
		return nil, nil, fmt.Errorf("failed to new document")
	}
	_, picMetaData := ParseHtml(rid, aid, rule.AlbumRules, doc)
	for _, pic := range picMetaData {
		go net.GetWorkerPoolManager().GetWorkerPool(rid).GetPicWithProxyAndSave(pic.Url, rid, pic.Aid, pic.Pid)
	}
	return nil, picMetaData, nil
}

// parse items from html
func parseItems(selector *Selector, html findable) (findable, error) {
	if selector == nil {
		return nil, errors.New("selector is nil")
	}
	for _, selectPath := range selector.Selector {
		html = html.Find(selectPath)
	}
	return html, nil
}

// parse element inside item
func parseFromItem(selector *Selector, item findable) (string, error) {
	var (
		str   string
		exist bool
		s     *goquery.Selection
	)
	if len(selector.Selector) < 1 {
		s = item.(*goquery.Selection)
	} else {
		s = item.Find(selector.Selector[0])
	}

	if selector.Func == "attr" {
		str, exist = s.Attr(selector.Param)
		if !exist {
			return "", errors.New("attr not exist")
		}
	} else if selector.Func == "text" {
		str = s.Text()
	} else {
		return "", errors.New("unknown func")
	}

	if selector.Regexp != "" {
		re, err := regexp.Compile(selector.Regexp)
		if err != nil {
			logger.Log.Errorf("regexp(%s) invalid: %v", selector.Regexp, err)
		} else {
			oldStr := str
			str = re.FindString(str)
			logger.Log.Infof("regexp(%s) actived: %s->%s", selector.Regexp, oldStr, str)
		}
	}

	if selector.ReplacesPatten != "" {
		str = strings.Replace(str, selector.ReplacesPatten, selector.ReplacesString, -1)
		logger.Log.Infof("replace(%s) actived: %s->%s", selector.ReplacesPatten, str, selector.ReplacesString)
	}

	if selector.Prefix != "" {
		str = selector.Prefix + str
		logger.Log.Debug("add prefix: ", str)
	}

	return str, nil
}

func peekEtcd(id string) *storage_engine.KeyValue {
	data, err := storage_engine.GetInstance().Get(id)
	if err != nil {
		logger.Log.Errorf("get etcd err: %v", err)
		return nil
	} else if data == nil {
		return nil
	} else {
		logger.Log.Infof("get stored data: id %s", data.Key)
		return data
	}
}
