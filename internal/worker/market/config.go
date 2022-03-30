package market

import (
	"context"
	"encoding/json"

	"git.cplus.link/go/akit/errors"
	"git.cplus.link/go/akit/logger"

	"git.cplus.link/crema/backend/internal/etcd"
	"git.cplus.link/crema/backend/pkg/domain"
)

const (
	baseConfigKey = "/crema/"
	confListKey   = "conf_list"
)

var (
	configCache      map[string][]byte
	tokenConfigCache []byte
)

func SyncConfigJob() error {
	logger.Info("config syncing ......")
	configMap := make(map[string][]byte)

	listVal, err := etcd.Api().Get(context.TODO(), getEtcdConfigKey(confListKey), nil)
	if err != nil || listVal == nil {
		logger.Error("config sync fail:", logger.Errorv(err))
		return errors.Wrap(err)
	}
	if listVal == nil {
		logger.Error("config sync fail: conf_list is nil")
		return errors.Wrap(errors.RecordNotFound)
	}

	var confList []string

	err = json.Unmarshal([]byte(listVal.Node.Value), &confList)
	if err != nil {
		logger.Error("config sync fail:", logger.Errorv(err))
		return errors.Wrap(err)
	}

	for _, v := range confList {

		confVal, err := etcd.Api().Get(context.TODO(), getEtcdConfigKey(v), nil)
		if err != nil || confVal == nil {
			continue
		}

		if v == "swap-pairs" {
			setTokenConfig([]byte(confVal.Node.Value))
		}

		configMap[v] = []byte(confVal.Node.Value)
	}

	configCache = configMap

	logger.Info("config sync complete!")
	return nil
}

func setTokenConfig(b []byte) {
	var swapConfig []*domain.SwapConfig
	err := json.Unmarshal(b, &swapConfig)
	if err != nil {
		return
	}

	tokenConfig := make(map[string]*domain.TokenConfig, 0)
	for _, v := range swapConfig {
		tokenConfig[v.TokenA.Symbol] = &domain.TokenConfig{
			Name:   v.TokenA.Name,
			Symbol: v.TokenA.Symbol,
		}

		tokenConfig[v.TokenB.Symbol] = &domain.TokenConfig{
			Name:   v.TokenB.Name,
			Symbol: v.TokenB.Symbol,
		}
	}

	newTokenConfigCache, err := json.Marshal(tokenConfig)
	if err != nil {
		return
	}
	tokenConfigCache = newTokenConfigCache
}

func GetTokenConfig() []byte {
	return tokenConfigCache
}

// GetConfig ...
func GetConfig(configName string) []byte {
	return configCache[configName]
}

// getEtcdConfigKey ...
func getEtcdConfigKey(key string) string {
	return baseConfigKey + key
}
