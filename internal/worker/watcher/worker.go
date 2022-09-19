package watcher

import (
	"sync"

	redisV8 "git.cplus.link/go/akit/client/redis/v8"
	"git.cplus.link/go/akit/config"
	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"
	"git.cplus.link/go/akit/pkg/worker/xcron"
	"git.cplus.link/go/akit/pkg/xlog"
	"github.com/robfig/cron/v3"

	event "git.cplus.link/crema/backend/chain/event/parser"
)

const defaultBaseSpec = "0 * * * * *"

var (
	job         *Job
	conf        *config.Config
	redisClient *redisV8.Client
)

type Job struct {
	CronConf *xcron.Config
	Cron     *xcron.Cron
	JobList  map[string]*JobInfo
}

type JobInfo struct {
	Name    string
	Structs *sync.Map
	Watch   Watch
}

type Watch struct {
	OldNap *sync.Map
	NewNap *sync.Map
}

type JobStruct struct {
	JStruct JobInterface
	EntryID cron.EntryID
	Spec    string
}

type JobInterface interface {
	GetSpecFunc() string
	DeleteJobFunc(*JobInfo) error
	Run() error
	Name() string
}

// Init 定时任务
func Init(viperConf *config.Config) error {
	job = NewJob()
	conf = viperConf

	event.Init()

	err := conf.UnmarshalKey("cron_job_conf", &job.CronConf)

	redisClient, err = initRedis(conf)
	if err != nil {
		return errors.Wrap(err)
	}

	if err != nil {
		return errors.Wrap(err)
	}
	job.CronConf.WithLogger(xlog.Config{}.Build())

	job.Cron = job.CronConf.Build()

	// create sync tvl cron job
	syncTvlJob := NewJobInfo("SyncTvl")
	job.JobList["SyncTvl"] = syncTvlJob
	// _, err = job.Cron.AddFunc(defaultBaseSpec, CreateSyncTvl)

	// create sync transaction cron job
	syncTransactionJob := NewJobInfo("SyncTvl")
	job.JobList["SyncTransaction"] = syncTransactionJob
	_, err = job.Cron.AddFunc(defaultBaseSpec, CreateSyncTransaction)

	// 同步vol(24h)
	_, err = job.Cron.AddFunc(getSpec("sum_tvl"), SyncVol24H)

	// 同步总vol
	_, err = job.Cron.AddFunc(getSpec("sum_tvl"), SyncTotalVol)

	// 同步价格至kline
	_, err = job.Cron.AddFunc(getSpec("sync_kline"), SyncSwapPrice)

	// TODO 由于未测试完成其他功能上线，此处暂时关闭
	// _, err = job.Cron.AddFunc(getSpec("activity_history"), SyncActivityTransaction)

	// 解析已经同步的数据，这些数据在第一次同步时没有解析类型和user_address
	_, err = job.Cron.AddFunc(getSpec("sync_kline"), SyncTypeAndUserAddressHistory)

	job.Cron.Start()

	return nil
}

func NewJob() *Job {
	return &Job{
		JobList: map[string]*JobInfo{},
	}
}

func NewJobInfo(name string) *JobInfo {
	return &JobInfo{
		Name:    name,
		Structs: &sync.Map{},
		Watch: Watch{
			OldNap: &sync.Map{},
			NewNap: &sync.Map{},
		},
	}
}

// initRedis 初始化redis
func initRedis(conf *config.Config) (*redisV8.Client, error) {
	c := redisV8.DefaultRedisConfig()
	err := conf.UnmarshalKey("redis", c)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	return redisV8.NewClient(c)
}

func (j *Job) WatchJobForMap(name string, newMap *sync.Map, createFunc func(interface{}) JobInterface) error {
	info, ok := j.JobList[name]
	if !ok {
		return errors.RecordNotFound
	}
	info.Watch.OldNap = info.Watch.NewNap

	// delete old job
	info.Watch.NewNap.Range(func(key, value interface{}) bool {
		newValue, ok := newMap.Load(key)
		if value == newValue {
			return true
		}

		info.Watch.NewNap.Delete(key)

		str, ok := info.Structs.Load(key)
		if !ok {
			return true
		}

		err := str.(*JobStruct).JStruct.DeleteJobFunc(info)
		if err != nil {
			logger.Error("delete job failed :", logger.Errorv(err))
		}

		j.Cron.Remove(str.(*JobStruct).EntryID)

		info.Structs.Delete(key)

		return true
	})

	// create job
	newMap.Range(func(key, value interface{}) bool {
		if _, ok := info.Watch.NewNap.Load(key); !ok {
			jStruct := createFunc(value)
			spec := jStruct.GetSpecFunc()

			entryID, err := j.Cron.AddJob(spec, jStruct)
			if err != nil {
				return false
			}

			info.Structs.Store(key, &JobStruct{
				JStruct: jStruct,
				EntryID: entryID,
				Spec:    spec,
			})

			info.Watch.NewNap.Store(key, value)
		}
		return true
	})

	return nil
}

func getSpec(key string) string {
	return conf.Get("cron_job_interval." + key).(string)
}
