package market

import (
	"context"
	"encoding/json"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"

	"git.cplus.link/crema/backend/internal/etcd"
)

const (
	baseConfigKey = "/crema/"
	confListKey   = "conf_list"
)

var (
	configCache map[string][]byte
)

func SyncConfigJob() error {
	logger.Info("config syncing ......")
	configMap := make(map[string][]byte)

	listVal, err := etcd.Client().GetKeyValue(context.TODO(), getEtcdConfigKey(confListKey))
	if err != nil || listVal == nil {
		logger.Error("config sync fail:", logger.Errorv(err))
		return errors.Wrap(err)
	}
	if listVal == nil {
		logger.Error("config sync fail: conf_list is nil")
		return errors.Wrap(errors.RecordNotFound)
	}

	var confList []string

	err = json.Unmarshal(listVal.Value, &confList)
	if err != nil {
		logger.Error("config sync fail:", logger.Errorv(err))
		return errors.Wrap(err)
	}

	for _, v := range confList {

		confVal, err := etcd.Client().GetKeyValue(context.TODO(), getEtcdConfigKey(v))
		if err != nil || confVal == nil {
			continue
		}

		configMap[v] = confVal.Value
	}

	configCache = configMap

	logger.Info("config sync complete!")
	return nil
}

// GetConfig ...
func GetConfig(configName string) []byte {
	return configCache[configName]
}

// getEtcdConfigKey ...
func getEtcdConfigKey(key string) string {
	return baseConfigKey + key
}
