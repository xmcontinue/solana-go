package process

import (
	"sync"

	redisV8 "git.cplus.link/go/akit/client/redis/v8"
	"git.cplus.link/go/akit/config"
	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"
	"git.cplus.link/go/akit/pkg/worker/xcron"
	"git.cplus.link/go/akit/pkg/xlog"
	akHttp "git.cplus.link/go/akit/transport/http"
	ag_solanago "github.com/gagliardetto/solana-go"
	"github.com/go-redis/redis/v8"
	"github.com/robfig/cron/v3"

	event "git.cplus.link/crema/backend/chain/event/parser"
)

var (
	redisClient     *redisV8.Client
	httpClient      *akHttp.Client
	conf            *config.Config
	job             *Job
	delAndAddByZSet *redis.Script
	collectionMint  string
	lock            sync.Mutex
)

const (
	defaultBaseSpec = "0 * * * * *"
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
	conf = viperConf
	var err error

	httpClient = akHttp.DefaultClient()

	event.Init()

	job = NewJob()

	err = conf.UnmarshalKey("cron_job_conf", &job.CronConf)
	if err != nil {
		return errors.Wrap(err)
	}
	//logger.Info("migrate begin")
	////先做数据迁移，做完后才能获取新数据
	//err = migrate()
	//if err != nil {
	//	return errors.Wrap(err)
	//}
	//
	//logger.Info("migrate done")
	//logger.Info("migrate done")

	err = conf.UnmarshalKey("collection_mint", &collectionMint)
	if err != nil {
		return errors.Wrap(err)
	}

	job.CronConf.WithLogger(xlog.Config{}.Build())

	job.Cron = job.CronConf.Build()
	redisClient, err = initRedis(conf)
	if err != nil {
		return errors.Wrap(err)
	}
	delAndAddByZSet = redis.NewScript(DelAndAddSZSetScript)

	// xCron init
	err = conf.UnmarshalKey("cron_job_conf", &job.CronConf)
	if err != nil {
		return errors.Wrap(err)
	}

	_, err = job.Cron.AddFunc(getSpec("sum_total_swap_account"), sumTotalSwapAccount)
	if err != nil {
		panic(err)
	}

	_, err = job.Cron.AddFunc(getSpec("sync_total_swap_cache"), SwapTotalCount)
	if err != nil {
		panic(err)
	}

	_, err = job.Cron.AddFunc(getSpec("sync_gallery"), SyncGalleryJob)
	if err != nil {
		panic(err)
	}

	// 统计相同币种的tvl
	//_, err = job.Cron.AddFunc(getSpec("tvl_of_token"), tvlOfToken)
	if err != nil {
		panic(err)
	}

	//清理数据
	//_, err = job.Cron.AddFunc(getSpec("clear_old_data"), clearSwapCountKline)
	//if err != nil {
	//	panic(err)
	//}

	//_, err = job.Cron.AddFunc(getSpec("tvl_of_token"), parserTransactionUserCount)
	if err != nil {
		panic(err)
	}

	// create sync transaction cron job
	syncTransactionJob := NewJobInfo("SyncSwapCountKline")
	job.JobList["SyncSwapCountKline"] = syncTransactionJob
	_, err = job.Cron.AddFunc(defaultBaseSpec, CreateSyncKLine)

	userSyncTransactionJob := NewJobInfo("SyncUserCountKLine")
	job.JobList["SyncUserCountKLine"] = userSyncTransactionJob
	//_, err = job.Cron.AddFunc(defaultBaseSpec, CreateUserSyncKLine)
	if err != nil {
		panic(err)
	}

	job.Cron.Start()

	subMetadata := &SubMetadata{
		account:        ag_solanago.TokenMetadataProgramID.String(),
		collectionMint: collectionMint,
	}
	go subMetadata.Sub()

	return nil
}

func getSpec(key string) string {
	return conf.Get("cron_job_interval." + key).(string)
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

const DelAndAddSZSetScript string = "if redis.call('zcard', KEYS[1]) > 0 then\n" +
	"   redis.call('del', KEYS[1])\n" +
	"   for i, v in pairs(ARGV) do\n" +
	"       if i > (table.getn(ARGV)) / 2 then\n" +
	"           break\n" +
	"       end\n" +
	"       redis.call('zadd', KEYS[1], ARGV[2*i - 1], ARGV[2*i])\n" +
	"   end\n" +
	"   return 1\n" +
	"else\n" +
	"   for i, v in pairs(ARGV) do\n" +
	"       if i > (table.getn(ARGV)) / 2 then\n" +
	"           break\n" +
	"       end\n" +
	"       redis.call('zadd',KEYS[1], ARGV[2*i - 1], ARGV[2*i])\n" +
	"   end\n" +
	"   return 1\n" +
	"end"
