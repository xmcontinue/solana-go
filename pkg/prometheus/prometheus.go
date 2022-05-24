package prometheus

import (
	"encoding/json"
	"sync"

	"git.cplus.link/go/akit/config"
	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/rpcxio/libkv/store"

	"git.cplus.link/crema/backend/internal/etcd"
	"git.cplus.link/crema/backend/pkg/domain"
	"git.cplus.link/crema/backend/pkg/iface"
)

var (
	pushGatewayHost = "http://127.0.0.1:9091"
	configKey       = "/prometheus/auth"
	wg              sync.WaitGroup
	authConfig      map[string]AuthConfig
)

type AuthConfig struct {
	Projects []string
	Jobs     []string
}

func Init(conf *config.Config) error {
	host := conf.Get("prometheus.push_gateway_host")
	if host == nil {
		return errors.New("etcd host not found")
	}
	pushGatewayHost = host.(string)

	etcdSwapPairsKey := "/" + domain.GetPublicPrefix() + configKey
	// 加载swap pairs配置
	stopChan := make(chan struct{})
	resChan, err := etcd.Watch(etcdSwapPairsKey, stopChan)
	if err != nil {
		return errors.Wrap(err)
	}

	wg.Add(1)
	go WatchConfig(resChan)
	wg.Wait()
	return nil
}

func WatchConfig(resChan <-chan *store.KVPair) {
	var once sync.Once
	for {
		select {
		case res := <-resChan:
			err := json.Unmarshal(res.Value, &authConfig)
			if err != nil {
				logger.Error("prometheus config unmarshal failed :", logger.Errorv(err))
			}
			once.Do(func() {
				wg.Done()
			})
		}
	}
}

func ExamplePusherPush(req *iface.LogReq) error {
	completionTime := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: req.LogName,
		Help: req.LogHelp,
	})
	completionTime.SetToCurrentTime()
	completionTime.Set(req.LogValue)
	pusher := push.New(pushGatewayHost, req.JobName).Collector(completionTime)
	for k, v := range req.Tags {
		pusher = pusher.Grouping(k, v)
	}
	if err := pusher.Push(); err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func CheckAuth(req *iface.LogReq) error {
	has := func(key string, list []string) bool {
		for _, v := range list {
			if v == key {
				return true
			}
		}
		return false
	}

	if k, ok := authConfig[req.Key]; !ok {
		return errors.ParameterError
	} else {
		if has(req.Tags["project"], k.Projects) { // 暂时关闭对于job 的过滤，默认信任apikey的请求 has(req.JobName, k.Jobs) &&
			return nil
		}
	}

	return errors.ParameterError
}
