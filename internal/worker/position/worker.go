package position

import (
	"sync"

	"git.cplus.link/go/akit/config"
	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"
	"git.cplus.link/go/akit/pkg/worker/xcron"
	"git.cplus.link/go/akit/pkg/xlog"
	"github.com/robfig/cron/v3"
)

var (
	conf *config.Config
	job  *Job
)

const defaultBaseSpec = "0 * * * * *"

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
	conf = viperConf
	var err error

	job = NewJob()

	err = conf.UnmarshalKey("cron_job_conf", &job.CronConf)
	if err != nil {
		return errors.Wrap(err)
	}

	job.CronConf.WithLogger(xlog.Config{}.Build())

	job.Cron = job.CronConf.Build()

	// xCron init
	err = conf.UnmarshalKey("cron_job_conf", &job.CronConf)
	if err != nil {
		return errors.Wrap(err)
	}

	_, err = job.Cron.AddFunc(getSpec("sync_position"), syncPosition)
	if err != nil {
		panic(err)
	}

	job.Cron.Start()

	return nil
}

func getSpec(key string) string {
	return conf.Get("cron_job_interval." + key).(string)
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
