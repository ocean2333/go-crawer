package download

import (
	"sync"
	"sync/atomic"

	"github.com/ocean233/go-crawer/src/logger"
)

var (
	handler          *DownloadHandler
	workerPool       sync.Pool
	workerSleepTime  uint32
	maxWorkerNum     uint32
	workingWorkerNum uint32
	jobChan          chan string
	resChan          chan string
	doingJobs        map[string]*downloadWorker
)

type DownloadHandler struct{}

type Option struct {
	WorkerNum       uint8
	WorkerSleepTime uint32
	jobQueueBuffer  uint32
	resQueueBuffer  uint32
}

func init() {
	InitDownloadHandler(DefaultOption())
}

func DefaultOption() *Option {
	return &Option{
		WorkerNum:       1,
		WorkerSleepTime: 5,
		jobQueueBuffer:  5,
		resQueueBuffer:  5,
	}
}

func InitDownloadHandler(opt *Option) {
	workerPool = sync.Pool{}
	workerSleepTime = opt.WorkerSleepTime
	maxWorkerNum = uint32(opt.WorkerNum)
	workingWorkerNum = 0
	jobChan = make(chan string, opt.jobQueueBuffer)
	resChan = make(chan string, opt.resQueueBuffer)
	doingJobs = make(map[string]*downloadWorker)
	workerPool.New = func() any {
		return &downloadWorker{
			jobChan:   jobChan,
			sleepTime: uint64(workerSleepTime),
		}
	}
}

func GetHandler() *DownloadHandler {
	if handler == nil {
		handler = &DownloadHandler{}
	}
	return handler
}

func (h *DownloadHandler) newWorker() *downloadWorker {
	nowDownloadWorkerNum := workingWorkerNum
	if nowDownloadWorkerNum < maxWorkerNum {
		if ok := atomic.CompareAndSwapUint32(&workingWorkerNum, nowDownloadWorkerNum, nowDownloadWorkerNum+1); ok {
			return workerPool.Get().(*downloadWorker)
		}
	}
	return nil
}

func (h *DownloadHandler) AddJob(url string) {
	jobChan <- url
	logger.Log.Infof("job url:%s added", url)
}

func (h *DownloadHandler) registJob(jobUrl string, downloadWorker *downloadWorker) {
	doingJobs[jobUrl] = downloadWorker
}

func (h *DownloadHandler) doneJob(jobUrl string) {
	delete(doingJobs, jobUrl)
}

func (h *DownloadHandler) DoingJobs() []string {
	jobs := []string{}
	for job, _ := range doingJobs {
		jobs = append(jobs, job)
	}
	return jobs
}

func (h *DownloadHandler) Start() {
	h.start()
}

func (h *DownloadHandler) start() {
	for i := 0; i < int(maxWorkerNum); i++ {
		go func() {
			w := h.newWorker()
			if w == nil {
				logger.Log.Warn("created a nil worker")
				return
			}
			w.jobChan = jobChan
			w.resChan = resChan
			w.sleepTime = uint64(workerSleepTime)
			logger.Log.Infof("a worker is up")
			w.work()
		}()
	}

}
