package market

import (
	"context"
	"encoding/json"

	"git.cplus.link/go/akit/errors"

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
	configMap := make(map[string][]byte)

	listVal, err := etcd.Client().GetKeyValue(context.TODO(), getEtcdConfigKey(confListKey))
	if err != nil || listVal == nil {
		return errors.Wrap(err)
	}

	var confList []string

	err = json.Unmarshal(listVal.Value, &confList)
	if err != nil {
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
