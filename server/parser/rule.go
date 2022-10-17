package parser

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"github.com/ocean2333/go-crawer/src/logger"
	"github.com/ocean2333/go-crawer/src/net"
)

const rulePath = "./rules"

type ParseRules struct {
	Rid           string       `json:"rid"`
	Title         string       `json:"title"`
	HomepageUrl   string       `json:"homepage_url"`
	SearchUrl     string       `json:"search_url"`
	AlbumUrl      string       `json:"album_url"`
	LoginUrl      string       `json:"login_url"`
	HomepageRules *Rules       `json:"homepage_rules"`
	AlbumRules    *Rules       `json:"album_rules"`
	SearchRules   *Rules       `json:"search_rules"`
	Categories    []*Category  `json:"categories"`
	GlobalRules   *GlobalRules `json:"global_rules"`
}

func (pr *ParseRules) String() string {
	return fmt.Sprintf("rid:%s,\n title:%s,\n homepage_url:%s,\n search_url:%s,\n album_url:%s,\n login_url:%s,\n homepage_rules:%v,\n album_rules:%v,\n search_rules:%v,\n categories:%v,\n global_rules:%v\n", pr.Rid, pr.Title, pr.HomepageUrl, pr.SearchUrl, pr.AlbumUrl, pr.LoginUrl, pr.HomepageRules.Selectors, pr.AlbumRules.Selectors, pr.SearchRules.Selectors, pr.Categories, pr.GlobalRules)
}

type Rules struct {
	Selectors []*Selector `json:"selectors"`
}

func (r *Rules) String() string {
	str := ""
	for _, selector := range r.Selectors {
		str += fmt.Sprintf("selectors:%v", selector.String())
	}
	return str
}

type Category struct {
	Cid   int    `json:"cid"`
	Title string `json:"title"`
	Url   string `json:"url"`
}

type Selector struct {
	Name           string   `json:"name"`
	Selector       []string `json:"selector"`
	Func           string   `json:"fun"`
	Param          string   `json:"param"`
	Regexp         string   `json:"regexp"`
	ReplacesPatten string   `json:"replace_patten"`
	ReplacesString string   `json:"replaces_string"`
	Prefix         string   `json:"prefix"`
}

func (s *Selector) String() string {
	return fmt.Sprintf("name:%s, selector:%v, func:%s, param:%s, regexp:%s, replace_patten:%s, replaces_string:%s, prefix:%s", s.Name, s.Selector, s.Func, s.Param, s.Regexp, s.ReplacesPatten, s.ReplacesString, s.Prefix)
}

type GlobalRules struct {
	PreloadHome    int `json:"preload_home"`
	PreloadAlbum   int `json:"preload_album"`
	PreloadPic     int `json:"preload_pic"`
	MetaDataMaxAge int `json:"metadata_max_age"`
	MaxConnectNum  int `json:"max_connect_num"`
}

var (
	rulesMap map[string]*ParseRules
)

func LoadRules() {
	rulesMap = make(map[string]*ParseRules)
	fileInfos, err := ioutil.ReadDir(rulePath)
	if err != nil {
		panic(err)
	}
	for _, fileInfo := range fileInfos {
		data, err := ioutil.ReadFile(path.Join(rulePath, fileInfo.Name()))
		if err != nil {
			logger.Log.Errorf("failed to load file %s: %v", fileInfo.Name(), err)
		} else {
			parseRule := new(ParseRules)
			err = json.Unmarshal(data, parseRule)
			if err != nil {
				logger.Log.Errorf("failed to unmarshal data from file %s: %v", fileInfo.Name(), err)
			}
			rulesMap[parseRule.Rid] = parseRule
			logger.Log.Infof("load rule %s success", parseRule.String())
			net.GetWorkerPoolManager().RegisterPool(parseRule.Rid, parseRule.GlobalRules.MaxConnectNum)
		}
	}
}

func GetRule(rid string) *ParseRules {
	return rulesMap[rid]
}

func init() {
	LoadRules()
}

func RepalcePatten(patten map[string]string, str string) string {
	for k, v := range patten {
		str = strings.Replace(str, fmt.Sprintf("{patten-%s}", k), v, -1)
	}
	return str
}

func getSelectorByName(name string, selectors []*Selector) *Selector {
	for _, selector := range selectors {
		if selector.Name == name {
			return selector
		}
	}
	return nil
}

func forAllSelectorExceptItem(selectors []*Selector, f func(selector *Selector)) {
	for _, selector := range selectors {
		if selector.Name != "item" {
			f(selector)
		}
	}
}
