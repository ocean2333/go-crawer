package album

import "github.com/ocean233/go-crawer/src/crawer/logger"

var (
	maxSearchAlbum       int
	maxDisplayNumPerPage int
	albums               []*Album
	manager              = &AlbumManager{}
)

type AlbumManager struct {
}

type Option struct {
	MaxSearchAlbum       int
	MaxDisplayNumPerPage int
}

func init() {
	InitAlbumManager(DefaultOption())
}

func DefaultOption() *Option {
	return &Option{
		MaxSearchAlbum:       1000,
		MaxDisplayNumPerPage: 20,
	}
}

func InitAlbumManager(opt *Option) {
	maxSearchAlbum = opt.MaxSearchAlbum
	maxDisplayNumPerPage = opt.MaxDisplayNumPerPage
}

func GetAlbumManager() *AlbumManager {
	if manager == nil {
		manager = &AlbumManager{}
	}
	return manager
}

func GetAPageAlbums(pageIdx int) []*Album {
	if len(albums) < pageIdx*maxDisplayNumPerPage {
		logger.Log.Infof("get page %d albums(%d:%d), now have %d, shows %d", pageIdx, pageIdx*maxDisplayNumPerPage, (pageIdx+1)*maxDisplayNumPerPage, len(albums), 0)
		return nil
	} else if len(albums) < (pageIdx+1)*maxDisplayNumPerPage {
		logger.Log.Infof("get page %d albums(%d:%d), now have %d, shows %d", pageIdx, pageIdx*maxDisplayNumPerPage, (pageIdx+1)*maxDisplayNumPerPage, len(albums), len(albums)-pageIdx*maxDisplayNumPerPage)
		return albums[pageIdx*maxDisplayNumPerPage:]
	}
	logger.Log.Infof("get page %d albums(%d:%d), now have %d, shows %d", pageIdx, pageIdx*maxDisplayNumPerPage, (pageIdx+1)*maxDisplayNumPerPage, len(albums), maxDisplayNumPerPage)
	return albums[pageIdx*maxDisplayNumPerPage : (pageIdx+1)*maxDisplayNumPerPage]
}

func Add(album *Album) {
	logger.Log.Infof("add album to manager: %s", album.Name)
	albums = append(albums, album)
}

func Delete(url string) {
	for i, album := range albums {
		if album.URL == url {
			tmp := make([]*Album, 0)
			tmp = append(tmp, albums[:i]...)
			albums = append(tmp, albums[i+1:]...)
			logger.Log.Infof("delete album from manager: %s", album.Name)
			return
		}
	}
}

func Clean() {
	logger.Log.Info("clean album")
	albums = []*Album{}
}
